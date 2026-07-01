package response

import (
	"errors"
)

var (
	ErrInvalidPayload      = errors.New("invalid payload")
	ErrNoMatchingRequest   = errors.New("no matching request found")
	ErrNoMatchingSession   = errors.New("no matching session found")
	ErrSessionLocked       = errors.New("session is locked")
	ErrCallbackNotExpected = errors.New("callback not expected for this session")
	ErrActionFailed        = errors.New("action failed")
	ErrCompletionFailed    = errors.New("completion failed")
)

// NewError creates a JSON-formatted error response string from the given error.
// err: The error to be included in the response.
func NewError(err error) *Response {
	return newResponse(responseTypeError, err.Error())
}
