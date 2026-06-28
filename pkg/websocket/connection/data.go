package connection

import "github.com/google/uuid"

// Data represents the data associated with a websocket connection.
type Data struct {
	// ConnectionID is the unique identifier for the websocket connection.
	ConnectionID string `json:"connection_id"`
	// SessionID is the unique identifier for the session associated with the connection.
	SessionID *string `json:"session_id,omitempty"`
}

// NewData creates a new instance of Data with a unique ConnectionID.
func NewData() *Data {
	return &Data{
		ConnectionID: uuid.NewString(),
		SessionID:    nil,
	}
}
