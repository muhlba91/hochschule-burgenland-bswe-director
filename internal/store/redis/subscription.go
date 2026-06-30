package redis

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// Subscription wraps go-redis's PubSub object to conform to store.Subscription.
type Subscription struct {
	pubSub *redis.PubSub
	ch     chan string
}

// NewSubscription creates a subscription and starts a background forwarder loop.
// ctx: The context for the subscription.
// pubSub: The go-redis PubSub object to wrap.
func NewSubscription(ctx context.Context, pubSub *redis.PubSub) *Subscription {
	sub := &Subscription{
		pubSub: pubSub,
		ch:     make(chan string),
	}
	go sub.forwardLoop(ctx)

	return sub
}

// forwardLoop reads from go-redis channel and forwards message payloads to the generic channel.
// ctx: The context for the loop.
func (s *Subscription) forwardLoop(ctx context.Context) {
	defer close(s.ch)
	redisCh := s.pubSub.Channel()
	for {
		select {
		case msg, ok := <-redisCh:
			if !ok {
				return
			}
			s.ch <- msg.Payload
		case <-ctx.Done():
			return
		}
	}
}

// Channel satisfies store.Subscription by returning a read-only generic channel.
func (s *Subscription) Channel() <-chan string {
	return s.ch
}

// Close satisfies store.Subscription by closing the underlying connection.
func (s *Subscription) Close() error {
	return s.pubSub.Close()
}
