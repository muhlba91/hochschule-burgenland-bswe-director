package pokemon

import (
	"context"
	"crypto/rand"
	"math/big"

	"github.com/sirupsen/logrus"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/app/logging"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/flows/pokemon/constants"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/flows/pokemon/action"
	pkgSession "github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/session"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/transport/callback"
)

// NextTurn initiates the next turn for a given session.
// ctx: The context for managing request-scoped values, cancellation signals, and deadlines.
// session: The current game session containing player connection information.
func (gp *Gameplay) NextTurn(ctx context.Context, session pkgSession.Session) error {
	logrus.WithFields(logrus.Fields{
		logging.FieldSessionID: session.GetID(),
	}).Info("starting next turn")

	pokemonSession, psErr := gp.store.GetSession(ctx, session.GetID())
	if psErr != nil || pokemonSession == nil {
		logrus.WithFields(logrus.Fields{
			logging.FieldSessionID: session.GetID(),
			logging.FieldError:     psErr,
		}).Error("failed to get session")
		return psErr
	}

	state, sErr := gp.store.GetState(ctx, session.GetID())
	if sErr != nil || state == nil {
		logrus.WithFields(logrus.Fields{
			logging.FieldSessionID: session.GetID(),
			logging.FieldError:     sErr,
		}).Error("failed to get state")
		return sErr
	}

	if pokemonSession.NextTurn == nil {
		keys := make([]string, 0, len(pokemonSession.Players))
		for k := range pokemonSession.Players {
			keys = append(keys, k)
		}
		pokemonSession.NextTurn = &pokemonSession.Players[generateFirstTurn(keys)].ID
	} else {
		pokemonSession.NextTurn = &pokemonSession.GetOpponentForID(*pokemonSession.NextTurn).ID
	}

	player := pokemonSession.Players[*pokemonSession.NextTurn]
	requestBuilder := func(_ string) (*callback.Request, callback.RequestData) {
		data := &action.Turn{
			RequestBase: action.RequestBase{
				RequestDataBase: callback.RequestDataBase{
					SessionID: session.GetID(),
				},
				State: state,
			},
		}

		request := &callback.Request{
			SessionID:       session.GetID(),
			InternalID:      player.ID,
			Action:          string(action.TypeTurn),
			Parallelization: constants.RequestTypeTurnParallelization,
		}

		return request, data
	}

	gp.requestor.Send(ctx, pokemonSession, []string{player.URL}, requestBuilder)
	return nil
}

// TurnCallback handles the callback for the turn action.
// ctx: The context for managing request-scoped values, cancellation signals, and deadlines.
// data: The data received from the callback, containing the state of the player who has taken their turn.
// session: The current game session containing player connection information.
// request: The callback request containing action details and session information.
func (gp *Gameplay) TurnCallback(
	ctx context.Context,
	data *action.TurnResult,
	session pkgSession.Session,
	request *callback.Request,
) error {
	state, sErr := gp.updateState(ctx, session.GetID(), request.InternalID, data.State)
	if sErr != nil {
		return sErr
	}

	pokemonSession, psErr := gp.store.GetSession(ctx, session.GetID())
	if psErr != nil || pokemonSession == nil {
		logrus.WithFields(logrus.Fields{
			logging.FieldSessionID: session.GetID(),
			logging.FieldError:     psErr,
		}).Error("failed to get session")
		return psErr
	}

	if data.Attack != nil {
		logrus.WithFields(logrus.Fields{
			logging.FieldPlayerID:  request.InternalID,
			logging.FieldSessionID: session.GetID(),
		}).Debug("player performed an attack")

		opponent := pokemonSession.GetOpponentForID(request.InternalID)

		requestBuilder := func(_ string) (*callback.Request, callback.RequestData) {
			data := &action.Attack{
				RequestBase: action.RequestBase{
					RequestDataBase: callback.RequestDataBase{
						SessionID: session.GetID(),
					},
				},
				Attack: data.Attack,
			}

			attack := &callback.Request{
				SessionID:       session.GetID(),
				InternalID:      opponent.ID,
				Action:          string(action.TypeAttack),
				Parallelization: constants.RequestTypeAttackParallelization,
			}

			return attack, data
		}

		gp.requestor.Send(ctx, pokemonSession, []string{opponent.URL}, requestBuilder)
		return nil
	}
	_ = gp.store.BroadcastState(ctx, state)
	return gp.analyzeState(ctx, session, state)
}

// AttackCallback handles the callback for the attack action.
// ctx: The context for managing request-scoped values, cancellation signals, and deadlines.
// data: The data received from the callback, containing the state of the player who has taken their turn.
// session: The current game session containing player connection information.
// request: The callback request containing action details and session information.
func (gp *Gameplay) AttackCallback(
	ctx context.Context,
	data *action.Attacked,
	session pkgSession.Session,
	request *callback.Request,
) error {
	state, sErr := gp.updateState(ctx, session.GetID(), request.InternalID, data.State)
	if sErr != nil {
		return sErr
	}

	return gp.analyzeState(ctx, session, state)
}

// generateFirstTurn generates a random player ID from the provided list of keys to determine who takes the first turn.
// keys: A slice of player IDs from which to randomly select the first player.
func generateFirstTurn(keys []string) string {
	nBig, err := rand.Int(rand.Reader, big.NewInt(int64(len(keys))))
	if err != nil {
		nBig = big.NewInt(0)
	}

	return keys[nBig.Int64()]
}
