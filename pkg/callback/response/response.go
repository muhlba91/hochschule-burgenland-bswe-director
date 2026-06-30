package response

const (
	responseTypeError   = "error"
	responseTypeSuccess = "success"
)

// Response represents the structure of a response.
type Response struct {
	Type    string `json:"type"`
	Message any    `json:"message"`
}

// newResponse creates a JSON-formatted response string with the given type and message.
// typ: The type of the response (e.g., "success", "error").
// message: The message to be included in the response.
func newResponse(typ string, message any) *Response {
	return &Response{
		Type:    typ,
		Message: message,
	}
}

// NewSuccess creates a JSON-formatted response string with the given message.
// message: The message to be included in the response.
func NewSuccess(message any) *Response {
	return newResponse(responseTypeSuccess, message)
}
