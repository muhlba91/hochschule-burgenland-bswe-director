package cache

import (
	"context"
	"encoding/json"

	"github.com/sirupsen/logrus"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/flows/pokemon/constants"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/flows/pokemon/model"
)

// GenerateUniqueSessionID generates a unique session ID for the given flow name.
// ctx: The context for the operation.
func (w *Wrapper) GenerateUniqueSessionID(ctx context.Context) *string {
	return w.cache.GenerateUniqueSessionID(ctx, constants.Name)
}

// CreateSession creates a new session in the cache with the given session ID and data.
// ctx: The context for the operation.
// sessionID: The unique session ID for the new session.
// data: The data to be stored in the session.
func (w *Wrapper) CreateSession(ctx context.Context, session *model.Session) error {
	return w.cache.CreateSession(ctx, session.ID, session, sessionExpirationTime)
}

// UpdateSession updates an existing session in the cache with the given session ID and data.
// ctx: The context for the operation.
// sessionID: The unique session ID for the existing session.
// data: The data to be stored in the session.
func (w *Wrapper) UpdateSession(ctx context.Context, session *model.Session) error {
	return w.cache.UpdateSession(ctx, session.ID, session, sessionExpirationTime)
}

// ListSessions lists all session IDs for the given flow name.
// ctx: The context for the operation.
func (w *Wrapper) ListSessions(ctx context.Context) map[string]*model.Session {
	sessions := make(map[string]*model.Session)

	rawSessions := w.cache.ListSessions(ctx, constants.Name)
	for sid, s := range rawSessions {
		var l model.Session
		if err := json.Unmarshal([]byte(s), &l); err == nil {
			sessions[sid] = &l
		}
	}

	return sessions
}

// GetSession retrieves the session data for the given session ID from the cache.
// ctx: The context for the operation.
// sessionID: The unique session ID for the new session.
func (w *Wrapper) GetSession(ctx context.Context, sessionID string) (*model.Session, error) {
	sessionData, sErr := w.cache.GetSession(ctx, sessionID)
	if sErr != nil {
		return nil, sErr
	}

	var session model.Session
	if err := json.Unmarshal([]byte(*sessionData), &session); err != nil {
		logrus.Debugf("error occurred while unmarshaling session data: %v", err)
		return nil, err
	}

	return &session, nil
}
