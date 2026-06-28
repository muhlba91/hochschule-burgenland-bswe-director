package pokemon

import (
	"context"
	"encoding/json"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/flows/pokemon/event"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/flows/pokemon/model"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/websocket/connection"
	globalEvent "github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/websocket/event"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/websocket/message"
)

// expirationTime defines the duration for which a session remains valid before it expires.
const expirationTime = 24 * time.Hour

// Create represents the creation of a new Pokémon game.
// ctx: The context for managing request lifecycle.
// create: The Create struct containing the players for the new game.
// _: The connection data for the websocket connection.
func (gp *Gameplay) Create(ctx context.Context, create *model.Create, _ *connection.Data) (*string, *message.Message) {
	msgEvent := message.ErrorEvent
	payload, _ := json.Marshal(message.ErrSessionNotCreated.Error())

	sid := gp.Cache.GenerateUniqueSessionID(ctx, gp.Name)

	if sid != nil {
		err := gp.Cache.CreateSession(ctx, *sid, &model.Session{
			PlayerA:  create.PlayerA,
			PlayerB:  create.PlayerB,
			NextTurn: nil,
			Winner:   nil,
		}, expirationTime)
		if err == nil {
			msgEvent = gp.generateEventName(event.Created)
			payload, _ = json.Marshal(&model.Created{
				SessionID: *sid,
			})
		}
	}

	logrus.Debugf("Create session response: sid=%v, event=%s, payload=%s", *sid, msgEvent, payload)

	return nil, &message.Message{
		Event:   msgEvent,
		Payload: payload,
	}
}

// List represents the listing of existing Pokémon games.
// ctx: The context for managing request lifecycle.
// _ : The List struct (not used in this function).
// _ : The connection.Data struct (not used in this function).
func (gp *Gameplay) List(ctx context.Context, _ *model.List, _ *connection.Data) *message.Message {
	sessions := make(map[string]model.Session)

	rawSessions := gp.Cache.ListSessions(ctx, gp.Name)
	for sid, s := range rawSessions {
		var l model.Session
		if err := json.Unmarshal([]byte(s), &l); err == nil {
			sessions[sid] = l
		}
	}

	payload, _ := json.Marshal(&model.Listing{
		Sessions: sessions,
	})

	logrus.Debugf("List sessions response: payload=%s", payload)

	return &message.Message{
		Event:   gp.generateEventName(event.Listing),
		Payload: payload,
	}
}

// Join represents the joining of an existing Pokémon game.
// ctx: The context for managing request lifecycle.
// join: The Join struct containing the session ID and player information.
// connectionData: The connection data for the websocket connection.
func (gp *Gameplay) Join(
	ctx context.Context,
	join *model.Join,
	connectionData *connection.Data,
) (*string, *message.Message) {
	sid := &join.SessionID
	msgEvent := message.ErrorEvent
	payload, _ := json.Marshal(message.ErrSessionNotJoined.Error())

	sessionData, sErr := gp.Cache.GetSession(ctx, *sid)
	switch {
	case sErr != nil:
		payload, _ = json.Marshal(sErr.Error())
	case sessionData == nil:
		payload, _ = json.Marshal(message.ErrSessionNotFound.Error())
	default:
		var session model.Session
		if err := json.Unmarshal([]byte(*sessionData), &session); err != nil {
			logrus.Debugf("error occurred while unmarshaling session data: %v", err)
			payload, _ = json.Marshal(err.Error())
			break
		}

		connected := false
		if session.PlayerA.URL == join.Player && !session.PlayerA.Connected {
			logrus.Debugf(
				"player A is joining the session with connection ID: %s, %s",
				*sid,
				connectionData.ConnectionID,
			)
			session.PlayerA.Connected = true
			session.PlayerA.ConnectionID = &connectionData.ConnectionID
			connected = true
		} else if session.PlayerB.URL == join.Player && !session.PlayerB.Connected {
			logrus.Debugf(
				"player B is joining the session with connection ID: %s, %s",
				*sid,
				connectionData.ConnectionID,
			)
			session.PlayerB.Connected = true
			session.PlayerB.ConnectionID = &connectionData.ConnectionID
			connected = true
		}

		if !connected {
			logrus.Debugf("player %s is not a valid player for session: %s", join.Player, *sid)
			break
		}

		uErr := gp.Cache.UpdateSession(ctx, *sid, &session, expirationTime)
		if uErr == nil {
			msgEvent = gp.generateEventName(event.Joined)
			payload, _ = json.Marshal(&model.Joined{
				SessionID: *sid,
			})

			gp.StartStop(ctx, &session, connectionData)
		}
	}

	logrus.Debugf("Join session response: sid=%v, event=%s, payload=%s", *sid, msgEvent, payload)

	return sid, &message.Message{
		Event:   msgEvent,
		Payload: payload,
	}
}

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

	sessionData, sErr := gp.Cache.GetSession(ctx, *connectionData.SessionID)
	switch {
	case sErr != nil:
		payload, _ = json.Marshal(sErr.Error())
	case sessionData == nil:
		payload, _ = json.Marshal(message.ErrSessionNotFound.Error())
	default:
		var session model.Session
		if err := json.Unmarshal([]byte(*sessionData), &session); err != nil {
			logrus.Debugf("error occurred while unmarshaling session data: %v", err)
			payload, _ = json.Marshal(err.Error())
			break
		}

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

		uErr := gp.Cache.UpdateSession(ctx, *connectionData.SessionID, &session, expirationTime)
		if uErr == nil {
			msgEvent = string(globalEvent.Disconnected)
			payload, _ = json.Marshal(connectionData)

			gp.StartStop(ctx, &session, connectionData)
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
