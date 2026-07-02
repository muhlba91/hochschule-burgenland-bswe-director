package pokemon

import (
	"context"
	"encoding/json"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/app/logging"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/flows/pokemon/constants"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/flows/pokemon/action"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/flows/pokemon/event"
	pkgPokemonSession "github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/flows/pokemon/session"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/flows/pokemon/state"
	pkgSession "github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/session"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/transport/callback"
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

	logrus.WithFields(logrus.Fields{
		logging.FieldEvent:     evnt,
		logging.FieldSessionID: session.GetID(),
	}).Debug("broadcasting start/stop event")
	payload, _ := json.Marshal(connectionData)
	_ = gp.store.Broadcast(ctx, session.GetID(), &message.Message{
		Event:   gp.generateEventName(evnt),
		Payload: payload,
	})

	logrus.WithFields(logrus.Fields{
		logging.FieldSessionID: session.GetID(),
	}).Debug("broadcasting state")
	_ = gp.store.BroadcastStateForSession(ctx, session.GetID())

	switch evnt {
	case event.TypeNotStarted:
		logrus.WithFields(logrus.Fields{logging.FieldSessionID: session.GetID()}).Debug("session is not started yet")
	case event.TypeStarted:
		logrus.WithFields(logrus.Fields{logging.FieldSessionID: session.GetID()}).Debug("session is started")
		gp.Start(ctx, session)
	case event.TypePaused:
		logrus.WithFields(logrus.Fields{logging.FieldSessionID: session.GetID()}).Debug("session is paused")
		_ = gp.requestor.CleanRequestQueue(ctx, session)
	case event.TypeResumed:
		logrus.WithFields(logrus.Fields{logging.FieldSessionID: session.GetID()}).Debug("session is resumed")
		_ = gp.NextTurn(ctx, session)
	case event.TypeFinished:
		logrus.WithFields(logrus.Fields{logging.FieldSessionID: session.GetID()}).Debug("session is already finished")
		_ = gp.broadcastWinner(ctx, session)
	default:
		logrus.WithFields(logrus.Fields{
			logging.FieldEvent:     evnt,
			logging.FieldSessionID: session.GetID(),
		}).Warn("unknown event")
	}
}

// Start initiates the gameplay for a given session.
// ctx: The context for managing request-scoped values, cancellation signals, and deadlines.
// session: The current game session containing player connection information.
func (gp *Gameplay) Start(ctx context.Context, session *pkgPokemonSession.Session) {
	logrus.WithFields(logrus.Fields{
		logging.FieldSessionID: session.GetID(),
	}).Info("starting gameplay")

	session.StartedAt = time.Now().Unix()
	uErr := gp.store.UpdateSession(ctx, session)
	if uErr != nil {
		logrus.WithFields(logrus.Fields{
			logging.FieldSessionID: session.GetID(),
			logging.FieldError:     uErr,
		}).Error("failed to update session with start time")
	}

	state := &state.State{
		SessionID: session.GetID(),
	}
	sErr := gp.store.CreateState(ctx, state)
	if sErr != nil {
		logrus.WithFields(logrus.Fields{
			logging.FieldSessionID: session.GetID(),
			logging.FieldError:     sErr,
		}).Error("failed to create state")
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
	state, sErr := gp.updateState(ctx, session.GetID(), request.InternalID, data.State)
	if sErr != nil {
		return sErr
	}

	if len(session.GetNextRequests()) == 0 {
		logrus.WithFields(logrus.Fields{
			logging.FieldSessionID: session.GetID(),
		}).Debug("all start callbacks received, initiating next turn")

		pokemonSession, psErr := gp.store.GetSession(ctx, session.GetID())
		if psErr != nil || pokemonSession == nil {
			return psErr
		}
		_ = gp.store.BroadcastGlobalState(ctx, gp.store.ToGlobalState(state, pokemonSession))
		return gp.NextTurn(ctx, session)
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
		logrus.WithFields(logrus.Fields{
			logging.FieldSessionID: session.GetID(),
			logging.FieldError:     psErr,
		}).Error("failed to get session")
		return psErr
	}

	for playerID, playerState := range state.Players {
		opponent := pokemonSession.GetOpponentForID(playerID)

		if len(playerState.PrizeCards) == 0 {
			logrus.WithFields(logrus.Fields{
				logging.FieldPlayerID:  playerID,
				logging.FieldSessionID: pokemonSession.GetID(),
			}).Info("player has taken all prize cards, session finished")
			pokemonSession.Winner = &playerID
			break
		}

		if playerState.Active == nil || playerState.Active.HP <= 0 {
			logrus.WithFields(logrus.Fields{
				logging.FieldPlayerID:  playerID,
				"opponent_id":          opponent.ID,
				logging.FieldSessionID: pokemonSession.GetID(),
			}).Info("player's active pokemon is dead, opponent wins")
			pokemonSession.Winner = &opponent.ID
			break
		}

		if len(playerState.Deck) == 0 {
			logrus.WithFields(logrus.Fields{
				logging.FieldPlayerID:  playerID,
				"opponent_id":          opponent.ID,
				logging.FieldSessionID: pokemonSession.GetID(),
			}).Info("player's deck is empty, opponent wins")
			pokemonSession.Winner = &opponent.ID
			break
		}
	}

	if pokemonSession.Winner != nil {
		if err := gp.store.UpdateSession(ctx, pokemonSession); err != nil {
			logrus.WithFields(logrus.Fields{
				logging.FieldSessionID: pokemonSession.GetID(),
				logging.FieldError:     err,
			}).Error("failed to update session with winner")
		}

		return gp.broadcastWinner(ctx, pokemonSession)
	}

	return gp.NextTurn(ctx, session)
}

// broadcastWinner broadcasts the winner of the game session to all connected clients.
// ctx: The context for managing request-scoped values, cancellation signals, and deadlines.
// session: The current game session containing player connection information and the winner.
func (gp *Gameplay) broadcastWinner(ctx context.Context, session *pkgPokemonSession.Session) error {
	if session.Winner == nil {
		logrus.WithFields(logrus.Fields{
			logging.FieldSessionID: session.GetID(),
		}).Warn("no winner to broadcast")
		return nil
	}

	payload, _ := json.Marshal(session)
	return gp.store.Broadcast(ctx, session.GetID(), &message.Message{
		Event:   gp.generateEventName(event.TypeFinished),
		Payload: payload,
	})
}
