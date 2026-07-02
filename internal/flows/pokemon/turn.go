package pokemon

import (
	"context"
	"crypto/rand"
	"log/slog"
	"math/big"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/app/logging"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/flows/pokemon/constants"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/flows/pokemon/action"
	pkgSession "github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/session"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/transport"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/transport/callback"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/transport/callback/response"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/transport/websocket/message"
)

// NextTurn initiates the next turn for a given session.
// ctx: The context for managing request-scoped values, cancellation signals, and deadlines.
// session: The current game session containing player connection information.
func (gp *Gameplay) NextTurn(ctx context.Context, session pkgSession.Session) {
	unlock, lErr := gp.store.LockSession(ctx, session.GetID())
	if lErr != nil {
		slog.ErrorContext(ctx, "failed to lock session",
			slog.String(logging.FieldSessionID, session.GetID()),
			slog.Any(logging.FieldError, lErr),
		)
		gp.reportBackgroundError(ctx, session.GetID(), message.ErrSessionLocked)
		return
	}
	defer unlock(ctx)

	slog.InfoContext(ctx, "starting next turn",
		slog.String(logging.FieldSessionID, session.GetID()),
	)

	pokemonSession, psErr := gp.store.GetSession(ctx, session.GetID())
	if psErr != nil || pokemonSession == nil {
		slog.ErrorContext(ctx, "failed to get session",
			slog.String(logging.FieldSessionID, session.GetID()),
			slog.Any(logging.FieldError, psErr),
		)
		gp.reportBackgroundError(ctx, session.GetID(), message.ErrSessionNotFound)
		return
	}

	state, sErr := gp.store.GetState(ctx, session.GetID())
	if sErr != nil || state == nil {
		slog.ErrorContext(ctx, "failed to get state",
			slog.String(logging.FieldSessionID, session.GetID()),
			slog.Any(logging.FieldError, sErr),
		)
		gp.reportBackgroundError(ctx, session.GetID(), message.ErrSessionStateNotFound)
		return
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
				State: gp.store.ToCurrentStateForPlayer(state, pokemonSession, *pokemonSession.NextTurn),
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

	go gp.requestor.Send(context.WithoutCancel(ctx), pokemonSession, []string{player.URL}, requestBuilder)
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
	unlock, err := gp.store.LockSession(ctx, request.SessionID)
	if err != nil {
		return echo.NewHTTPError(http.StatusLocked, response.NewError(transport.ErrSessionLocked))
	}
	defer unlock(ctx)

	state, sErr := gp.updateState(ctx, session.GetID(), request.InternalID, data.State)
	if sErr != nil {
		return sErr
	}

	pokemonSession, psErr := gp.store.GetSession(ctx, session.GetID())
	if psErr != nil || pokemonSession == nil {
		slog.ErrorContext(ctx, "failed to get session",
			slog.String(logging.FieldSessionID, session.GetID()),
			slog.Any(logging.FieldError, psErr),
		)
		return psErr
	}

	if data.Attack != nil {
		slog.DebugContext(ctx, "player performed an attack",
			slog.String(logging.FieldPlayerID, request.InternalID),
			slog.String(logging.FieldSessionID, session.GetID()),
		)

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

		go gp.requestor.Send(context.WithoutCancel(ctx), pokemonSession, []string{opponent.URL}, requestBuilder)
		return nil
	}

	_ = gp.store.BroadcastGlobalState(ctx, gp.store.ToGlobalState(state, pokemonSession))
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
	unlock, err := gp.store.LockSession(ctx, request.SessionID)
	if err != nil {
		return echo.NewHTTPError(http.StatusLocked, response.NewError(transport.ErrSessionLocked))
	}
	defer unlock(ctx)

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
