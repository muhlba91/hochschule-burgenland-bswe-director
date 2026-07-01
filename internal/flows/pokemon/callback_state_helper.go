package pokemon

import (
	"context"

	"github.com/sirupsen/logrus"

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
		logrus.Errorf("failed to get state for session %s: %v", sessionID, sErr)
		return nil, sErr
	}

	state.Players[playerID] = playerState

	uErr := gp.store.UpdateState(ctx, state)
	if uErr != nil {
		logrus.Errorf("failed to update state for session %s: %v", sessionID, uErr)
		return nil, uErr
	}

	return state, nil
}
