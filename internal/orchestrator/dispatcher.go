package orchestrator

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/app/configuration"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/app/logging"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/store"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/store/constants"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/transport"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/transport/callback/response"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/transport/websocket/connection"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/transport/websocket/message"
)

// Dispatcher is responsible for dispatching flow actions to the appropriate handlers.
type Dispatcher struct {
	sessionStore   store.SessionStore
	requestStore   store.RequestStore
	broadcastStore store.BroadcastStore
	registry       *Registry
	configuration  *configuration.Data
}

// NewDispatcher creates a new Dispatcher instance.
// sessionStore: The store for managing game sessions.
// requestStore: The store for managing requests.
// broadcastStore: The store for managing broadcasts.
// registry: The registry containing the flow handlers.
func NewDispatcher(
	sessionStore store.SessionStore,
	requestStore store.RequestStore,
	broadcastStore store.BroadcastStore,
	registry *Registry,
	configuration *configuration.Data,
) *Dispatcher {
	return &Dispatcher{
		sessionStore:   sessionStore,
		requestStore:   requestStore,
		broadcastStore: broadcastStore,
		registry:       registry,
		configuration:  configuration,
	}
}

// Handle processes incoming WebSocket messages and dispatches them to the appropriate handlers based on the message type.
// ctx: The context for managing request lifecycle.
// connData: The connection data associated with the WebSocket connection.
// msg: The incoming message to be processed.
func (d *Dispatcher) Handle(
	ctx context.Context,
	connData *connection.Data,
	msg message.Message,
) (*string, *message.Message) {
	return d.registry.HandleEvent(ctx, msg.Event, connData, msg.Payload)
}

// Subscribe subscribes to the Redis pub/sub channel for the given session ID.
// ctx: The context for the subscription.
// sessionID: The ID of the session to subscribe to.
func (d *Dispatcher) Subscribe(ctx context.Context, sessionID string) (store.Subscription, error) {
	return d.broadcastStore.Subscribe(ctx, fmt.Sprintf("%s:%s", sessionID, constants.BroadcastChannel))
}

// HandleCallback processes incoming callback messages and dispatches them to the appropriate handlers based on the message type.
// ctx: The context for managing request lifecycle.
// echoCtx: The Echo context for the HTTP request.
// requestID: The request ID associated with the callback.
// body: The raw JSON payload of the callback.
func (d *Dispatcher) HandleCallback(
	ctx context.Context,
	echoCtx echo.Context,
	requestID string,
	body json.RawMessage,
) error {
	request, rErr := d.requestStore.GetRequest(ctx, requestID)
	if rErr != nil || request == nil {
		slog.InfoContext(ctx, "request not found",
			slog.String(logging.FieldRequestID, requestID),
			slog.Any(logging.FieldError, rErr),
		)
		return echo.NewHTTPError(http.StatusGone, response.NewError(response.ErrNoMatchingRequest))
	}

	session, sErr := d.sessionStore.GetBaseSession(ctx, request.SessionID)
	if sErr != nil {
		slog.InfoContext(ctx, "session not found",
			slog.String(logging.FieldSessionID, request.SessionID),
			slog.Any(logging.FieldError, sErr),
		)
		return echo.NewHTTPError(http.StatusGone, response.NewError(transport.ErrNoMatchingSession))
	}

	if request.Parallelization != 0 && !session.IsRequestExpected(requestID) {
		slog.InfoContext(ctx, "request is not expected for session",
			slog.String(logging.FieldRequestID, requestID),
			slog.String(logging.FieldSessionID, request.SessionID),
		)
		return echo.NewHTTPError(http.StatusConflict, response.NewError(response.ErrCallbackNotExpected))
	}

	if d.configuration.CallbackAuthEnabled {
		sigHeader := echoCtx.Request().Header.Get("X-Signature")
		if sigHeader == "" {
			slog.InfoContext(ctx, "missing callback signature",
				slog.String(logging.FieldRequestID, requestID),
			)
			return echo.NewHTTPError(http.StatusUnauthorized, response.NewError(response.ErrCallbackNotAuthorized))
		}

		mac := hmac.New(sha256.New, []byte(request.Secret))
		_, _ = mac.Write(body)
		expected := hex.EncodeToString(mac.Sum(nil))

		if !hmac.Equal([]byte(expected), []byte(sigHeader)) {
			slog.InfoContext(ctx, "invalid callback signature",
				slog.String(logging.FieldRequestID, requestID),
			)
			return echo.NewHTTPError(http.StatusUnauthorized, response.NewError(response.ErrCallbackNotAuthorized))
		}
	}

	cErr := d.registry.HandleCallback(ctx, session, request, body)
	if cErr != nil {
		slog.ErrorContext(ctx, "failed to handle callback",
			slog.String(logging.FieldRequestID, requestID),
			slog.String(logging.FieldSessionID, request.SessionID),
			slog.Any(logging.FieldError, cErr),
		)
		return echo.NewHTTPError(http.StatusInternalServerError, response.NewError(cErr))
	}

	// lock and reload session in case it was updated during callback handling
	unlock, err := d.sessionStore.LockSession(ctx, request.SessionID)
	if err != nil {
		return echo.NewHTTPError(http.StatusLocked, response.NewError(transport.ErrSessionLocked))
	}
	defer unlock(ctx)

	session, sErr = d.sessionStore.GetBaseSession(ctx, request.SessionID)
	if sErr != nil {
		slog.InfoContext(ctx, "session not found during reload",
			slog.String(logging.FieldSessionID, request.SessionID),
			slog.Any(logging.FieldError, sErr),
		)
		return echo.NewHTTPError(http.StatusGone, response.NewError(transport.ErrNoMatchingSession))
	}

	uErr := d.requestStore.CompleteRequest(ctx, session, request)
	if uErr != nil {
		slog.ErrorContext(ctx, "failed to complete request",
			slog.String(logging.FieldRequestID, requestID),
			slog.String(logging.FieldSessionID, request.SessionID),
			slog.Any(logging.FieldError, uErr),
		)
		return echo.NewHTTPError(http.StatusInternalServerError, response.NewError(response.ErrCompletionFailed))
	}

	return echoCtx.NoContent(http.StatusOK)
}
