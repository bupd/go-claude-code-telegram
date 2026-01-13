package ipc

type Request struct {
	Type    string `json:"type"`
	Session string `json:"session"`
	Message string `json:"message"`
	Timeout int    `json:"timeout"`
	WorkDir string `json:"workdir"`
}

type Response struct {
	Success bool   `json:"success"`
	Reply   string `json:"reply"`
	Error   string `json:"error,omitempty"`
}

const (
	RequestTypeSend = "send"
)
