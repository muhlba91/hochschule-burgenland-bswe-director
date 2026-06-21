package registry

import (
	"context"
	"encoding/json"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/websocket/message"
)

// EventHandlerFunc defines the function signature for handling events.
type EventHandlerFunc func(context.Context, json.RawMessage, string) *message.Message

// SessionGeneratorEventHandlerFunc defines the function signature for handling session generation events.
type SessionGeneratorEventHandlerFunc func(context.Context, json.RawMessage) (*string, *message.Message)

// Registry holds the main logic and handlers.
type Registry struct {
	eventHandlers            map[string]EventHandlerFunc
	sessionGeneratorHandlers map[string]SessionGeneratorEventHandlerFunc
}

// NewRegistry creates a new registry.
func NewRegistry() *Registry {
	return &Registry{
		eventHandlers:            make(map[string]EventHandlerFunc),
		sessionGeneratorHandlers: make(map[string]SessionGeneratorEventHandlerFunc),
	}
}

// HandleEvent executes the event handler for a specific event type.
// ctx: The context for managing request lifecycle.
// eventType: The type of the event to handle.
// sessionID: The ID of the session handling the event.
// payload: The raw JSON payload of the event.
func (r *Registry) HandleEvent(
	ctx context.Context,
	eventType string,
	sessionID string,
	payload json.RawMessage,
) (*string, *message.Message) {
	sessionHandler, isSessionGenerator := r.sessionGeneratorHandlers[eventType]
	if isSessionGenerator {
		return sessionHandler(ctx, payload)
	}

	eventHandler, isEventHandler := r.eventHandlers[eventType]
	if isEventHandler {
		msg := eventHandler(ctx, payload, sessionID)
		return nil, msg
	}

	errPayload, _ := json.Marshal(message.ErrUnexpectedMessageType.Error())
	return nil, &message.Message{
		Event:   message.ErrorEvent,
		Payload: errPayload,
	}
}

// RegisterEventHandler registers a new event handler for a specific event type.
// T: The type of the event payload.
// registry: The registry to register the handler with.
// eventType: The type of the event to handle.
// handleAction: The function that will handle the event.
func RegisterEventHandler[T any](
	registry *Registry,
	eventType string,
	handleAction func(context.Context, *T, string) *message.Message,
) {
	registry.eventHandlers[eventType] = func(ctx context.Context, raw json.RawMessage, sessionID string) *message.Message {
		var payload T
		if err := json.Unmarshal(raw, &payload); err != nil {
			errPayload, _ := json.Marshal(message.ErrInvalidPayload.Error())
			return &message.Message{
				Event:   message.ErrorEvent,
				Payload: errPayload,
			}
		}

		return handleAction(ctx, &payload, sessionID)
	}
}

// RegisterSessionGeneratorHandler registers a new session generator handler for a specific event type.
// T: The type of the event payload.
// registry: The registry to register the handler with.
// eventType: The type of the event to handle.
// handleAction: The function that will handle the session generation event.
func RegisterSessionGeneratorHandler[T any](
	registry *Registry,
	eventType string,
	handleAction func(context.Context, *T) (*string, *message.Message),
) {
	registry.sessionGeneratorHandlers[eventType] = func(ctx context.Context, raw json.RawMessage) (*string, *message.Message) {
		var payload T
		if err := json.Unmarshal(raw, &payload); err != nil {
			errPayload, _ := json.Marshal(message.ErrInvalidPayload.Error())
			return nil, &message.Message{
				Event:   message.ErrorEvent,
				Payload: errPayload,
			}
		}

		return handleAction(ctx, &payload)
	}
}
