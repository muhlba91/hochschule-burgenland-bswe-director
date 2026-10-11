package response

import "encoding/json"

// HTTPError is an error carrying an HTTP status code and a JSON error response body.
// It implements the status coder and JSON marshaler interfaces understood by echo's error handler.
type HTTPError struct {
	Code     int
	Response *Response
}

// NewHTTPError creates a new HTTPError with the given status code and error.
// code: The HTTP status code.
// err: The error to be included in the response body.
func NewHTTPError(code int, err error) *HTTPError {
	return &HTTPError{Code: code, Response: NewError(err)}
}

// StatusCode returns the HTTP status code.
func (e *HTTPError) StatusCode() int {
	return e.Code
}

// Error returns the error message.
func (e *HTTPError) Error() string {
	if msg, ok := e.Response.Message.(string); ok {
		return msg
	}

	return e.Response.Type
}

// MarshalJSON renders the response body as {"message": <response>}.
func (e *HTTPError) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{"message": e.Response})
}
