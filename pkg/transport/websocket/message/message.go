package message

import (
	"encoding/json"
)

// Message represents a websocket message.
type Message struct {
	// Event is the type of the message.
	Event string `json:"event"`
	// Payload is the payload of the message.
	Payload json.RawMessage `json:"payload"`
}
