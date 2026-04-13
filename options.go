package memstore

import "time"

// Option configures the cache.
type Option func(*cacheConfig)

// WithCleanupInterval sets the background cleanup interval.
// If interval == 0, background cleanup is disabled and expiry is lazy (on access).
func WithCleanupInterval(interval time.Duration) Option {
	return func(cfg *cacheConfig) {
		cfg.cleanupInterval = interval
	}
}
