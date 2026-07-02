package redis

import (
	"fmt"
	"log/slog"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/app/configuration"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/app/logging"
)

// Cache represents the redis cache.
type Cache struct {
	client  *redis.Client
	redsync *redsync.Redsync
}

// NewCache creates a new redis cache.
// cfg: The configuration data for the cache.
func NewCache(cfg *configuration.Data) *Cache {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.RedisHost, cfg.RedisPort),
		Password: cfg.RedisPassword,
		DB:       0,
	})

	pool := goredis.NewPool(client)
	rs := redsync.New(pool)

	return &Cache{
		client:  client,
		redsync: rs,
	}
}

// Stop closes the redis client.
func (c *Cache) Stop() {
	if c.client == nil {
		return
	}

	slog.Info("shutting down cache")
	if err := c.client.Close(); err != nil {
		slog.Error("failed to shutdown cache", slog.Any(logging.FieldError, err))
	}
}
