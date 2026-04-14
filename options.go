package memstore

import "time"

// Option is a functional option for NewCache.
type Option func(*cacheConfig)

// WithCleanupInterval returns an Option that sets the background cleanup interval.
// A value of 0 disables background cleanup; expired entries are removed on the next access instead.
func WithCleanupInterval(interval time.Duration) Option {
	return func(cfg *cacheConfig) {
		cfg.cleanupInterval = interval
	}
}
