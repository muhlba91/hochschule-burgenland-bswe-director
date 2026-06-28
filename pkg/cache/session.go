package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
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

	logrus.Debugf("saving session with ID: %s, data: %s", sessionID, session)
	if err := c.client.Set(ctx, sessionID, session, expiration).Err(); err != nil {
		logrus.Errorf("failed to save session: %v", err)
		return err
	}

	return nil
}

// UpdateSession updates an existing session in the cache with the given session ID and data.
// ctx: The context for the operation.
// sessionID: The unique session ID for the existing session.
// data: The data to be stored in the session.
// expiration: The expiration time for the session.
func (c *Cache) UpdateSession(ctx context.Context, sessionID string, data any, expiration time.Duration) error {
	logrus.Debugf("updating session with ID: %s", sessionID)
	return c.CreateSession(ctx, sessionID, data, expiration)
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

	logrus.Debugf("retrieved sessions for flow %s: %v", flowName, sessions)

	return sessions
}

// GetSession retrieves the session data for the given session ID from the cache.
// ctx: The context for the operation.
// sessionID: The unique session ID for the new session.
//
//nolint:nilnil // This function returns nil, nil when the session is not found, which is a valid case.
func (c *Cache) GetSession(ctx context.Context, sessionID string) (*string, error) {
	session, err := c.client.Get(ctx, sessionID).Result()
	logrus.Debugf("retrieved session data for key %s: %s", sessionID, session)

	if errors.Is(err, redis.Nil) {
		logrus.Warnf("session not found: %v", err)
		return nil, nil
	} else if err != nil {
		logrus.Errorf("failed to get session: %v", err)
		return nil, err
	}

	return &session, nil
}
