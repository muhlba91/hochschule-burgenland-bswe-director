package constants

import "time"

const (
	// DefaultRequestExpiration defines the default expiration time for requests in seconds.
	DefaultRequestExpiration = 1 * 60 * 60 * time.Second
	// DefaultSessionExpiration defines the default expiration time for sessions in seconds.
	DefaultSessionExpiration = 24 * 60 * 60 * time.Second
)
