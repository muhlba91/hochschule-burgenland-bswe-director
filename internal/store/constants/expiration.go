package constants

import "time"

const (
	// DefaultRequestExpiration defines the default expiration time for requests in seconds.
	DefaultRequestExpiration = 1 * 60 * 60 * time.Second

	// DefaultSessionExpiration defines the default expiration time for sessions in seconds.
	DefaultSessionExpiration = 24 * 60 * 60 * time.Second

	// DefaultLockExpiry defines the default expiration time for locks in seconds.
	DefaultLockExpiry = 5 * time.Second

	// DefaultLockTries defines the default number of attempts to acquire a lock.
	DefaultLockTries = 100

	// DefaultLockDelay defines the default delay between each attempt to acquire a lock.
	DefaultLockDelay = 100 * time.Millisecond
)
