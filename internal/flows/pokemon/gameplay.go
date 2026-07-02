package pokemon

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/flows/pokemon/constants"
	pokemonStore "github.com/muhlba91/hochschule-burgenland-bswe-director/internal/flows/pokemon/store"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/flows/pokemon/store/state"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/orchestrator"
	globalStore "github.com/muhlba91/hochschule-burgenland-bswe-director/internal/store"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/flows/pokemon/action"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/flows/pokemon/event"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/transport/websocket/message"
)

// Gameplay represents a gameplay type.
type Gameplay struct {
	store     *pokemonStore.Wrapper
	requestor *orchestrator.Requestor
}

// NewGameplay creates a new instance of the Pokémon gameplay.
// cache: The cache instance for the gameplay.
// registry: The flow registry to register the gameplay.
// sessionStore: The session store for managing sessions.
// requestStore: The request store for managing requests.
// broadcastStore: The broadcast store for managing broadcasts.
// locker: The locker for managing locks.
// requestor: The requestor for handling requests.
func NewGameplay(
	sessionStore globalStore.SessionStore,
	requestStore globalStore.RequestStore,
	broadcastStore globalStore.BroadcastStore,
	locker globalStore.Locker,
	requestor *orchestrator.Requestor,
	store globalStore.Store,
) *Gameplay {
	stateStore := state.NewStore(store)
	cacheWrapper := pokemonStore.NewWrapper(sessionStore, requestStore, broadcastStore, locker, stateStore)

	return &Gameplay{
		store:     cacheWrapper,
		requestor: requestor,
	}
}

// Name returns the unique identifier of the Pokémon gameplay.
func (gp *Gameplay) Name() string {
	return constants.Name
}

// generateEventName generates the event name for a specific event type.
// event: The event type for which to generate the name.
func (gp *Gameplay) generateEventName(event event.Type) string {
	return generateName(string(event))
}

// generateActionName generates the action name for a specific action type.
// action: The action type for which to generate the name.
func (gp *Gameplay) generateActionName(action action.Type) string {
	return generateName(string(action))
}

// generateName generates a unique name by combining the gameplay name and the provided name.
// name: The name to be combined with the gameplay name.
func generateName(name string) string {
	return fmt.Sprintf("%s:%s", constants.Name, name)
}

// reportBackgroundError broadcasts a user-friendly error to all connected clients.
// ctx: The context for managing request-scoped values, cancellation signals, and deadlines.
// sessionID: The ID of the session where the error occurred.
// clientErr: The safe, user-friendly error to send to clients.
func (gp *Gameplay) reportBackgroundError(ctx context.Context, sessionID string, clientErr error) {
	payload, _ := json.Marshal(clientErr.Error())
	_ = gp.store.Broadcast(ctx, sessionID, &message.Message{
		Event:   message.ErrorEvent,
		Payload: payload,
	})
}
