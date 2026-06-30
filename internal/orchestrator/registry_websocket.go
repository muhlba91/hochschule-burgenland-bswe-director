package orchestrator

import (
	"context"
	"encoding/json"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/transport/websocket/connection"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/transport/websocket/message"
)

// HandleEvent executes the event handler for a specific event type.
// ctx: The context for managing request lifecycle.
// eventType: The type of the event to handle.
// connectionData: The data associated with the websocket connection.
// payload: The raw JSON payload of the event.
func (r *Registry) HandleEvent(
	ctx context.Context,
	eventType string,
	connectionData *connection.Data,
	payload json.RawMessage,
) (*string, *message.Message) {
	sessionHandler, isSessionGenerator := r.sessionGeneratorHandlers[eventType]
	if isSessionGenerator {
		return sessionHandler(ctx, payload, connectionData)
	}

	eventHandler, isEventHandler := r.eventHandlers[eventType]
	if isEventHandler {
		msg := eventHandler(ctx, payload, connectionData)
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
	handleAction func(context.Context, *T, *connection.Data) *message.Message,
) {
	registry.eventHandlers[eventType] = func(ctx context.Context, raw json.RawMessage, connectionData *connection.Data) *message.Message {
		var payload T
		if err := json.Unmarshal(raw, &payload); err != nil {
			errPayload, _ := json.Marshal(message.ErrInvalidPayload.Error())
			return &message.Message{
				Event:   message.ErrorEvent,
				Payload: errPayload,
			}
		}

		return handleAction(ctx, &payload, connectionData)
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
	handleAction func(context.Context, *T, *connection.Data) (*string, *message.Message),
) {
	registry.sessionGeneratorHandlers[eventType] = func(ctx context.Context, raw json.RawMessage, connectionData *connection.Data) (*string, *message.Message) {
		var payload T
		if err := json.Unmarshal(raw, &payload); err != nil {
			errPayload, _ := json.Marshal(message.ErrInvalidPayload.Error())
			return nil, &message.Message{
				Event:   message.ErrorEvent,
				Payload: errPayload,
			}
		}

		return handleAction(ctx, &payload, connectionData)
	}
}
