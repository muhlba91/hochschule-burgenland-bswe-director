package callback

import (
	"errors"
)

var (
	ErrRequestNotAllowed        = errors.New("request not allowed due to parallelization restrictions")
	ErrInvalidRequest           = errors.New("invalid request")
	ErrTransportFailure         = errors.New("transport failure")
	ErrNotAccepted              = errors.New("request not accepted")
	ErrRequestNotSaved          = errors.New("request not saved")
	ErrSessionNotUpdated        = errors.New("session not updated")
	ErrCallbackNotReceived      = errors.New("callback not received within the expected time frame")
	ErrCallbackContextCancelled = errors.New("callback context cancelled")
)
