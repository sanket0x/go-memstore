package memstore

import "time"

// Option configures the cache.
type Option func(*cache)

// WithCleanupInterval sets the cleanup interval. If interval == 0, background
// cleanup is disabled and expiry is lazy (on access).
func WithCleanupInterval(interval time.Duration) Option {
	return func(c *cache) {
		c.cleanupInterval = interval
	}
}
