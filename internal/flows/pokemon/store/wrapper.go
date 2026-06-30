package store

import globalStore "github.com/muhlba91/hochschule-burgenland-bswe-director/internal/store"

// Wrapper is a wrapper for the global cache.
type Wrapper struct {
	sessionStore   globalStore.SessionStore
	requestStore   globalStore.RequestStore
	broadcastStore globalStore.BroadcastStore
}

// NewWrapper creates a new instance of the Wrapper.
// sessionStore: The session store for managing sessions.
// requestStore: The request store for managing requests.
// broadcastStore: The broadcast store for managing broadcasts.
func NewWrapper(
	sessionStore globalStore.SessionStore,
	requestStore globalStore.RequestStore,
	broadcastStore globalStore.BroadcastStore,
) *Wrapper {
	return &Wrapper{
		sessionStore:   sessionStore,
		requestStore:   requestStore,
		broadcastStore: broadcastStore,
	}
}
