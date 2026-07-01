package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/sirupsen/logrus"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/app/logging"
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

	logrus.WithFields(logrus.Fields{
		logging.FieldLockKey: lockKey,
	}).Debug("attempting to acquire lock")
	if err := mutex.LockContext(ctx); err != nil {
		logrus.WithFields(logrus.Fields{
			logging.FieldLockKey: lockKey,
			logging.FieldError:   err,
		}).Error("failed to acquire lock")
		return nil, err
	}
	logrus.WithFields(logrus.Fields{
		logging.FieldLockKey: lockKey,
	}).Debug("acquired lock")

	return func(uCtx context.Context) {
		logrus.WithFields(logrus.Fields{
			logging.FieldLockKey: lockKey,
		}).Debug("releasing lock")
		_, err := mutex.UnlockContext(uCtx)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				logging.FieldLockKey: lockKey,
				logging.FieldError:   err,
			}).Error("error releasing lock")
		}
	}, nil
}
