package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/store"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/store/constants"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/session"
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
) *Dispatcher {
	return &Dispatcher{
		sessionStore:   sessionStore,
		requestStore:   requestStore,
		broadcastStore: broadcastStore,
		registry:       registry,
	}
}

// Handle processes incoming WebSocket messages and dispatches them to the appropriate handlers based on the message type.
// ctx: The context for managing request lifecycle.
// conn: The WebSocket connection through which the message was received.
// connData: The connection data associated with the WebSocket connection.
// msg: The incoming message to be processed.
func (d *Dispatcher) Handle(
	ctx context.Context,
	conn *websocket.Conn,
	connData *connection.Data,
	msg message.Message,
) *string {
	sid, rMsg := d.registry.HandleEvent(ctx, msg.Event, connData, msg.Payload)
	if rMsg != nil {
		_ = wsjson.Write(ctx, conn, rMsg)
	}
	return sid
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
		logrus.Infof("request not found: %s", requestID)
		return echo.NewHTTPError(http.StatusGone, response.NewError(response.ErrNoMatchingRequest))
	}
	if request.Completed {
		logrus.Infof("request already completed: %s", requestID)
		return echo.NewHTTPError(http.StatusGone, response.NewError(response.ErrRequestAlreadyCompleted))
	}

	sess, sErr := d.sessionStore.GetSession(ctx, request.SessionID)
	if sErr != nil || sess == nil {
		logrus.Infof("session not found: %s", request.SessionID)
		return echo.NewHTTPError(http.StatusGone, response.NewError(response.ErrNoMatchingSession))
	}
	var session session.Base
	usErr := json.Unmarshal([]byte(*sess), &session)
	if usErr != nil {
		logrus.Errorf("failed to unmarshal session for request %s: %v", requestID, usErr)
		return echo.NewHTTPError(http.StatusInternalServerError, response.NewError(response.ErrNoMatchingSession))
	}

	if request.Parallelization != 0 && !session.IsRequestExpected(requestID) {
		logrus.Infof("request %s is not expected for session %s", requestID, request.SessionID)
		return echo.NewHTTPError(http.StatusConflict, response.NewError(response.ErrCallbackNotExpected))
	}

	rMsg, cErr := d.registry.HandleCallback(ctx, &session, request, body)
	if cErr != nil {
		logrus.Errorf("failed to handle callback for request %s: %v", requestID, cErr)
		return echo.NewHTTPError(http.StatusInternalServerError, response.NewError(cErr))
	}

	request.Completed = true
	uErr := d.requestStore.UpdateRequest(ctx, request)
	if uErr != nil {
		logrus.Errorf("failed to update request %s: %v", requestID, uErr)
		return echo.NewHTTPError(http.StatusInternalServerError, response.NewError(response.ErrCompletionFailed))
	}

	if rMsg != nil {
		return echoCtx.JSON(http.StatusOK, response.NewSuccess(rMsg))
	}

	return echoCtx.NoContent(http.StatusOK)
}
