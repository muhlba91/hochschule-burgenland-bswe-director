package message

import "errors"

//nolint:gochecknoglobals // globals are fine for error definitions
var (
	ErrorEvent               = "error"
	ErrUnexpectedMessageType = errors.New("unexpected message type")
	ErrInvalidPayload        = errors.New("invalid payload")
	ErrSessionNotFound       = errors.New("session not found")
)
