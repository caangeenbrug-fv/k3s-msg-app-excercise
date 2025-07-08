package model

type MessageRequest struct {
	Message     string   `json:"message"`
	Trace       []string `json:"trace"`
	SenderPodIP string   `json:"sender_pod_ip"`
}

type CustomMessageRequest struct {
	Message string `json:"message"`
}
