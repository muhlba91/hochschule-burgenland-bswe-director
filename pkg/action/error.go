package action

import (
	"errors"
)

var (
	ErrInvalidAction       = errors.New("invalid action")
	ErrNoHandlerRegistered = errors.New("no handler registered for action")
)
