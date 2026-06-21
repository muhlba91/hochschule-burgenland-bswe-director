package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// maxIterations defines the maximum number of attempts to generate a unique session ID before giving up.
const maxIterations = 5

// GenerateUniqueSessionID generates a unique session ID for the given flow name.
// ctx: The context for the operation.
// flowName: The name of the flow for which to generate the session ID.
func (c *Cache) GenerateUniqueSessionID(ctx context.Context, flowName string) *string {
	var id string

	for range maxIterations {
		id = uuid.NewString()
		key := fmt.Sprintf("session:%s:%s", flowName, id)

		exists, err := c.Exists(ctx, key)
		if err != nil {
			logrus.WithContext(ctx).Errorf("failed to check if session exists: %v", err)
			continue
		}
		if exists == 0 {
			return &key
		}
	}

	return nil
}

// CreateSession creates a new session in the cache with the given session ID and data.
// ctx: The context for the operation.
// sessionID: The unique session ID for the new session.
// data: The data to be stored in the session.
// expiration: The expiration time for the session.
func (c *Cache) CreateSession(ctx context.Context, sessionID string, data any, expiration time.Duration) error {
	session, _ := json.Marshal(data)
	if err := c.client.Set(ctx, sessionID, session, expiration).Err(); err != nil {
		logrus.Errorf("failed to save session: %v", err)
		return err
	}

	return nil
}

// ListSessions lists all session IDs for the given flow name.
// ctx: The context for the operation.
// flowName: The name of the flow for which to list session IDs.
func (c *Cache) ListSessions(ctx context.Context, flowName string) map[string]string {
	sessions := make(map[string]string)

	pattern := fmt.Sprintf("session:%s:*", flowName)
	keys, kErr := c.client.Keys(ctx, pattern).Result()
	if kErr != nil {
		return sessions
	}

	for _, key := range keys {
		val, gEerr := c.client.Get(ctx, key).Result()
		if gEerr != nil {
			logrus.Errorf("failed to get session data for key %s: %v", key, gEerr)
			continue
		}
		sessions[key] = val
	}

	return sessions
}
