package cache

import globalCache "github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/cache"

// sessionExpirationTime defines the duration for which a session remains valid before it expires.
const sessionExpirationTime = globalCache.DefaultSessionExpiration

// requestExpirationTime defines the duration for which a request remains valid before it expires.
const requestExpirationTime = globalCache.DefaultRequestExpiration
