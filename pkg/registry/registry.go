package registry

import (
	"context"
	"encoding/json"

	"github.com/sirupsen/logrus"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/callback/model"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/callback/response"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/websocket/connection"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/websocket/message"
)

// EventHandlerFunc defines the function signature for handling events.
type EventHandlerFunc func(context.Context, json.RawMessage, *connection.Data) *message.Message

// SessionGeneratorEventHandlerFunc defines the function signature for handling session generation events.
type SessionGeneratorEventHandlerFunc func(context.Context, json.RawMessage, *connection.Data) (*string, *message.Message)

// CallbackHandlerFunc defines the function signature for handling events.
type CallbackHandlerFunc func(context.Context, *model.Request, json.RawMessage) (any, error)

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

// HandleCallback executes the callback handler for a specific event type.
// ctx: The context for managing request lifecycle.
// request: The request data associated with the callback.
// payload: The raw JSON payload of the event.
func (r *Registry) HandleCallback(
	ctx context.Context,
	request *model.Request,
	payload json.RawMessage,
) (any, error) {
	requestHandler, isRegistered := r.callbackHandlers[request.Action]
	if isRegistered {
		return requestHandler(ctx, request, payload)
	}

	logrus.Infof("no callback handler registered for action: %s", request.Action)
	return nil, response.ErrNoMatchingRequest
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

// RegisterCallbackHandler registers a new callback handler for a specific action.
// T: The type of the callback payload.
// registry: The registry to register the handler with.
// action: The action for which to register the handler.
// handleAction: The function that will handle the callback.
func RegisterCallbackHandler[T any](
	registry *Registry,
	action string,
	handleAction func(context.Context, *model.Request, *T) (any, error),
) {
	registry.callbackHandlers[action] = func(ctx context.Context, request *model.Request, raw json.RawMessage) (any, error) {
		var payload T
		if err := json.Unmarshal(raw, &payload); err != nil {
			logrus.Infof("no callback handler registered for action: %s", request.Action)
			return nil, response.ErrInvalidPayload
		}

		return handleAction(ctx, request, &payload)
	}
}
