package store

import (
	"context"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/flows/pokemon/session"
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

// GetGlobalStateForSession retrieves the global state for a given session ID.
// ctx: The context for the operation.
// sessionID: The unique session ID for the game session.
func (w *Wrapper) GetGlobalStateForSession(ctx context.Context, sessionID string) (*state.GlobalState, error) {
	state, err := w.stateStore.GetState(ctx, sessionID)
	if err != nil || state == nil {
		return nil, err
	}

	session, sErr := w.GetSession(ctx, sessionID)
	if sErr != nil || session == nil {
		return nil, sErr
	}

	return w.ToGlobalState(state, session), nil
}

// GetCurrentStateForPlayerAndSession retrieves the current state for a specific player in a given session.
// ctx: The context for the operation.
// playerID: The ID of the player for whom to retrieve the current state.
// sessionID: The unique session ID for the game session.
func (w *Wrapper) GetCurrentStateForPlayerAndSession(
	ctx context.Context,
	playerID string,
	sessionID string,
) (*state.CurrentState, error) {
	state, err := w.stateStore.GetState(ctx, sessionID)
	if err != nil || state == nil {
		return nil, err
	}

	session, sErr := w.GetSession(ctx, sessionID)
	if sErr != nil || session == nil {
		return nil, sErr
	}

	return w.ToCurrentStateForPlayer(state, session, playerID), nil
}

// ToCurrentStateForPlayer creates a CurrentState for a specific player.
// gameState: The State to migrate.
// session: The session data for the Pokémon game session.
// currentPlayer: The ID of the current player.
func (w *Wrapper) ToCurrentStateForPlayer(
	gameState *state.State,
	session *session.Session,
	currentPlayer string,
) *state.CurrentState {
	currentState := &state.CurrentState{
		Session:   session,
		Player:    gameState.Players[currentPlayer],
		Opponents: make(map[string]*state.Opponent),
	}

	for playerID, player := range gameState.Players {
		if playerID != currentPlayer {
			currentState.Opponents[playerID] = &state.Opponent{
				Active:      player.Active,
				Bench:       player.Bench,
				Hand:        len(player.Hand),
				Deck:        len(player.Deck),
				DiscardPile: player.DiscardPile,
				PrizeCards:  len(player.PrizeCards),
			}
		}
	}

	return currentState
}

// ToGlobalState creates a CurrentState for a specific player.
// gameState: The State to migrate.
// session: The session data for the Pokémon game session.
func (w *Wrapper) ToGlobalState(gameState *state.State, session *session.Session) *state.GlobalState {
	globalState := &state.GlobalState{
		Session: session,
		Players: make(map[string]*state.Opponent),
	}

	for playerID, player := range gameState.Players {
		globalState.Players[playerID] = &state.Opponent{
			Active:      player.Active,
			Bench:       player.Bench,
			Hand:        len(player.Hand),
			Deck:        len(player.Deck),
			DiscardPile: player.DiscardPile,
			PrizeCards:  len(player.PrizeCards),
		}
	}

	return globalState
}

// BroadcastGlobalState broadcasts the global state to all connected clients.
// ctx: The context for the operation.
// gameState: The state to broadcast.
func (w *Wrapper) BroadcastGlobalState(ctx context.Context, gameState *state.GlobalState) error {
	return w.broadcastStore.Broadcast(ctx, gameState.Session.GetID(), gameState)
}

// BroadcastStateForSession retrieves the current state for a given session ID, then broadcasts it to all connected clients.
// ctx: The context for the operation.
// sessionID: The unique session ID for the game session.
func (w *Wrapper) BroadcastStateForSession(ctx context.Context, sessionID string) error {
	state, err := w.GetGlobalStateForSession(ctx, sessionID)
	if err != nil || state == nil {
		return err
	}

	return w.BroadcastGlobalState(ctx, state)
}
