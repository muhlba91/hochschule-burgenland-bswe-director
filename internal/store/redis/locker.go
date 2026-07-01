package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/sirupsen/logrus"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/store"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/store/constants"
)

// Lock locks the given key in the cache.
// ctx: The context for the operation.
// key: The key to lock.
func (c *Cache) Lock(ctx context.Context, key string) (store.UnlockFunc, error) {
	return c.LockWithOptions(
		ctx,
		key,
		constants.DefaultLockExpiry,
		constants.DefaultLockTries,
		constants.DefaultLockDelay,
	)
}

// LockWithOptions locks the given key in the cache with custom options.
// ctx: The context for the operation.
// key: The key to lock.
// expiry: The duration for which the lock will be held.
// tries: The number of attempts to acquire the lock.
// delay: The delay between each attempt to acquire the lock.
func (c *Cache) LockWithOptions(
	ctx context.Context,
	key string,
	expiry time.Duration,
	tries int,
	delay time.Duration,
) (store.UnlockFunc, error) {
	lockKey := fmt.Sprintf("%s:%s", key, constants.LockChannel)

	mutex := c.redsync.NewMutex(
		lockKey,
		redsync.WithExpiry(expiry),
		redsync.WithTries(tries),
		redsync.WithRetryDelay(delay),
	)

	logrus.Debugf("attempting to acquire lock for key: %s", lockKey)
	if err := mutex.LockContext(ctx); err != nil {
		logrus.Errorf("failed to acquire lock for key %s: %v", lockKey, err)
		return nil, err
	}
	logrus.Debugf("acquired lock for key: %s", lockKey)

	return func(uCtx context.Context) {
		logrus.Debugf("releasing lock for key: %s", lockKey)
		_, err := mutex.UnlockContext(uCtx)
		if err != nil {
			logrus.Errorf("error releasing lock for key %s: %v", lockKey, err)
		}
	}, nil
}
