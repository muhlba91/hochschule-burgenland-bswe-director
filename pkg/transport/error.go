package transport

import (
	"errors"
)

var (
	ErrNoMatchingSession = errors.New("no matching session found")
	ErrSessionLocked     = errors.New("session is locked")
)
