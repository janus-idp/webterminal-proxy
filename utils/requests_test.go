package utils

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	USERS_ENDPOINT             = "/apis/user.openshift.io/v1/users/~"
	WORKSPACE_ENDPOINT         = "/apis/workspace.devfile.io/v1alpha2/namespaces/test/devworkspaces/"
	WORKSPACE_SERVICE_ENDPOINT = "/api/v1/namespaces/test/services/https:test-workspace-service:4444/proxy/"
	TOKEN                      = "test-token"
	NAMESPACE                  = "test"
	WORKSPACE_ID               = "test-workspace"
	TERMINAL_ID                = "test-terminal"
	CONTAINER_NAME             = "test-container"
)

func TestSetupPod(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, WORKSPACE_SERVICE_ENDPOINT+"exec/init", r.URL.Path)
		assert.Equal(t, "Bearer "+TOKEN, r.Header.Get("Authorization"))
		assert.Equal(t, TOKEN, r.Header.Get("X-Forwarded-Access-Token"))

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"pod": "test-pod"}`))
	}))
	defer server.Close()

	connectionData := ConnectionData{
		Link:        strings.ReplaceAll(server.URL, "https://", ""),
		Namespace:   NAMESPACE,
		WorkspaceID: WORKSPACE_ID,
		Token:       TOKEN,
	}

	config := Config{
		Container: CONTAINER_NAME,
	}

	pod, err := SetupUserPod(connectionData, &config)
	assert.NoError(t, err)
	assert.Equal(t, "test-pod", pod)
}

func TestGetUserName(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, USERS_ENDPOINT, r.URL.Path)
		assert.Equal(t, "Bearer "+TOKEN, r.Header.Get("Authorization"))

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"metadata": {"name": "test-user"}}`))
	}))
	defer server.Close()

	connectionData := ConnectionData{
		Link:  strings.ReplaceAll(server.URL, "https://", ""),
		Token: TOKEN,
	}

	name, err := GetUserName(connectionData)
	assert.NoError(t, err)
	assert.Equal(t, "test-user", name)
}

func TestCleanAfterDisconnect(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, WORKSPACE_ENDPOINT+TERMINAL_ID, r.URL.Path)
		assert.Equal(t, "Bearer "+TOKEN, r.Header.Get("Authorization"))

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	connectionData := ConnectionData{
		Link:       strings.ReplaceAll(server.URL, "https://", ""),
		TerminalID: TERMINAL_ID,
		Namespace:  NAMESPACE,
		Token:      TOKEN,
	}
	var buffer bytes.Buffer
	log.SetOutput(&buffer)
	CleanAfterDisconnect(connectionData)
	assert.Empty(t, buffer.String())
}

func TestSendActivityTick(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, WORKSPACE_SERVICE_ENDPOINT+"activity/tick", r.URL.Path)
		assert.Equal(t, "Bearer "+TOKEN, r.Header.Get("Authorization"))
		assert.Equal(t, TOKEN, r.Header.Get("X-Forwarded-Access-Token"))

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	connectionData := ConnectionData{
		Link:        strings.ReplaceAll(server.URL, "https://", ""),
		Namespace:   NAMESPACE,
		WorkspaceID: WORKSPACE_ID,
		Token:       TOKEN,
	}

	err := SendActivityTick(connectionData)
	assert.NoError(t, err)
}

func TestSendPostRequest(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/test", r.URL.Path)
		assert.Equal(t, "test-token", r.Header.Get("Authorization"))
		w.Write([]byte(`{"output": "ok"}`))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	url := server.URL + "/api/test"
	payload := []byte(`{"test": "test"}`)
	authorizationToken := "test-token"
	body, statusCode, err := SendPostRequest(url, payload, authorizationToken)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, `{"output": "ok"}`, string(body))
}

func TestSendGetRequest(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/test", r.URL.Path)
		assert.Equal(t, "test-token", r.Header.Get("Authorization"))
		w.Write([]byte(`{"output": "ok"}`))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	url := server.URL + "/api/test"
	authorizationToken := "test-token"
	body, statusCode, err := SendGetRequest(url, authorizationToken)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, `{"output": "ok"}`, string(body))
}
