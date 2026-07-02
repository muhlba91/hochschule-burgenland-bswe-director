package pokemon

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/google/uuid"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/app/logging"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/flows/pokemon/event"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/flows/pokemon/session"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/transport/websocket/connection"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/transport/websocket/message"
)

// Create represents the creation of a new Pokémon game.
// ctx: The context for managing request lifecycle.
// create: The Create struct containing the players for the new game.
// _: The connection data for the websocket connection.
func (gp *Gameplay) Create(ctx context.Context, create *event.Create, _ *connection.Data) (*string, *message.Message) {
	msgEvent := message.ErrorEvent
	payload, _ := json.Marshal(message.ErrSessionNotCreated.Error())

	sessionID := gp.store.GenerateUniqueSessionID()
	create.PlayerA.ID = uuid.NewString()
	create.PlayerB.ID = uuid.NewString()
	session := &session.Session{
		Players: map[string]*session.Player{
			create.PlayerA.ID: &create.PlayerA,
			create.PlayerB.ID: &create.PlayerB,
		},
	}
	session.ID = sessionID

	err := gp.store.CreateSession(ctx, session)
	if err == nil {
		msgEvent = gp.generateEventName(event.TypeCreated)
		payload, _ = json.Marshal(&event.Created{
			SessionID: sessionID,
		})
	}

	slog.DebugContext(ctx, "create session response",
		slog.String(logging.FieldSessionID, sessionID),
		slog.String(logging.FieldEvent, msgEvent),
		slog.String(logging.FieldPayload, string(payload)),
	)

	return nil, &message.Message{
		Event:   msgEvent,
		Payload: payload,
	}
}

// List represents the listing of existing Pokémon games.
// ctx: The context for managing request lifecycle.
// _ : The List struct (not used in this function).
// _ : The connection.Data struct (not used in this function).
func (gp *Gameplay) List(ctx context.Context, _ *event.List, _ *connection.Data) *message.Message {
	sessions := gp.store.ListSessions(ctx)

	payload, _ := json.Marshal(&event.Listing{
		Sessions: sessions,
	})

	slog.DebugContext(ctx, "list sessions response",
		slog.String(logging.FieldPayload, string(payload)),
	)

	return &message.Message{
		Event:   gp.generateEventName(event.TypeListing),
		Payload: payload,
	}
}

// Join represents the joining of an existing Pokémon game.
// ctx: The context for managing request lifecycle.
// join: The Join struct containing the session ID and player information.
// connectionData: The connection data for the websocket connection.
func (gp *Gameplay) Join(
	ctx context.Context,
	join *event.Join,
	connectionData *connection.Data,
) (*string, *message.Message) {
	sid := &join.SessionID

	unlock, err := gp.store.LockSession(ctx, *sid)
	if err != nil {
		payload, _ := json.Marshal(message.ErrSessionNotJoined.Error())
		return nil, &message.Message{Event: message.ErrorEvent, Payload: payload}
	}
	defer unlock(ctx)

	msgEvent := message.ErrorEvent
	payload, _ := json.Marshal(message.ErrSessionNotJoined.Error())

	session, sErr := gp.store.GetSession(ctx, *sid)
	switch {
	case sErr != nil:
		payload, _ = json.Marshal(sErr.Error())
	case session == nil:
		payload, _ = json.Marshal(message.ErrSessionNotFound.Error())
	default:
		connected := false
		if player, ok := session.Players[join.Player]; ok && !player.Connected {
			slog.DebugContext(ctx, "player is joining the session",
				slog.String(logging.FieldPlayerID, player.ID),
				slog.String(logging.FieldSessionID, *sid),
				slog.String(logging.FieldConnectionID, connectionData.ConnectionID),
			)
			player.Connected = true
			player.ConnectionID = &connectionData.ConnectionID
			connectionData.InternalID = &player.ID
			connected = true
		}

		if !connected {
			slog.DebugContext(ctx, "player is not a valid player for session",
				slog.String(logging.FieldPlayerID, join.Player),
				slog.String(logging.FieldSessionID, *sid),
			)
			break
		}

		uErr := gp.store.UpdateSession(ctx, session)
		if uErr == nil {
			msgEvent = gp.generateEventName(event.TypeJoined)
			payload, _ = json.Marshal(&event.Joined{
				SessionID: *sid,
			})

			go gp.StartStop(context.WithoutCancel(ctx), session, connectionData)
		}
	}

	slog.DebugContext(ctx, "join session response",
		slog.String(logging.FieldSessionID, *sid),
		slog.String(logging.FieldEvent, msgEvent),
		slog.String(logging.FieldPayload, string(payload)),
	)

	return sid, &message.Message{
		Event:   msgEvent,
		Payload: payload,
	}
}
