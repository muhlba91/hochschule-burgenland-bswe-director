package pokemon

import (
	"context"
	"encoding/json"

	"github.com/sirupsen/logrus"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/app/logging"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/flows/pokemon/event"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/transport/websocket/connection"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/transport/websocket/message"
)

// PlayerInformation provides the current information of an existing Pokémon game session.
// ctx: The context for managing request lifecycle.
// _: The event.PlayerInformation struct (not used in this function).
// connectionData: The connection data for the websocket connection.
func (gp *Gameplay) PlayerInformation(
	ctx context.Context,
	_ *event.PlayerInformation,
	connectionData *connection.Data,
) *message.Message {
	state, err := gp.store.GetCurrentStateForPlayerAndSession(
		ctx,
		*connectionData.InternalID,
		*connectionData.SessionID,
	)
	if err != nil {
		payload, _ := json.Marshal(message.ErrSessionNotFound.Error())
		return &message.Message{Event: message.ErrorEvent, Payload: payload}
	}

	playerDetails := &event.PlayerDetails{
		State:      state,
		Connection: connectionData,
	}

	msgEvent := gp.generateEventName(event.TypePlayerDetails)
	payload, _ := json.Marshal(playerDetails)

	logrus.WithFields(logrus.Fields{
		logging.FieldEvent:   msgEvent,
		logging.FieldPayload: string(payload),
	}).Debug("current player information response")

	return &message.Message{
		Event:   msgEvent,
		Payload: payload,
	}
}
