package cache

import (
	"context"
)

// Broadcast saves a broadcast message in the cache with the specified session ID, channel, and data.
// ctx: The context for the operation.
// sessionID: The unique session ID for the new session.
// channel: The channel to which the broadcast message belongs.
// data: The data to be stored in the broadcast.
func (w *Wrapper) Broadcast(ctx context.Context, sessionID string, data any) error {
	return w.cache.Broadcast(ctx, sessionID, data)
}
