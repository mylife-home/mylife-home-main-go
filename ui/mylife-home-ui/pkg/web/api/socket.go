package api

import "encoding/json"

type MessageType string

const (
	// server to client
	MessageState     MessageType = "state"
	MessageAdd       MessageType = "add"
	MessageRemove    MessageType = "remove"
	MessageChange    MessageType = "change"
	MessageModelHash MessageType = "modelHash"
	MessagePong      MessageType = "pong"

	// client to server
	MessagePing   MessageType = "ping"
	MessageAction MessageType = "action"
)

type SocketMessage struct {
	Type MessageType     `json:"type" tstype:"\"state\" | \"add\" | \"remove\" | \"change\" | \"modelHash\" | \"ping\" | \"pong\" | \"action\""`
	Data json.RawMessage `json:"data"`
}
