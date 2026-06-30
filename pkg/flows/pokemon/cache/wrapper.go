package cache

import globalCache "github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/cache"

// Wrapper is a wrapper for the global cache.
type Wrapper struct {
	cache *globalCache.Cache
}

// NewWrapper creates a new instance of the Wrapper.
// cache: The global cache instance to be wrapped.
func NewWrapper(cache *globalCache.Cache) *Wrapper {
	return &Wrapper{cache: cache}
}
