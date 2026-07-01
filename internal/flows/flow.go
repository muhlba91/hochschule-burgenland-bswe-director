package flows

import (
	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/flows/pokemon"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/orchestrator"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/store"
)

// Flow defines the interface that all flows must implement.
type Flow interface {
	// Name returns the unique, snake_case identifier of the flow (e.g., "pokemon", "chess", "order_approval").
	Name() string

	// Init is called during startup to let the flow register its event and callback handlers.
	// reg: The flow registry to register the flow.
	Init(reg *orchestrator.Registry)
}

// Init initializes the flows.
// sessionStore: The session store for managing sessions.
// requestStore: The request store for managing requests.
// broadcastStore: The broadcast store for managing broadcasts.
// locker: The locker for managing locks.
// requestor: The requestor instance for the flows.
// store: The global store instance for the flows.
// registry: The registry to register the flows.
func Init(
	sessionStore store.SessionStore,
	requestStore store.RequestStore,
	broadcastStore store.BroadcastStore,
	locker store.Locker,
	requestor *orchestrator.Requestor,
	store store.Store,
	registry *orchestrator.Registry,
) {
	activeFlows := []Flow{
		pokemon.NewGameplay(sessionStore, requestStore, broadcastStore, locker, requestor, store),
	}

	for _, flow := range activeFlows {
		flow.Init(registry)
	}
}
