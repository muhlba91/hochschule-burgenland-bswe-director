package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/app/logging"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/store"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/store/constants"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/session"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/transport/callback/response"
)

// maxBatchSize defines the maximum number of session keys to retrieve in a single batch when listing sessions.
const maxBatchSize = 100

// CreateSession creates a new session in the cache with the given session ID and data.
// ctx: The context for the operation.
// sessionID: The unique session ID for the new session.
// data: The data to be stored in the session.
func (c *Cache) CreateSession(ctx context.Context, sessionID string, data any) error {
	if err := c.Set(ctx, sessionID, data, constants.DefaultSessionExpiration); err != nil {
		logrus.WithFields(logrus.Fields{
			logging.FieldSessionID: sessionID,
			logging.FieldError:     err,
		}).Error("failed to save session")
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

	var keys []string
	var cursor uint64

	for {
		var batch []string
		var err error
		batch, cursor, err = c.client.Scan(ctx, cursor, pattern, maxBatchSize).Result()
		if err != nil {
			logrus.WithFields(logrus.Fields{
				logging.FieldFlowName: flowName,
				logging.FieldError:    err,
			}).Error("failed to scan session keys")
			return sessions
		}

		keys = append(keys, batch...)
		if cursor == 0 {
			break
		}
	}

	if len(keys) == 0 {
		return sessions
	}

	const batchSize = maxBatchSize
	for i := 0; i < len(keys); i += batchSize {
		end := min(i+batchSize, len(keys))

		batchKeys := keys[i:end]
		vals, err := c.client.MGet(ctx, batchKeys...).Result()
		if err != nil {
			logrus.WithFields(logrus.Fields{
				logging.FieldFlowName: flowName,
				logging.FieldError:    err,
			}).Error("failed to retrieve session data")
			continue
		}

		for j, key := range batchKeys {
			if vals[j] != nil {
				if val, ok := vals[j].(string); ok {
					sessions[key] = val
				}
			}
		}
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

// GetBaseSession retrieves the base session data for the given session ID from the cache.
// ctx: The context for the operation.
// sessionID: The unique session ID for the new session.
func (c *Cache) GetBaseSession(ctx context.Context, sessionID string) (session.Session, error) {
	sess, sErr := c.GetSession(ctx, sessionID)
	if sErr != nil || sess == nil {
		logrus.Debugf("session not found: %s", sessionID)
		return nil, echo.NewHTTPError(http.StatusGone, response.NewError(response.ErrNoMatchingSession))
	}

	var session session.Base
	usErr := json.Unmarshal([]byte(*sess), &session)
	if usErr != nil {
		logrus.WithFields(logrus.Fields{
			logging.FieldSessionID: sessionID,
			logging.FieldError:     usErr,
		}).Error("failed to unmarshal session")
		return nil, echo.NewHTTPError(http.StatusInternalServerError, response.NewError(response.ErrNoMatchingSession))
	}

	return &session, nil
}

// LockSession locks the session with the given session ID to prevent concurrent access.
// ctx: The context for the operation.
// sessionID: The unique session ID for the session to be locked.
func (c *Cache) LockSession(ctx context.Context, sessionID string) (store.UnlockFunc, error) {
	return c.Lock(ctx, sessionID)
}
