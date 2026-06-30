package pokemon

import (
	"context"
	"encoding/json"

	"github.com/sirupsen/logrus"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/flows/pokemon/constants"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/transport/websocket/connection"
	globalEvent "github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/transport/websocket/event"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/transport/websocket/message"
)

// Disconnect represents the disconnection from an existing Pokémon game.
// ctx: The context for managing request lifecycle.
// data: The data struct containing the session ID and connection information.
// connectionData: The connection data for the websocket connection.
func (gp *Gameplay) Disconnect(
	ctx context.Context,
	_ *connection.Data,
	connectionData *connection.Data,
) *message.Message {
	var sid *string
	msgEvent := message.ErrorEvent
	payload, _ := json.Marshal(message.ErrSessionDisconnected.Error())

	session, sErr := gp.store.GetSession(ctx, *connectionData.SessionID)
	switch {
	case sErr != nil:
		payload, _ = json.Marshal(sErr.Error())
	case session == nil:
		payload, _ = json.Marshal(message.ErrSessionNotFound.Error())
	default:
		disconnected := false
		switch connectionData.ConnectionID {
		case *session.PlayerA.ConnectionID:
			logrus.Debugf(
				"player A is disconnecting from the session with connection ID: %s, %s",
				*connectionData.SessionID,
				connectionData.ConnectionID,
			)
			session.PlayerA.Connected = false
			session.PlayerA.ConnectionID = nil
			disconnected = true
		case *session.PlayerB.ConnectionID:
			logrus.Debugf(
				"player B is disconnecting from the session with connection ID: %s, %s",
				*connectionData.SessionID,
				connectionData.ConnectionID,
			)
			session.PlayerB.Connected = false
			session.PlayerB.ConnectionID = nil
			disconnected = true
		default:
			logrus.Infof(
				"connection ID: %s is not associated with any player in session for disconnect: %s",
				connectionData.ConnectionID,
				*connectionData.SessionID,
			)
		}

		if !disconnected {
			break
		}

		uErr := gp.store.UpdateSession(ctx, session)
		if uErr == nil {
			msgEvent = globalEvent.GenerateEventName(constants.Name, globalEvent.Disconnected)
			payload, _ = json.Marshal(connectionData)

			gp.StartStop(ctx, session, connectionData)
		}
	}

	logrus.Debugf("Disconnect session response: sid=%v, event=%s, payload=%s", sid, msgEvent, payload)

	return &message.Message{
		Event:   msgEvent,
		Payload: payload,
	}
}

// ConnectionInformation provides the connection information for a given websocket connection.
// _: The context for managing request lifecycle.
// _: The connection.Data struct (not used in this function).
// connectionData: The connection data for the websocket connection.
func (gp *Gameplay) ConnectionInformation(
	_ context.Context,
	_ *connection.Data,
	connectionData *connection.Data,
) *message.Message {
	msgEvent := string(globalEvent.ConnectionInformation)
	payload, _ := json.Marshal(connectionData)

	logrus.Debugf("Connection information response: event=%s, payload=%s", msgEvent, payload)

	return &message.Message{
		Event:   msgEvent,
		Payload: payload,
	}
}
