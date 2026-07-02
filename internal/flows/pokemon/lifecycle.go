package pokemon

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/app/logging"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/flows/pokemon/constants"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/flows/pokemon/action"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/flows/pokemon/event"
	pkgPokemonSession "github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/flows/pokemon/session"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/flows/pokemon/state"
	pkgSession "github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/session"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/transport"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/transport/callback"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/transport/callback/response"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/transport/websocket/connection"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/transport/websocket/message"
)

// StartStop checks if both players are connected and broadcasts the appropriate event to the clients.
// ctx: The context for managing request-scoped values, cancellation signals, and deadlines.
// session: The current game session containing player connection information.
// connectionData: The data related to the WebSocket connection, including session ID and other relevant details.
func (gp *Gameplay) StartStop(
	ctx context.Context,
	session *pkgPokemonSession.Session,
	connectionData *connection.Data,
) {
	evnt := event.TypeNotStarted
	if session.Winner != nil {
		evnt = event.TypeFinished
	}

	allConnected := true
	for _, player := range session.Players {
		if !player.Connected {
			allConnected = false
			break
		}
	}

	if session.StartedAt != 0 {
		evnt = event.TypePaused
		if allConnected {
			evnt = event.TypeResumed
		}
	}
	if session.StartedAt == 0 && allConnected && len(session.Players) == 2 {
		evnt = event.TypeStarted
	}

	slog.Debug("broadcasting start/stop event",
		slog.String(logging.FieldEvent, string(evnt)),
		slog.String(logging.FieldSessionID, session.GetID()),
	)
	payload, _ := json.Marshal(connectionData)
	_ = gp.store.Broadcast(ctx, session.GetID(), &message.Message{
		Event:   gp.generateEventName(evnt),
		Payload: payload,
	})

	slog.Debug("broadcasting state",
		slog.String(logging.FieldSessionID, session.GetID()),
	)
	_ = gp.store.BroadcastStateForSession(ctx, session.GetID())

	switch evnt {
	case event.TypeNotStarted:
		slog.Debug("session is not started yet", slog.String(logging.FieldSessionID, session.GetID()))
	case event.TypeStarted:
		slog.Debug("session is started", slog.String(logging.FieldSessionID, session.GetID()))
		go gp.Start(context.WithoutCancel(ctx), session)
	case event.TypePaused:
		slog.Debug("session is paused", slog.String(logging.FieldSessionID, session.GetID()))
		_ = gp.requestor.CleanRequestQueue(ctx, session)
	case event.TypeResumed:
		slog.Debug("session is resumed", slog.String(logging.FieldSessionID, session.GetID()))
		go gp.NextTurn(context.WithoutCancel(ctx), session)
	case event.TypeFinished:
		slog.Debug("session is already finished", slog.String(logging.FieldSessionID, session.GetID()))
		_ = gp.broadcastWinner(ctx, session)
	default:
		slog.Warn("unknown event",
			slog.String(logging.FieldEvent, string(evnt)),
			slog.String(logging.FieldSessionID, session.GetID()),
		)
	}
}

// Start initiates the gameplay for a given session.
// ctx: The context for managing request-scoped values, cancellation signals, and deadlines.
// session: The current game session containing player connection information.
func (gp *Gameplay) Start(ctx context.Context, session *pkgPokemonSession.Session) {
	slog.Info("starting gameplay",
		slog.String(logging.FieldSessionID, session.GetID()),
	)

	session.StartedAt = time.Now().Unix()
	uErr := gp.store.UpdateSession(ctx, session)
	if uErr != nil {
		slog.Error("failed to update session with start time",
			slog.String(logging.FieldSessionID, session.GetID()),
			slog.Any(logging.FieldError, uErr),
		)
		gp.reportBackgroundError(ctx, session.GetID(), message.ErrSessionNotStarted)
		return
	}

	state := &state.State{
		SessionID: session.GetID(),
	}
	sErr := gp.store.CreateState(ctx, state)
	if sErr != nil {
		slog.Error("failed to create state",
			slog.String(logging.FieldSessionID, session.GetID()),
			slog.Any(logging.FieldError, sErr),
		)
		gp.reportBackgroundError(ctx, session.GetID(), message.ErrSessionNotStarted)
		return
	}

	requestBuilder := func(url string) (*callback.Request, callback.RequestData) {
		data := &action.Start{
			RequestBase: action.RequestBase{
				RequestDataBase: callback.RequestDataBase{
					SessionID: session.GetID(),
				},
			},
		}

		request := &callback.Request{
			SessionID:       session.GetID(),
			InternalID:      session.GetPlayerByURL(url).ID,
			Action:          string(action.TypeStart),
			Parallelization: constants.RequestTypeStartParallelization,
		}

		return request, data
	}

	gp.requestor.Send(ctx, session, session.GetPlayerURLs(), requestBuilder)
}

// StartCallback handles the callback for the start action.
// ctx: The context for managing request-scoped values, cancellation signals, and deadlines.
// data: The data received from the callback, containing the state of the player who has started the game.
// session: The current game session containing player connection information.
// request: The callback request containing action details and session information.
func (gp *Gameplay) StartCallback(
	ctx context.Context,
	data *action.Started,
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

	if len(session.GetNextRequests()) == 0 {
		slog.DebugContext(ctx, "all start callbacks received, initiating next turn",
			slog.String(logging.FieldSessionID, session.GetID()),
		)

		pokemonSession, psErr := gp.store.GetSession(ctx, session.GetID())
		if psErr != nil || pokemonSession == nil {
			return psErr
		}
		_ = gp.store.BroadcastGlobalState(ctx, gp.store.ToGlobalState(state, pokemonSession))
		go gp.NextTurn(context.WithoutCancel(ctx), session)
	}

	return nil
}

// analyzeState analyzes the current state of the game and determines the next action to take.
// ctx: The context for managing request-scoped values, cancellation signals, and deadlines.
// session: The current game session containing player connection information.
// state: The current state of the game, including player statuses and other relevant information.
func (gp *Gameplay) analyzeState(ctx context.Context, session pkgSession.Session, state *state.State) error {
	pokemonSession, psErr := gp.store.GetSession(ctx, session.GetID())
	if psErr != nil || pokemonSession == nil {
		slog.ErrorContext(ctx, "failed to get session",
			slog.String(logging.FieldSessionID, session.GetID()),
			slog.Any(logging.FieldError, psErr),
		)
		return psErr
	}

	for playerID, playerState := range state.Players {
		opponent := pokemonSession.GetOpponentForID(playerID)

		if len(playerState.PrizeCards) == 0 {
			slog.InfoContext(ctx, "player has taken all prize cards, session finished",
				slog.String(logging.FieldPlayerID, playerID),
				slog.String(logging.FieldSessionID, pokemonSession.GetID()),
			)
			pokemonSession.Winner = &playerID
			break
		}

		if playerState.Active == nil || playerState.Active.HP <= 0 {
			slog.InfoContext(ctx, "player's active pokemon is dead, opponent wins",
				slog.String(logging.FieldPlayerID, playerID),
				slog.String("opponent_id", opponent.ID),
				slog.String(logging.FieldSessionID, pokemonSession.GetID()),
			)
			pokemonSession.Winner = &opponent.ID
			break
		}

		if len(playerState.Deck) == 0 {
			slog.InfoContext(ctx, "player's deck is empty, opponent wins",
				slog.String(logging.FieldPlayerID, playerID),
				slog.String("opponent_id", opponent.ID),
				slog.String(logging.FieldSessionID, pokemonSession.GetID()),
			)
			pokemonSession.Winner = &opponent.ID
			break
		}
	}

	if pokemonSession.Winner != nil {
		if err := gp.store.UpdateSession(ctx, pokemonSession); err != nil {
			slog.ErrorContext(ctx, "failed to update session with winner",
				slog.String(logging.FieldSessionID, pokemonSession.GetID()),
				slog.Any(logging.FieldError, err),
			)
		}

		return gp.broadcastWinner(ctx, pokemonSession)
	}

	go gp.NextTurn(context.WithoutCancel(ctx), session)
	return nil
}

// broadcastWinner broadcasts the winner of the game session to all connected clients.
// ctx: The context for managing request-scoped values, cancellation signals, and deadlines.
// session: The current game session containing player connection information and the winner.
func (gp *Gameplay) broadcastWinner(ctx context.Context, session *pkgPokemonSession.Session) error {
	if session.Winner == nil {
		slog.WarnContext(ctx, "no winner to broadcast",
			slog.String(logging.FieldSessionID, session.GetID()),
		)
		return nil
	}

	payload, _ := json.Marshal(session)
	return gp.store.Broadcast(ctx, session.GetID(), &message.Message{
		Event:   gp.generateEventName(event.TypeFinished),
		Payload: payload,
	})
}
