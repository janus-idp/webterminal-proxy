package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	utils "github.com/janus-idp/webterminal-proxy/utils"
)

const (
	CONTAINER                       = "web-terminal-tooling"
	HANDSHAKE_SUBPROTOCOL           = "terminal.k8s.io"
	SERVER_ADDRESS_SUBPROTOCOL      = "base64url.console.link.k8s.io."
	AUTHORIZATION_TOKEN_SUBPROTOCOL = "base64url.bearer.authorization.k8s.io."
	WORKSPACE_ID_SUBPROTOCOL        = "base64url.workspace.id.k8s.io."
	TERMINAL_ID_SUBPROTOCOL         = "base64url.terminal.id.k8s.io."
	TERMINAL_SIZE_SUBPROTOCOL       = "base64url.terminal.size.k8s.io."
	NAMESPACE                       = "base64url.namespace.k8s.io."
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	Subprotocols:    []string{HANDSHAKE_SUBPROTOCOL},
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func setupCommandString(connectionData utils.ConnectionData) string {
	commands := []string{"/bin/sh", "-i", "-c", fmt.Sprintf("stty cols %s rows %s; TERM=xterm bash", connectionData.TerminalSize[0], connectionData.TerminalSize[1])}
	finalCommand := ""
	for _, command := range commands {
		finalCommand += "&command=" + url.QueryEscape(command)
	}
	finalCommand += "&stdout=1&stdin=1&stderr=1&tty=1"
	return finalCommand
}

func setupPod(connectionData utils.ConnectionData) (string, error) {
	username, err := utils.GetUserName(connectionData)
	if err != nil {
		return "", err
	}
	config := &utils.Config{
		Container: CONTAINER,
		Kubeconfig: utils.KubeConfig{
			Username:  username,
			Namespace: connectionData.Namespace,
		},
	}
	podID, err := utils.SetupUserPod(connectionData, config)
	if err != nil {
		return "", err
	}
	return podID, nil
}

func parseSubprotocols(r *http.Request) (utils.ConnectionData, error) {
	var connectionData utils.ConnectionData
	var err error

	for _, subprotocol := range strings.Split(r.Header.Get("Sec-WebSocket-Protocol"), ", ") {
		subprotocol = strings.TrimSpace(subprotocol)

		myMap := map[string]*string{
			SERVER_ADDRESS_SUBPROTOCOL:      &connectionData.Link,
			AUTHORIZATION_TOKEN_SUBPROTOCOL: &connectionData.Token,
			WORKSPACE_ID_SUBPROTOCOL:        &connectionData.WorkspaceID,
			TERMINAL_ID_SUBPROTOCOL:         &connectionData.TerminalID,
			NAMESPACE:                       &connectionData.Namespace,
		}
		if strings.HasPrefix(subprotocol, TERMINAL_SIZE_SUBPROTOCOL) {
			terminalSize, err := url.QueryUnescape(strings.TrimPrefix(subprotocol, TERMINAL_SIZE_SUBPROTOCOL))
			if err != nil {
				return connectionData, err
			}
			connectionData.TerminalSize = strings.Split(terminalSize, "x")
		}

		for prefix, field := range myMap {
			if strings.HasPrefix(subprotocol, prefix) {
				*field, err = url.QueryUnescape(strings.TrimPrefix(subprotocol, prefix))
				if err != nil {
					return connectionData, err
				}
				break
			}
		}

	}
	return connectionData, nil

}

func connectWebsocketServer(connectionData utils.ConnectionData) *websocket.Conn {
	dialer := websocket.DefaultDialer
	dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	command := setupCommandString(connectionData)
	c, response, err := dialer.Dial(fmt.Sprintf("wss://%s/api/v1/namespaces/%s/pods/%s/exec?&container=%s%s", connectionData.Link, connectionData.Namespace, connectionData.PodID, CONTAINER, command), http.Header{"Authorization": []string{"Bearer " + connectionData.Token}})
	if err != nil {
		log.Fatal("Unable to connect a pod: ", err, " With response: ", response)
		return nil
	}
	return c
}

func keepAlive(podConnection *websocket.Conn) {
	buffer := bytes.NewBuffer([]byte{0})
	buffer.WriteString("Keep alive")
	for {
		err := podConnection.WriteMessage(websocket.PingMessage, buffer.Bytes())
		if err != nil {
			log.Println("Unable to send keep alive message to pod: ", err)
			return
		}
		time.Sleep(30 * time.Second)
	}
}

func clientInput(clientConnection *websocket.Conn, podConnection *websocket.Conn, connectionData utils.ConnectionData) {
	ticker := time.NewTicker(5 * time.Minute)
	for {
		messageType, message, err := clientConnection.ReadMessage()
		if err != nil {
			log.Println("Client read error: ", err)
			return
		}
		buffer := bytes.NewBuffer([]byte{0})
		buffer.WriteString(string(message))
		err = podConnection.WriteMessage(messageType, buffer.Bytes())
		if err != nil {
			log.Println("Pod write error: ", err)
			return
		}
		select {
		case <-ticker.C:
			err = utils.SendActivityTick(connectionData)
			if err != nil {
				log.Println("Unable to send activity tick: ", err)
				return
			}
		default:
		}
	}
}

func terminalOutput(clientConnection *websocket.Conn, podConnection *websocket.Conn, connectionData utils.ConnectionData) {
	ticker := time.NewTicker(1 * time.Minute)
	for {
		messageType, message, err := podConnection.ReadMessage()
		if err != nil {
			log.Println("Pod read: ", err, messageType, message)
			return
		}
		err = clientConnection.WriteMessage(messageType, message)
		if err != nil {
			log.Println("Client write: ", err)
			return
		}
		select {
		case <-ticker.C:
			err = utils.SendActivityTick(connectionData)
			if err != nil {
				log.Println("Unable to send activity tick: ", err)
				return
			}
		default:
		}
	}
}
func handleWebsocket(w http.ResponseWriter, r *http.Request) {
	h := http.Header{}
	h.Set("Sec-WebSocket-Protocol", HANDSHAKE_SUBPROTOCOL)
	connectionData, err := parseSubprotocols(r)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer utils.CleanAfterDisconnect(connectionData)
	connectionData.PodID, err = setupPod(connectionData)
	if err != nil {
		log.Fatal(err)
		return
	}
	clientConnection, err := upgrader.Upgrade(w, r, h)
	if err != nil {
		log.Fatal(err)
		return
	}
	if err != nil {
		log.Fatal(err)
		return
	}
	podConnection := connectWebsocketServer(connectionData)
	go terminalOutput(clientConnection, podConnection, connectionData)
	go keepAlive(podConnection)
	clientInput(clientConnection, podConnection, connectionData)
	log.Println("Connection closed")
}

func setupRoute() {
	http.HandleFunc("/", handleWebsocket)
}

func main() {
	setupRoute()
	log.Fatal(http.ListenAndServe(":8080", nil))
}
