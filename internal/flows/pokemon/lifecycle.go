package pokemon

import (
	"context"
	"encoding/json"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/flows/pokemon/constants"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/flows/pokemon/action"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/flows/pokemon/event"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/flows/pokemon/session"
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
func (gp *Gameplay) StartStop(ctx context.Context, session *session.Session, connectionData *connection.Data) {
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

	logrus.Debugf("broadcasting start/stop event: %s for session: %s", evnt, session.GetID())
	payload, _ := json.Marshal(connectionData)
	_ = gp.store.Broadcast(ctx, session.GetID(), &message.Message{
		Event:   gp.generateEventName(evnt),
		Payload: payload,
	})

	// FIXME: broadcast winner and last state again (if it still exists)

	switch evnt {
	case event.TypeNotStarted:
		logrus.Debugf("session %s is not started yet", session.GetID())
	case event.TypeStarted:
		logrus.Debugf("session %s is started", session.GetID())
		gp.Start(ctx, session)
	case event.TypePaused:
		logrus.Debugf("session %s is paused", session.GetID())
		_ = gp.requestor.CleanRequestQueue(ctx, session)
	case event.TypeResumed:
		logrus.Debugf("session %s is resumed", session.GetID())
		gp.NextTurn(ctx, session)
	case event.TypeFinished:
		logrus.Debugf("session %s is already finished", session.GetID())
	default:
		logrus.Warnf("unknown event: %s for session: %s", evnt, session.GetID())
	}
}

// Start initiates the gameplay for a given session.
// ctx: The context for managing request-scoped values, cancellation signals, and deadlines.
// session: The current game session containing player connection information.
func (gp *Gameplay) Start(ctx context.Context, session *session.Session) {
	logrus.Infof("starting gameplay for session: %s", session.GetID())

	session.StartedAt = time.Now().Unix()
	uErr := gp.store.UpdateSession(ctx, session)
	if uErr != nil {
		logrus.Errorf("failed to update session %s with start time: %v", session.GetID(), uErr)
	}

	state := &state.State{
		SessionID: session.GetID(),
	}
	sErr := gp.store.CreateState(ctx, state)
	if sErr != nil {
		logrus.Errorf("failed to create state for session %s: %v", session.GetID(), sErr)
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

	err := gp.requestor.Send(ctx, session, session.GetPlayerURLs(), requestBuilder)
	if err != nil {
		logrus.Errorf("failed to send start requests for session %s: %v", session.GetID(), err)
	}
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
		logrus.Debugf("all start callbacks received for session %s, initiating next turn", session.GetID())
		_ = gp.store.BroadcastState(ctx, state)
		gp.NextTurn(ctx, session)
	}

	return nil
}

// analyzeState analyzes the current state of the game and determines the next action to take.
// ctx: The context for managing request-scoped values, cancellation signals, and deadlines.
// session: The current game session containing player connection information.
// state: The current state of the game, including player statuses and other relevant information.
func (gp *Gameplay) analyzeState(ctx context.Context, session pkgSession.Session, state *state.State) {
	// FIXME: implement winner detection and broadcast winner event
	gp.NextTurn(ctx, session)
}
