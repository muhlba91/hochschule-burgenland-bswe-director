package state

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"

	globalStore "github.com/muhlba91/hochschule-burgenland-bswe-director/internal/store"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/store/constants"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/flows/pokemon/state"
)

// Store represents the state store for managing game states in the Pokémon flow.
type Store struct {
	store globalStore.Store
}

// NewStore creates a new instance of the Store.
// store: The global store for managing game states.
func NewStore(
	store globalStore.Store,
) *Store {
	return &Store{
		store: store,
	}
}

// CreateState creates a new state in the cache.
// ctx: The context for the operation.
// state: The state to create.
func (s *Store) CreateState(ctx context.Context, state *state.State) error {
	return s.store.Set(ctx, getStateKey(state.SessionID), state, defaultExpiration)
}

// UpdateState updates an existing state in the cache.
// ctx: The context for the operation.
// state: The state to update.
func (s *Store) UpdateState(ctx context.Context, state *state.State) error {
	return s.store.Set(ctx, getStateKey(state.SessionID), state, defaultExpiration)
}

// GetState retrieves a state from the cache based on its session ID.
// ctx: The context for the operation.
// sessionID: The unique session ID for the game session.
func (s *Store) GetState(ctx context.Context, sessionID string) (*state.State, error) {
	stateData, err := s.store.Get(ctx, getStateKey(sessionID))
	if err != nil || stateData == nil {
		return nil, err
	}

	var state state.State
	if sErr := json.Unmarshal([]byte(*stateData), &state); sErr != nil {
		logrus.Debugf("error occurred while unmarshaling state: %v", sErr)
		return nil, sErr
	}

	return &state, nil
}

// GetStateKey generate the key for storing the state in the cache based on the session ID.
// sessionID: The unique session ID for the game session.
func getStateKey(sessionID string) string {
	return fmt.Sprintf("%s:%s", sessionID, constants.StateChannel)
}
