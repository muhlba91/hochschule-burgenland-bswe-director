package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/store"
)

// Subscribe subscribes to the given channels and returns a PubSub instance.
// ctx: The context for the operation.
// channels: The channels to subscribe to.
func (c *Cache) Subscribe(ctx context.Context, channels ...string) (store.Subscription, error) {
	pubSub := c.client.Subscribe(ctx, channels...)
	if _, err := pubSub.Receive(ctx); err != nil {
		_ = pubSub.Close()
		return nil, err
	}
	return NewSubscription(ctx, pubSub), nil
}

// Broadcast saves a broadcast message in the cache with the specified session ID, channel, and data.
// ctx: The context for the operation.
// sessionID: The unique session ID for the new session.
// channel: The channel to which the broadcast message belongs.
// data: The data to be stored in the broadcast.
func (c *Cache) Broadcast(ctx context.Context, sessionID string, data any) error {
	channelKey := fmt.Sprintf("%s:%s", sessionID, BroadcastChannel)
	payload, _ := json.Marshal(data)

	logrus.Debugf("publishing broadcast to channelKey: %s, data: %s", channelKey, payload)
	if err := c.client.Publish(ctx, channelKey, payload).Err(); err != nil {
		logrus.Errorf("failed to save broadcast: %v", err)
		return err
	}

	return nil
}
