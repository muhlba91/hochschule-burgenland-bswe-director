package response

import (
	"errors"
)

var (
	ErrInvalidPayload          = errors.New("invalid payload")
	ErrNoMatchingRequest       = errors.New("no matching request found")
	ErrRequestAlreadyCompleted = errors.New("request already completed")
	ErrActionFailed            = errors.New("action failed")
	ErrCompletionFailed        = errors.New("completion failed")
)

// NewError creates a JSON-formatted error response string from the given error.
// err: The error to be included in the response.
func NewError(err error) *Response {
	return newResponse(responseTypeError, err.Error())
}
