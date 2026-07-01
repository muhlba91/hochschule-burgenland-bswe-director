package store

import (
	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/flows/pokemon/store/state"
	globalStore "github.com/muhlba91/hochschule-burgenland-bswe-director/internal/store"
)

// Wrapper is a wrapper for the global cache.
type Wrapper struct {
	sessionStore   globalStore.SessionStore
	requestStore   globalStore.RequestStore
	broadcastStore globalStore.BroadcastStore
	locker         globalStore.Locker
	stateStore     *state.Store
}

// NewWrapper creates a new instance of the Wrapper.
// sessionStore: The session store for managing sessions.
// requestStore: The request store for managing requests.
// broadcastStore: The broadcast store for managing broadcasts.
// locker: The locker for managing locks.
func NewWrapper(
	sessionStore globalStore.SessionStore,
	requestStore globalStore.RequestStore,
	broadcastStore globalStore.BroadcastStore,
	locker globalStore.Locker,
	stateStore *state.Store,
) *Wrapper {
	return &Wrapper{
		sessionStore:   sessionStore,
		requestStore:   requestStore,
		broadcastStore: broadcastStore,
		locker:         locker,
		stateStore:     stateStore,
	}
}
