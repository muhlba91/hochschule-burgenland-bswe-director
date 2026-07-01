package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/app/logging"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/store"
	"github.com/muhlba91/hochschule-burgenland-bswe-director/internal/store/constants"
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
	channelKey := fmt.Sprintf("%s:%s", sessionID, constants.BroadcastChannel)
	payload, _ := json.Marshal(data)

	logrus.WithFields(logrus.Fields{
		logging.FieldSessionID: sessionID,
		"channel_key":          channelKey,
		logging.FieldData:      string(payload),
	}).Debug("publishing broadcast")
	if err := c.client.Publish(ctx, channelKey, payload).Err(); err != nil {
		logrus.WithFields(logrus.Fields{
			logging.FieldSessionID: sessionID,
			"channel_key":          channelKey,
			logging.FieldError:     err,
		}).Error("failed to publish broadcast")
		return err
	}

	return nil
}
