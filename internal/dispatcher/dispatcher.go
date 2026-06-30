package dispatcher

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/cache"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/callback/response"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/registry"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/websocket/connection"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/websocket/message"
)

// Dispatcher is responsible for dispatching flow actions to the appropriate handlers.
type Dispatcher struct {
	Cache    *cache.Cache
	Registry *registry.Registry
}

// NewDispatcher creates a new Dispatcher instance.
// cache: The cache instance for session management and pub/sub.
// registry: The flow registry instance.
func NewDispatcher(cache *cache.Cache, registry *registry.Registry) *Dispatcher {
	return &Dispatcher{
		Cache:    cache,
		Registry: registry,
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
	sid, rMsg := d.Registry.HandleEvent(ctx, msg.Event, connData, msg.Payload)
	if rMsg != nil {
		_ = wsjson.Write(ctx, conn, rMsg)
	}
	return sid
}

// Subscribe subscribes to the Redis pub/sub channel for the given session ID.
// ctx: The context for the subscription.
// sessionID: The ID of the session to subscribe to.
func (d *Dispatcher) Subscribe(ctx context.Context, sessionID string) *redis.PubSub {
	return d.Cache.Subscribe(ctx, fmt.Sprintf("%s:%s", sessionID, cache.BroadcastChannel))
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
	request, rErr := d.Cache.GetRequest(ctx, requestID)
	if rErr != nil || request == nil {
		logrus.Infof("request not found: %s", requestID)
		return echo.NewHTTPError(http.StatusGone, response.NewError(response.ErrNoMatchingRequest))
	}
	if request.Completed {
		logrus.Infof("request already completed: %s", requestID)
		return echo.NewHTTPError(http.StatusGone, response.NewError(response.ErrRequestAlreadyCompleted))
	}

	rMsg, cErr := d.Registry.HandleCallback(ctx, request, body)
	if cErr != nil {
		logrus.Errorf("failed to handle callback for request %s: %v", requestID, cErr)
		return echo.NewHTTPError(http.StatusInternalServerError, response.NewError(cErr))
	}

	request.Completed = true
	uErr := d.Cache.UpdateRequest(ctx, request, cache.DefaultRequestExpiration)
	if uErr != nil {
		logrus.Errorf("failed to update request %s: %v", requestID, uErr)
		return echo.NewHTTPError(http.StatusInternalServerError, response.NewError(response.ErrCompletionFailed))
	}

	// if pending.Blocking {
	// 	current, err := currentBlockingRequestId(ctx, redisClient, sessionId)
	// 	if err != nil || current != requestId {
	// 		// stale callback — a retry already superseded this request
	// 		return echo.NewHTTPError(http.StatusConflict, "callback no longer expected")
	// 	}
	// }

	if rMsg != nil {
		return echoCtx.JSON(http.StatusOK, response.NewSuccess(rMsg))
	}

	return echoCtx.NoContent(http.StatusOK)
}
