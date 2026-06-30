package model

// Request represents a request in the system.
type Request struct {
	// ID is the unique identifier for the request.
	ID string `json:"id" redis:"id"`
	// SessionID is the unique identifier for the session associated with the request.
	SessionID string `json:"session_id" redis:"session_id"`
	// Endpoint is the endpoint to which the request is directed.
	Endpoint string `json:"endpoint" redis:"endpoint"`
	// Body contains the body of the request.
	Body string `json:"data" redis:"data"`
	// Action is the action associated with the request.
	Action string `json:"action" redis:"action"`
	// Blocking indicates whether the request is blocking or non-blocking.
	Blocking bool `json:"blocking" redis:"blocking"`
	// Singleton indicates whether the request must run as the only open request.
	Singleton bool `json:"singleton" redis:"singleton"`
	// CreatedAt is the timestamp when the request was created.
	CreatedAt int64 `json:"created_at" redis:"created_at"`
	// Completed indicates whether the request has been completed.
	Completed bool `json:"completed" redis:"completed"`
}
