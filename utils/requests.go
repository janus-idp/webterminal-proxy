package utils

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

var restClient = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	},
}

const (
	NAMESPACE                 = "openshift-terminal"
	USER_ENDPOINT             = "apis/user.openshift.io/v1/users/~"
	SERVICE_ENDPOINT          = "api/v1/namespaces/" + NAMESPACE + "/services"
	SERVICE_EXEC_ENDPOINT     = "service:4444/proxy/exec/init"
	SERVICE_ACTIVITY_ENDPOINT = "service:4444/proxy/activity/tick"
)

var DEVWORKSPACE_ENDPOINT = "apis/workspace.devfile.io/v1alpha2/namespaces/" + NAMESPACE + "/devworkspaces"

func SetupUserPod(connectionData ConnectionData, config *Config) (string, error) {
	payload, _ := json.Marshal(config)
	request, err := http.NewRequest("POST", fmt.Sprintf("https://%s/%s/https:%s-%s", connectionData.Link, SERVICE_ENDPOINT, connectionData.WorkspaceID, SERVICE_EXEC_ENDPOINT), bytes.NewBuffer(payload))
	if err != nil {
		return "", err
	}
	request.Header = http.Header{
		"Authorization":            {"Bearer " + connectionData.Token},
		"Content-Type":             {"application/json"},
		"X-Forwarded-Access-Token": {connectionData.Token},
	}
	response, err := restClient.Do(request)
	if err != nil {
		return "", err
	}
	body, _ := io.ReadAll(response.Body)
	var workspaceRunning WorkspacePod
	json.Unmarshal(body, &workspaceRunning)
	pod := workspaceRunning.Pod
	return pod, nil
}

func GetUserName(connectionData ConnectionData) (string, error) {
	request, err := http.NewRequest("GET", fmt.Sprintf("https://%s/%s", connectionData.Link, USER_ENDPOINT), nil)
	if err != nil {
		return "", err
	}
	request.Header = http.Header{
		"Authorization": {"Bearer " + connectionData.Token},
		"Content-Type":  {"application/json"},
	}
	response, err := restClient.Do(request)
	if err != nil {
		return "", err
	}
	body, _ := io.ReadAll(response.Body)
	var user User
	json.Unmarshal(body, &user)
	return user.Metadata.Name, nil
}

func CleanAfterDisconnect(connectionData ConnectionData) {
	request, err := http.NewRequest("DELETE", fmt.Sprintf("https://%s/%s/%s", connectionData.Link, DEVWORKSPACE_ENDPOINT, connectionData.TerminalID), nil)
	if err != nil {
		log.Println(err)
		return
	}
	request.Header = http.Header{
		"Authorization": {"Bearer " + connectionData.Token},
	}
	response, err := restClient.Do(request)
	if err != nil {
		log.Println(err)
		return
	}
	if response.StatusCode != 200 {
		log.Println("Error while deleting workspace")
	}
}

func SendActivityTick(connectionData ConnectionData) error {
	request, err := http.NewRequest("POST", fmt.Sprintf("https://%s/%s/https:%s-%s", connectionData.Link, SERVICE_ENDPOINT, connectionData.WorkspaceID, SERVICE_ACTIVITY_ENDPOINT), nil)
	if err != nil {
		return err
	}
	log.Println("Sending activity tick")
	request.Header = http.Header{
		"Authorization":            {"Bearer " + connectionData.Token},
		"Content-Type":             {"application/json"},
		"X-Forwarded-Access-Token": {connectionData.Token},
	}
	_, err = restClient.Do(request)
	if err != nil {
		return err
	}
	return nil
}
