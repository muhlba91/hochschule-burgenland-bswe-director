package store

import (
	"context"
	"time"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/session"
	callbackModel "github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/transport/callback"
)

// Store defines the interface for a generic store that can be used for various purposes, such as session management, request handling, and broadcasting messages.
type Store interface {
	// IsConnected checks if the store is connected and operational.
	IsConnected() bool

	// Exists checks if the given key exists in the store.
	// ctx: The context for the operation.
	// key: The key to check for existence.
	Exists(ctx context.Context, key string) (bool, error)

	// Get retrieves the value associated with the given key from the store.
	// ctx: The context for the operation.
	// key: The key to retrieve the value for.
	Get(ctx context.Context, key string) (*string, error)

	// Set stores the given value in the store with the specified key and expiration duration.
	// ctx: The context for the operation.
	// key: The key to store the value under.
	// value: The value to be stored in the store.
	// expiration: The duration after which the key-value pair should expire.
	Set(ctx context.Context, key string, value any, expiration time.Duration) error

	// Delete removes the value associated with the given key from the store.
	// ctx: The context for the operation.
	// key: The key to be deleted from the store.
	Delete(ctx context.Context, key string) error
}

// SessionStore defines the interface for managing sessions in the store.
type SessionStore interface {
	Store

	// CreateSession creates a new session with the given session ID, data, and expiration duration.
	// ctx: The context for the operation.
	// sessionID: The unique identifier for the session.
	// data: The data to be stored in the session.
	CreateSession(ctx context.Context, sessionID string, data any) error

	// UpdateSession updates an existing session with the given session ID, data, and expiration duration.
	// ctx: The context for the operation.
	// sessionID: The unique identifier for the session.
	// data: The new data to be stored in the session.
	UpdateSession(ctx context.Context, sessionID string, data any) error

	// GetSession retrieves the data associated with a given session ID.
	// ctx: The context for the operation.
	// sessionID: The unique identifier for the session.
	GetSession(ctx context.Context, sessionID string) (*string, error)

	// ListSessions lists all active sessions for a given flow name.
	// ctx: The context for the operation.
	// flowName: The name of the flow for which to list sessions.
	ListSessions(ctx context.Context, flowName string) map[string]string

	// Broadcast broadcasts data to all active sessions for a given flow name.
	// ctx: The context for the operation.
	// flowName: The name of the flow for which to broadcast data.
	// data: The data to be broadcasted to all active sessions.
	Broadcast(ctx context.Context, sessionID string, data any) error

	// LockSession locks a session for exclusive access, preventing other operations from modifying it.
	// ctx: The context for the operation.
	// sessionID: The unique identifier for the session to be locked.
	LockSession(ctx context.Context, sessionID string) (UnlockFunc, error)
}

// RequestStore defines the interface for managing requests in the store.
type RequestStore interface {
	Store

	// CreateRequest creates a new request with the given request data and expiration duration.
	// ctx: The context for the operation.
	// request: The request data to be stored.
	CreateRequest(ctx context.Context, request *callbackModel.Request) error

	// UpdateRequest updates an existing request with the given request data and expiration duration.
	// ctx: The context for the operation.
	// request: The new request data to be stored.
	UpdateRequest(ctx context.Context, request *callbackModel.Request) error

	// GetRequest retrieves the request data associated with a given request ID.
	// ctx: The context for the operation.
	// requestID: The unique identifier for the request.
	GetRequest(ctx context.Context, requestID string) (*callbackModel.Request, error)

	// CompleteRequest marks a request as completed and performs any necessary cleanup or finalization.
	// ctx: The context for the operation.
	// session: The current game session containing player connection information.
	// request: The request to be marked as completed.
	CompleteRequest(ctx context.Context, session session.Session, request *callbackModel.Request) error
}

// BroadcastStore defines the interface for broadcasting messages to subscribers.
type BroadcastStore interface {
	Store

	// Broadcast broadcasts data to all subscribers of a given session ID.
	// ctx: The context for the operation.
	// sessionID: The unique identifier for the session to which the data will be broadcasted.
	// data: The data to be broadcasted to all subscribers.
	Broadcast(ctx context.Context, sessionID string, data any) error

	// Subscribe subscribes to the given channels and returns a PubSub instance for receiving messages.
	// ctx: The context for the operation.
	// channels: The channels to which to subscribe.
	Subscribe(ctx context.Context, channels ...string) (Subscription, error)
}
