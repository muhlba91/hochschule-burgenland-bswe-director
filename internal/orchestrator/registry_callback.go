package orchestrator

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/app/logging"
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
) error {
	requestHandler, isRegistered := r.callbackHandlers[request.Action]
	if isRegistered {
		return requestHandler(ctx, payload, session, request)
	}

	slog.InfoContext(ctx, "no callback handler registered",
		slog.String(logging.FieldAction, request.Action),
		slog.String(logging.FieldSessionID, session.GetID()),
		slog.String(logging.FieldRequestID, request.ID),
	)
	return response.ErrNoMatchingRequest
}

// RegisterCallbackHandler registers a new callback handler for a specific action.
// T: The type of the callback payload.
// registry: The registry to register the handler with.
// action: The action for which to register the handler.
// handleAction: The function that will handle the callback.
func RegisterCallbackHandler[T any](
	registry *Registry,
	action string,
	handleAction func(context.Context, *T, session.Session, *callback.Request) error,
) {
	registry.callbackHandlers[action] = func(ctx context.Context, raw json.RawMessage, session session.Session, request *callback.Request) error {
		var payload T
		if err := json.Unmarshal(raw, &payload); err != nil {
			slog.ErrorContext(ctx, "failed to unmarshal callback payload",
				slog.String(logging.FieldAction, request.Action),
				slog.String(logging.FieldSessionID, session.GetID()),
				slog.String(logging.FieldRequestID, request.ID),
				slog.Any(logging.FieldError, err),
			)
			return response.ErrInvalidPayload
		}

		return handleAction(ctx, &payload, session, request)
	}
}
