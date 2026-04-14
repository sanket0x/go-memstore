package memstore

// EvictionPolicy controls what happens when the cache reaches its MaxKeys limit.
type EvictionPolicy int

const (
	// PolicyNone rejects new keys when the cache is full; Set returns ErrCacheFull.
	// Overwrites of existing keys are always allowed.
	PolicyNone EvictionPolicy = iota

	// PolicyLRU evicts the least-recently-used key to make room.
	PolicyLRU

	// PolicyLFU evicts the least-frequently-used key to make room.
	PolicyLFU
)

// evictionTracker is the internal interface for LRU/LFU bookkeeping.
// All methods must be called with cache.trackerMu held.
type evictionTracker interface {
	onInsert(key string)
	onAccess(key string)
	onDelete(key string)
	evict() string
}

// WithMaxKeys returns an Option that caps the cache at n keys, applying policy when the limit is reached.
//
//	c := memstore.NewCache[string](memstore.WithMaxKeys(1000, memstore.PolicyLRU))
func WithMaxKeys(n int, policy EvictionPolicy) Option {
	return func(cfg *cacheConfig) {
		cfg.maxKeys = n
		cfg.evictionPolicy = policy
		switch policy {
		case PolicyLRU:
			cfg.tracker = newLRUTracker()
		case PolicyLFU:
			cfg.tracker = newLFUTracker()
		}
	}
}
