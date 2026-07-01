package callback

// RequestData represents the data associated with a request in the system.
type RequestData interface {
	// GetSessionID returns the unique identifier for the game session associated with the request.
	GetSessionID() string

	// SetCallback sets the callback URL for the request.
	SetCallback(callback string)
}

// RequestDataBase is a base struct that implements the RequestData interface.
type RequestDataBase struct {
	// Callback is the URL to which the requestor should send the response.
	Callback string `json:"callback"`
	// SessionID is the unique identifier for the game session associated with the request.
	SessionID string `json:"sessionId"`
}

// GetSessionID returns the unique identifier for the game session associated with the request.
func (r *RequestDataBase) GetSessionID() string {
	return r.SessionID
}

// SetCallback sets the callback URL for the request.
func (r *RequestDataBase) SetCallback(callback string) {
	r.Callback = callback
}
