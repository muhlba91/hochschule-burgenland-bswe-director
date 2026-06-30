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
	if session.StartedAt != 0 {
		evnt = event.TypePaused
		if session.PlayerA.Connected && session.PlayerB.Connected {
			evnt = event.TypeResumed
		}
	}
	if session.StartedAt == 0 && session.PlayerA.Connected && session.PlayerB.Connected {
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
		// FIXME: initiate next turn
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

	requestBuilder := func() (*callback.Request, callback.RequestData) {
		data := &action.Start{
			SessionID: session.GetID(),
		}
		request := &callback.Request{
			SessionID:       session.GetID(),
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
// data: The current state of the game session.
// session: The current game session containing player connection information.
// request: The callback request containing action details and session information.
func (gp *Gameplay) StartCallback(
	ctx context.Context,
	data *state.State,
	session pkgSession.Session,
	request *callback.Request,
) (any, error) {
	// FIXME: implement - take care of parallel requests!
	return nil, nil
}
