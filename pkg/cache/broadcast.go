package cache

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
)

// broadcastExpiration defines the expiration time for broadcast messages in seconds.
const broadcastExpiration = 15 * 60

// Broadcast saves a broadcast message in the cache with the specified session ID, channel, and data.
// ctx: The context for the operation.
// sessionID: The unique session ID for the new session.
// channel: The channel to which the broadcast message belongs.
// data: The data to be stored in the broadcast.
func (c *Cache) Broadcast(ctx context.Context, sessionID string, channel Channel, data any) error {
	channelKey := fmt.Sprintf("%s:%s", sessionID, channel)
	payload, _ := json.Marshal(data)

	logrus.Debugf(
		"saving broadcast with ID: %s, channel: %s, data: %s, channelKey: %s",
		sessionID,
		channel,
		payload,
		channelKey,
	)
	if err := c.client.Set(ctx, channelKey, payload, broadcastExpiration).Err(); err != nil {
		logrus.Errorf("failed to save broadcast: %v", err)
		return err
	}

	return nil
}
