package store

import (
	"context"

	callbackModel "github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/transport/callback"
)

// SessionStore defines the interface for managing sessions in the store.
type SessionStore interface {
	// IsConnected checks if the store is connected and operational.
	IsConnected() bool

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
}

// RequestStore defines the interface for managing requests in the store.
type RequestStore interface {
	// IsConnected checks if the store is connected and operational.
	IsConnected() bool

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
}

// BroadcastStore defines the interface for broadcasting messages to subscribers.
type BroadcastStore interface {
	// IsConnected checks if the store is connected and operational.
	IsConnected() bool

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
