package redis

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/app/logging"
)

// IsConnected checks the connection to the redis server.
func (c *Cache) IsConnected() bool {
	if c.client == nil {
		return false
	}

	return c.client.Ping(context.Background()).Err() == nil
}

// Exists checks if the given key exists in the cache.
// ctx: The context for the operation.
// key: The key to check for existence.
func (c *Cache) Exists(ctx context.Context, key string) (bool, error) {
	count, err := c.client.Exists(ctx, key).Result()
	return count > 0, err
}

// Get retrieves the value associated with the given key from the cache.
// ctx: The context for the operation.
// key: The key to retrieve the value for.
//
//nolint:nilnil // This function returns nil, nil when the key is not found, which is a valid case.
func (c *Cache) Get(ctx context.Context, key string) (*string, error) {
	data, err := c.client.Get(ctx, key).Result()
	logrus.WithFields(logrus.Fields{
		logging.FieldKey:  key,
		logging.FieldData: data,
	}).Debug("retrieved data from redis")

	if errors.Is(err, redis.Nil) || data == "" {
		logrus.WithFields(logrus.Fields{
			logging.FieldKey:   key,
			logging.FieldError: err,
		}).Debug("key not found in redis")
		return nil, nil
	} else if err != nil {
		logrus.WithFields(logrus.Fields{
			logging.FieldKey:   key,
			logging.FieldError: err,
		}).Error("failed to get data from redis")
		return nil, err
	}

	return &data, nil
}

// Set stores the given value in the cache with the specified key and expiration duration.
// ctx: The context for the operation.
// key: The key to store the value under.
// value: The value to be stored in the cache.
// expiration: The duration after which the key-value pair should expire.
func (c *Cache) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	data, _ := json.Marshal(value)

	logrus.WithFields(logrus.Fields{
		logging.FieldKey:  key,
		logging.FieldData: string(data),
	}).Debug("saving data to redis")
	if err := c.client.Set(ctx, key, data, expiration).Err(); err != nil {
		logrus.WithFields(logrus.Fields{
			logging.FieldKey:   key,
			logging.FieldError: err,
		}).Error("failed to save data to redis")
		return err
	}

	return nil
}

// Delete removes the key-value pair associated with the given key from the cache.
// ctx: The context for the operation.
// key: The key to be deleted from the cache.
func (c *Cache) Delete(ctx context.Context, key string) error {
	if err := c.client.Del(ctx, key).Err(); err != nil {
		logrus.WithFields(logrus.Fields{
			logging.FieldKey:   key,
			logging.FieldError: err,
		}).Error("failed to delete data from redis")
		return err
	}

	return nil
}
