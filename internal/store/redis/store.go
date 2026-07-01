package redis

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
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
	logrus.Debugf("retrieved data for key %s: %s", key, data)

	if errors.Is(err, redis.Nil) || data == "" {
		logrus.Warnf("key not found: %v", err)
		return nil, nil
	} else if err != nil {
		logrus.Errorf("failed to get data: %v", err)
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

	logrus.Debugf("saving data for key %s: %s", key, data)
	if err := c.client.Set(ctx, key, data, expiration).Err(); err != nil {
		logrus.Errorf("failed to save data: %v", err)
		return err
	}

	return nil
}

// Delete removes the key-value pair associated with the given key from the cache.
// ctx: The context for the operation.
// key: The key to be deleted from the cache.
func (c *Cache) Delete(ctx context.Context, key string) error {
	if err := c.client.Del(ctx, key).Err(); err != nil {
		logrus.Errorf("failed to delete data: %v", err)
		return err
	}

	return nil
}
