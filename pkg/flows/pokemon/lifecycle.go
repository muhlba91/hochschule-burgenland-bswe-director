package pokemon

import (
	"context"
	"encoding/json"

	"github.com/sirupsen/logrus"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/cache"
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
	evnt := event.Paused
	if session.PlayerA.Connected && session.PlayerB.Connected {
		evnt = event.Started
	}
	if session.Winner != nil {
		evnt = event.Finished
	}

	logrus.Debugf("broadcasting start/stop event: %s for session: %s", evnt, *connectionData.SessionID)

	payload, _ := json.Marshal(connectionData)
	_ = gp.Cache.Broadcast(ctx, *connectionData.SessionID, cache.Director, &message.Message{
		Event:   gp.generateEventName(evnt),
		Payload: payload,
	})
}
