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
	ChatID  int64  `json:"chat_id,omitempty"`
	Error   string `json:"error,omitempty"`
}

const (
	RequestTypeSend      = "send"
	RequestTypeGetChatID = "get_chat_id"
)
