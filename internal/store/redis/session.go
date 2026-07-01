package redis

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/store"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/store/constants"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/session"
)

// CreateSession creates a new session in the cache with the given session ID and data.
// ctx: The context for the operation.
// sessionID: The unique session ID for the new session.
// data: The data to be stored in the session.
func (c *Cache) CreateSession(ctx context.Context, sessionID string, data any) error {
	if err := c.Set(ctx, sessionID, data, constants.DefaultSessionExpiration); err != nil {
		logrus.Errorf("failed to save session: %v", err)
		return err
	}

	return nil
}

// UpdateSession updates an existing session in the cache with the given session ID and data.
// ctx: The context for the operation.
// sessionID: The unique session ID for the existing session.
// data: The data to be stored in the session.
func (c *Cache) UpdateSession(ctx context.Context, sessionID string, data any) error {
	logrus.Debugf("updating session with ID: %s", sessionID)
	return c.CreateSession(ctx, sessionID, data)
}

// ListSessions lists all session IDs for the given flow name.
// ctx: The context for the operation.
// flowName: The name of the flow for which to list session IDs.
func (c *Cache) ListSessions(ctx context.Context, flowName string) map[string]string {
	sessions := make(map[string]string)

	pattern := fmt.Sprintf("%s:*", session.GetSessionPrefix(flowName))
	keys, kErr := c.client.Keys(ctx, pattern).Result()
	if kErr != nil {
		return sessions
	}

	for _, key := range keys {
		val, gEerr := c.Get(ctx, key)
		if gEerr != nil {
			logrus.Errorf("failed to get session data for key %s: %v", key, gEerr)
			continue
		}
		sessions[key] = *val
	}

	logrus.Debugf("retrieved sessions for flow %s: %v", flowName, sessions)

	return sessions
}

// GetSession retrieves the session data for the given session ID from the cache.
// ctx: The context for the operation.
// sessionID: The unique session ID for the new session.
func (c *Cache) GetSession(ctx context.Context, sessionID string) (*string, error) {
	return c.Get(ctx, sessionID)
}

// LockSession locks the session with the given session ID to prevent concurrent access.
// ctx: The context for the operation.
// sessionID: The unique session ID for the session to be locked.
func (c *Cache) LockSession(ctx context.Context, sessionID string) (store.UnlockFunc, error) {
	return c.Lock(ctx, sessionID)
}
