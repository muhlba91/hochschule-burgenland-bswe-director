package orchestrator

import (
	"context"
	"encoding/json"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/session"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/transport/callback"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/transport/websocket/connection"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/transport/websocket/message"
)

// EventHandlerFunc defines the function signature for handling events.
type EventHandlerFunc func(context.Context, json.RawMessage, *connection.Data) *message.Message

// SessionGeneratorEventHandlerFunc defines the function signature for handling session generation events.
type SessionGeneratorEventHandlerFunc func(context.Context, json.RawMessage, *connection.Data) (*string, *message.Message)

// CallbackHandlerFunc defines the function signature for handling events.
type CallbackHandlerFunc func(context.Context, json.RawMessage, session.Session, *callback.Request) (any, error)

// Registry holds the main logic and handlers.
type Registry struct {
	eventHandlers            map[string]EventHandlerFunc
	sessionGeneratorHandlers map[string]SessionGeneratorEventHandlerFunc
	callbackHandlers         map[string]CallbackHandlerFunc
}

// NewRegistry creates a new registry.
func NewRegistry() *Registry {
	return &Registry{
		eventHandlers:            make(map[string]EventHandlerFunc),
		sessionGeneratorHandlers: make(map[string]SessionGeneratorEventHandlerFunc),
		callbackHandlers:         make(map[string]CallbackHandlerFunc),
	}
}
