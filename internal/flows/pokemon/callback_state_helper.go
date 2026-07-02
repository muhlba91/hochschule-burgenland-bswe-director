package pokemon

import (
	"context"
	"log/slog"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/app/logging"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/flows/pokemon/state"
)

// updateState updates the state of a player in the game session.
// ctx: The context for managing request-scoped values, cancellation signals, and deadlines.
// data: The data received from the callback, containing the state of the player who has taken their turn.
// session: The current game session containing player connection information.
// request: The callback request containing action details and session information.
func (gp *Gameplay) updateState(
	ctx context.Context,
	sessionID string,
	playerID string,
	playerState *state.Player,
) (*state.State, error) {
	state, sErr := gp.store.GetState(ctx, sessionID)
	if sErr != nil || state == nil {
		slog.Error("failed to get state",
			slog.String(logging.FieldSessionID, sessionID),
			slog.Any(logging.FieldError, sErr),
		)
		return nil, sErr
	}

	state.Players[playerID] = playerState

	uErr := gp.store.UpdateState(ctx, state)
	if uErr != nil {
		slog.Error("failed to update state",
			slog.String(logging.FieldSessionID, sessionID),
			slog.Any(logging.FieldError, uErr),
		)
		return nil, uErr
	}

	return state, nil
}
