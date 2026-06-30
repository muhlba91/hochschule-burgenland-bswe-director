package orchestrator

import (
	"context"
	"encoding/json"

	"github.com/sirupsen/logrus"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/session"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/transport/callback"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/transport/callback/response"
)

// HandleCallback executes the callback handler for a specific event type.
// ctx: The context for managing request lifecycle.
// session: The current game session associated with the callback.
// request: The request data associated with the callback.
// payload: The raw JSON payload of the event.
func (r *Registry) HandleCallback(
	ctx context.Context,
	session session.Session,
	request *callback.Request,
	payload json.RawMessage,
) (any, error) {
	requestHandler, isRegistered := r.callbackHandlers[request.Action]
	if isRegistered {
		return requestHandler(ctx, payload, session, request)
	}

	logrus.Infof("no callback handler registered for action: %s", request.Action)
	return nil, response.ErrNoMatchingRequest
}

// RegisterCallbackHandler registers a new callback handler for a specific action.
// T: The type of the callback payload.
// registry: The registry to register the handler with.
// action: The action for which to register the handler.
// handleAction: The function that will handle the callback.
func RegisterCallbackHandler[T any](
	registry *Registry,
	action string,
	handleAction func(context.Context, *T, session.Session, *callback.Request) (any, error),
) {
	registry.callbackHandlers[action] = func(ctx context.Context, raw json.RawMessage, session session.Session, request *callback.Request) (any, error) {
		var payload T
		if err := json.Unmarshal(raw, &payload); err != nil {
			logrus.Infof("no callback handler registered for action: %s", request.Action)
			return nil, response.ErrInvalidPayload
		}

		return handleAction(ctx, &payload, session, request)
	}
}
