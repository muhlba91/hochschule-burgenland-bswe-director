package store

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/flows/pokemon/constants"
	pokemonSession "github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/flows/pokemon/session"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/session"
)

// GenerateUniqueSessionID generates a unique session ID for the given flow name.
func (w *Wrapper) GenerateUniqueSessionID() string {
	return session.GetSessionID(constants.Name, uuid.NewString())
}

// CreateSession creates a new session in the cache with the given session ID and data.
// ctx: The context for the operation.
// sessionID: The unique session ID for the new session.
// data: The data to be stored in the session.
func (w *Wrapper) CreateSession(ctx context.Context, session session.Session) error {
	return w.sessionStore.CreateSession(ctx, session.GetID(), session)
}

// UpdateSession updates an existing session in the cache with the given session ID and data.
// ctx: The context for the operation.
// sessionID: The unique session ID for the existing session.
// data: The data to be stored in the session.
func (w *Wrapper) UpdateSession(ctx context.Context, session session.Session) error {
	return w.sessionStore.UpdateSession(ctx, session.GetID(), session)
}

// ListSessions lists all session IDs for the given flow name.
// ctx: The context for the operation.
func (w *Wrapper) ListSessions(ctx context.Context) map[string]*pokemonSession.Session {
	sessions := make(map[string]*pokemonSession.Session)

	rawSessions := w.sessionStore.ListSessions(ctx, constants.Name)
	for sid, s := range rawSessions {
		var l pokemonSession.Session
		if err := json.Unmarshal([]byte(s), &l); err == nil {
			sessions[sid] = &l
		}
	}

	return sessions
}

// GetSession retrieves the session data for the given session ID from the cache.
// ctx: The context for the operation.
// sessionID: The unique session ID for the new session.
func (w *Wrapper) GetSession(ctx context.Context, sessionID string) (*pokemonSession.Session, error) {
	sessionData, sErr := w.sessionStore.GetSession(ctx, sessionID)
	if sErr != nil {
		return nil, sErr
	}

	var session pokemonSession.Session
	if err := json.Unmarshal([]byte(*sessionData), &session); err != nil {
		logrus.Debugf("error occurred while unmarshaling session data: %v", err)
		return nil, err
	}

	return &session, nil
}
