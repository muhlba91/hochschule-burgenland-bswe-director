package websocket

import (
	"context"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/redis/go-redis/v9"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/cache"
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
	return d.Cache.Subscribe(ctx, sessionID)
}
