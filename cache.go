package memstore

import (
	"sync"
	"time"
)

// Cache is the public interface for the in-memory store.
type Cache interface {
	// Set stores a key-value pair; ttl <= 0 means no expiry
	Set(key string, value interface{})

	// SetWithDuration stores value with duration (alias)
	SetWithDuration(key string, value interface{}, d time.Duration)

	// Get retrieves a value (value, true) if present and not expired.
	Get(key string) (interface{}, bool)

	// Delete removes a key; returns true if the key existed
	Delete(key string) bool

	// Exists returns true if key exists and not expired
	Exists(key string) bool

	// Keys returns all keys matching pattern (supports '*' wildcard)
	Keys(pattern string) []string

	// Len returns number of non-expired keys
	Len() int

	// Close stops background cleanup goroutine (idempotent)
	Close()
}

// NewCache constructs a Cache using functional options.
// Defaults:
//   - cleanup interval: 1 minute
//
// Example:
//
//	c := NewCache(WithCleanupInterval(0)) // lazy expiry only
func NewCache(opts ...Option) Cache {
	// defaults
	c := &cache{
		items:           make(map[string]*Entry),
		cleanupInterval: time.Minute,
		stopChan:        make(chan struct{}),
	}

	// apply options
	for _, o := range opts {
		o(c)
	}

	// start cleanup goroutine only if enabled
	if c.cleanupInterval > 0 {
		go c.startCleanup()
	}

	return c
}

// concrete implementation
type cache struct {
	mu              sync.RWMutex
	items           map[string]*Entry
	cleanupInterval time.Duration
	stopChan        chan struct{}
	stopOnce        sync.Once
	statsRing       statsRing
	maxKeys         int
	evictionPolicy  EvictionPolicy
}

// Close stops background cleanup (safe to call multiple times).
func (c *cache) Close() {
	c.stopOnce.Do(func() {
		if c.cleanupInterval > 0 && c.stopChan != nil {
			close(c.stopChan)
		}
	})
}
