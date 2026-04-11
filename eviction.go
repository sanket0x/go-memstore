package memstore

// EvictionPolicy controls what happens when the cache reaches its MaxKeys limit.
type EvictionPolicy int

const (
	// PolicyNone silently rejects new keys once MaxKeys is reached.
	// Overwrites of existing keys are always allowed.
	PolicyNone EvictionPolicy = iota

	// PolicyLRU evicts the least-recently-used key to make room.
	// Not yet implemented — reserved for v1.1.
	PolicyLRU

	// PolicyLFU evicts the least-frequently-used key to make room.
	// Not yet implemented — reserved for v1.1.
	PolicyLFU
)

// WithMaxKeys sets the maximum number of keys the cache will hold and the
// eviction policy to apply when the limit is reached.
//
// Example — reject new keys once 1 000 are stored:
//
//	c := memstore.NewCache(memstore.WithMaxKeys(1000, memstore.PolicyNone))
func WithMaxKeys(n int, policy EvictionPolicy) Option {
	return func(c *cache) {
		c.maxKeys = n
		c.evictionPolicy = policy
	}
}
