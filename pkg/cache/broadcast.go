package cache

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
)

// Broadcast saves a broadcast message in the cache with the specified session ID, channel, and data.
// ctx: The context for the operation.
// sessionID: The unique session ID for the new session.
// channel: The channel to which the broadcast message belongs.
// data: The data to be stored in the broadcast.
func (c *Cache) Broadcast(ctx context.Context, sessionID string, data any) error {
	channelKey := fmt.Sprintf("%s:%s", sessionID, BroadcastChannel)
	payload, _ := json.Marshal(data)

	logrus.Debugf(
		"saving broadcast with ID: %s, channelKey: %s, data: %s",
		sessionID,
		channelKey,
		payload,
	)
	if err := c.client.Publish(ctx, channelKey, payload).Err(); err != nil {
		logrus.Errorf("failed to save broadcast: %v", err)
		return err
	}

	return nil
}
