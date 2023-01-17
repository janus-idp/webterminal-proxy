package utils

type ConnectionData struct {
	Link         string
	Token        string
	PodID        string
	TerminalID   string
	WorkspaceID  string
	TerminalSize []string
}

type KubeConfig struct {
	Username  string
	Namespace string
}
type Config struct {
	Container  string
	Kubeconfig KubeConfig
}

type WorkspacePod struct {
	Pod string `json:"pod"`
}

type User struct {
	Metadata struct {
		Name string `json:"name"`
	} `json:"metadata"`
}
