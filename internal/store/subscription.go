package store

// Subscription defines the interface for managing subscriptions in the store.
type Subscription interface {
	// Channel returns a channel that can be used to receive subscription messages.
	Channel() <-chan string
	// Close closes the subscription and releases any associated resources.
	Close() error
}
