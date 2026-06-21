package cache

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/cmd/configuration"
)

// Cache represents the redis cache.
type Cache struct {
	Client *redis.Client
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
		Client: client,
	}
}

// IsConnected checks the connection to the redis server.
func (c *Cache) IsConnected() bool {
	if c.Client == nil {
		return false
	}

	return c.Client.Ping(context.Background()).Err() == nil
}

// Stop closes the redis client.
func (c *Cache) Stop() {
	if c.Client == nil {
		return
	}

	logrus.Info("shutting down cache")
	if err := c.Client.Close(); err != nil {
		logrus.Errorf("failed to shutdown cache: %v", err)
	}
}

// Exists checks if the given key exists in the cache.
// ctx: The context for the operation.
// key: The key to check for existence.
func (c *Cache) Exists(ctx context.Context, key string) (int64, error) {
	return c.Client.Exists(ctx, key).Result()
}

// Subscribe subscribes to the given channels and returns a PubSub instance.
// ctx: The context for the operation.
// channels: The channels to subscribe to.
func (c *Cache) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return c.Client.Subscribe(ctx, channels...)
}

// GenerateUniqueSessionID generates a unique session ID for the given flow name.
// ctx: The context for the operation.
// flowName: The name of the flow for which to generate the session ID.
func (c *Cache) GenerateUniqueSessionID(ctx context.Context, flowName string) string {
	var id string

	for {
		id = uuid.NewString()
		key := fmt.Sprintf("session:%s:%s", flowName, id)

		exists, err := c.Exists(ctx, key)
		if err != nil {
			logrus.Errorf("failed to check if session exists: %v", err)
			continue
		}
		if exists == 0 {
			break
		}
	}

	return id
}
