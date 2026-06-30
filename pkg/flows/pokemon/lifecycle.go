package pokemon

import (
	"context"
	"encoding/json"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/flows/pokemon/event"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/flows/pokemon/model"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/websocket/connection"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/websocket/message"
)

// StartStop checks if both players are connected and broadcasts the appropriate event to the clients.
// ctx: The context for managing request-scoped values, cancellation signals, and deadlines.
// session: The current game session containing player connection information.
// connectionData: The data related to the WebSocket connection, including session ID and other relevant details.
func (gp *Gameplay) StartStop(ctx context.Context, session *model.Session, connectionData *connection.Data) {
	evnt := event.NotStarted
	if session.Winner != nil {
		evnt = event.Finished
	}
	if session.StartedAt != 0 {
		evnt = event.Paused
		if session.PlayerA.Connected && session.PlayerB.Connected {
			evnt = event.Resumed
		}
	}
	if session.StartedAt == 0 && session.PlayerA.Connected && session.PlayerB.Connected {
		evnt = event.Started
	}

	logrus.Debugf("broadcasting start/stop event: %s for session: %s", evnt, session.ID)
	payload, _ := json.Marshal(connectionData)
	_ = gp.Cache.Broadcast(ctx, session.ID, &message.Message{
		Event:   gp.generateEventName(evnt),
		Payload: payload,
	})

	// FIXME: broadcast winner and last state again (if it still exists)

	switch evnt {
	case event.NotStarted:
		logrus.Debugf("session %s is not started yet", session.ID)
	case event.Started:
		logrus.Debugf("session %s is started", session.ID)
		gp.Start(ctx, session)
	case event.Paused:
		logrus.Debugf("session %s is paused", session.ID)
		_ = gp.Requestor.CleanRequestQueue(ctx, session)
	case event.Resumed:
		logrus.Debugf("session %s is resumed", session.ID)
		// FIXME: initiate next turn
	case event.Finished:
		logrus.Debugf("session %s is already finished", session.ID)
	default:
		logrus.Warnf("unknown event: %s for session: %s", evnt, session.ID)
	}
}

// Start initiates the gameplay for a given session.
// ctx: The context for managing request-scoped values, cancellation signals, and deadlines.
// session: The current game session containing player connection information.
func (gp *Gameplay) Start(ctx context.Context, session *model.Session) {
	logrus.Infof("starting gameplay for session: %s", session.ID)

	session.StartedAt = time.Now().Unix()
	uErr := gp.Cache.UpdateSession(ctx, session)
	if uErr != nil {
		logrus.Errorf("failed to update session %s with start time: %v", session.ID, uErr)
	}

	// FIXME: implement
}
