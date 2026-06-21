package pokemon

import (
	"context"
	"encoding/json"
	"time"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/flows/pokemon/event"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/flows/pokemon/model"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/websocket/message"
)

// expirationTime defines the duration for which a session remains valid before it expires.
const expirationTime = 24 * time.Hour

// Create represents the creation of a new Pokémon game.
// ctx: The context for managing request lifecycle.
// create: The Create struct containing the players for the new game.
func (gp *Gameplay) Create(ctx context.Context, create *model.Create) (*string, *message.Message) {
	var sid *string
	msgEvent := message.ErrorEvent
	payload, _ := json.Marshal(message.ErrSessionNotCreated.Error())

	sid = gp.Cache.GenerateUniqueSessionID(ctx, gp.Name)

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

	return sid, &message.Message{
		Event:   msgEvent,
		Payload: payload,
	}
}

// List represents the listing of existing Pokémon games.
// ctx: The context for managing request lifecycle.
// _ : The List struct (not used in this function).
// _ : The session ID (not used in this function).
func (gp *Gameplay) List(ctx context.Context, _ *model.List, _ string) *message.Message {
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
	return &message.Message{
		Event:   gp.generateEventName(event.Listing),
		Payload: payload,
	}
}

// Join represents the joining of an existing Pokémon game.
// ctx: The context for managing request lifecycle.
// join: The Join struct containing the session ID and player information.
func (gp *Gameplay) Join(ctx context.Context, join *model.Join) (*string, *message.Message) {
	var sid *string
	msgEvent := message.ErrorEvent
	payload, _ := json.Marshal(message.ErrSessionNotJoined.Error())

	return sid, &message.Message{
		Event:   msgEvent,
		Payload: payload,
	}
}
