package cache

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/cmd/configuration"
)

// Cache represents the redis cache.
type Cache struct {
	client *redis.Client
}

// NewCache creates a new redis cache.
// cfg: The configuration data for the cache.
func NewCache(cfg *configuration.Data) *Cache {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.RedisHost, cfg.RedisPort),
		Password: cfg.RedisPassword,
		DB:       0,
	})

	return &Cache{
		client: client,
	}
}

// IsConnected checks the connection to the redis server.
func (c *Cache) IsConnected() bool {
	if c.client == nil {
		return false
	}

	return c.client.Ping(context.Background()).Err() == nil
}

// Stop closes the redis client.
func (c *Cache) Stop() {
	if c.client == nil {
		return
	}

	logrus.Info("shutting down cache")
	if err := c.client.Close(); err != nil {
		logrus.Errorf("failed to shutdown cache: %v", err)
	}
}

// Exists checks if the given key exists in the cache.
// ctx: The context for the operation.
// key: The key to check for existence.
func (c *Cache) Exists(ctx context.Context, key string) (int64, error) {
	return c.client.Exists(ctx, key).Result()
}

// Subscribe subscribes to the given channels and returns a PubSub instance.
// ctx: The context for the operation.
// channels: The channels to subscribe to.
func (c *Cache) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return c.client.Subscribe(ctx, channels...)
}
