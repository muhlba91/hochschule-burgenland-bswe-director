package cache

// DefaultRequestExpiration defines the default expiration time for cached requests in seconds.
const DefaultRequestExpiration = 1 * 60 * 60

// DefaultSessionExpiration defines the default expiration time for cached sessions in seconds.
const DefaultSessionExpiration = 24 * 60 * 60

// BroadcastChannel defines the channel name used for broadcasting messages in the cache.
const BroadcastChannel = "broadcast"

// maxIterations defines the maximum number of attempts to generate a unique session ID before giving up.
const maxIterations = 5
