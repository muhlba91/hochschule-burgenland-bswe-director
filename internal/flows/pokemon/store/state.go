package store

import (
	"context"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/flows/pokemon/state"
)

// CreateState creates a new state in the cache.
// ctx: The context for the operation.
// state: The state to create.
func (w *Wrapper) CreateState(ctx context.Context, state *state.State) error {
	return w.stateStore.CreateState(ctx, state)
}

// UpdateState updates an existing state in the cache.
// ctx: The context for the operation.
// state: The state to update.
func (w *Wrapper) UpdateState(ctx context.Context, state *state.State) error {
	return w.stateStore.UpdateState(ctx, state)
}

// GetState retrieves a state from the cache based on its session ID.
// ctx: The context for the operation.
// sessionID: The unique session ID for the game session.
func (w *Wrapper) GetState(ctx context.Context, sessionID string) (*state.State, error) {
	return w.stateStore.GetState(ctx, sessionID)
}

// BroadcastState broadcasts the current state to all connected clients.
// ctx: The context for the operation.
// state: The state to broadcast.
func (w *Wrapper) BroadcastState(ctx context.Context, state *state.State) error {
	return w.broadcastStore.Broadcast(ctx, state.SessionID, state)
}

// BroadcastStateForSession retrieves the current state for a given session ID and broadcasts it to all connected clients.
// ctx: The context for the operation.
// sessionID: The unique session ID for the game session.
func (w *Wrapper) BroadcastStateForSession(ctx context.Context, sessionID string) error {
	state, err := w.GetState(ctx, sessionID)
	if err != nil || state == nil {
		return err
	}

	return w.broadcastStore.Broadcast(ctx, sessionID, state)
}
