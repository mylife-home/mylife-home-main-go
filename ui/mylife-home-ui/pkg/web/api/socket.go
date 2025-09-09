package api

type SocketMessage struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}
