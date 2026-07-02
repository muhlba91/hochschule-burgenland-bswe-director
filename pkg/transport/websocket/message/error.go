package message

import "errors"

//nolint:gochecknoglobals // globals are fine for error definitions
var (
	ErrorEvent               = "error"
	ErrUnexpectedMessageType = errors.New("unexpected message type")
	ErrInvalidPayload        = errors.New("invalid payload")
	ErrSessionNotFound       = errors.New("session not found")
	ErrSessionNotCreated     = errors.New("session not created")
	ErrSessionNotJoined      = errors.New("session not joined")
	ErrSessionDisconnected   = errors.New("session disconnected")
	ErrSessionFull           = errors.New("session full")
	ErrSessionNotStarted     = errors.New("session could not be started")
	ErrSessionLocked         = errors.New("session is locked")
	ErrSessionStateNotFound  = errors.New("session state not found")
)
