package store

import (
	"context"
	"time"
)

// UnlockFunc defines a function type that represents an unlock operation for a lock.
// ctx: The context for the operation.
type UnlockFunc func(ctx context.Context)

// Locker defines the interface for managing locks in the store. It provides methods to acquire and release locks for specific keys, allowing for synchronization and coordination between different processes or goroutines.
type Locker interface {
	// Lock acquires a lock for the given key. If the lock is already held by another process, it will block until the lock is available.
	// ctx: The context for the operation.
	// key: The unique identifier for the lock.
	Lock(ctx context.Context, key string) (UnlockFunc, error)

	// LockWithOptions acquires a lock for the given key with additional options. If the lock is already held by another process, it will retry acquiring the lock based on the specified number of tries and delay between attempts.
	// ctx: The context for the operation.
	// key: The unique identifier for the lock.
	// expiry: The duration for which the lock will be held before it expires.
	// tries: The number of attempts to acquire the lock before giving up.
	// delay: The duration to wait between each attempt to acquire the lock.
	LockWithOptions(
		ctx context.Context,
		key string,
		expiry time.Duration,
		tries int,
		delay time.Duration,
	) (UnlockFunc, error)
}
