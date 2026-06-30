package callback

// RequestData represents the data associated with a request in the system.
type RequestData interface {
	// SetCallback sets the callback URL for the request.
	SetCallback(callback string)
}

// RequestDataBase is a base struct that implements the RequestData interface.
type RequestDataBase struct {
	// Callback is the URL to which the requestor should send the response.
	Callback string `json:"callback"`
}

// SetCallback sets the callback URL for the request.
func (r *RequestDataBase) SetCallback(callback string) {
	r.Callback = callback
}
