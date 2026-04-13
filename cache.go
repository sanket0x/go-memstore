package memstore

import (
	"sync"
	"time"
)

// Cache is the public interface for the in-memory store.
type Cache[V any] interface {
	// Set stores a key-value pair with no expiry.
	Set(key string, value V)

	// SetWithDuration stores a value that expires after d.
	SetWithDuration(key string, value V, d time.Duration)

	// Get retrieves a value; returns (value, true) if present and not expired.
	Get(key string) (V, bool)

	// Delete removes a key; returns true if the key existed.
	Delete(key string) bool

	// Exists returns true if the key exists and has not expired.
	Exists(key string) bool

	// Keys returns all live keys matching pattern (supports '*' wildcard).
	Keys(pattern string) []string

	// Len returns the number of live (non-expired) keys.
	Len() int

	// Close stops the background cleanup goroutine (idempotent).
	Close()
}

// cacheConfig holds all configuration that does not depend on the value type V.
// Option functions apply to *cacheConfig, keeping option call sites type-parameter-free.
type cacheConfig struct {
	cleanupInterval time.Duration
	maxKeys         int
	evictionPolicy  EvictionPolicy
	tracker         evictionTracker
}

// NewCache constructs a Cache[V] using functional options.
//
// Example:
//
//	c := memstore.NewCache[string](memstore.WithMaxKeys(1000, memstore.PolicyLRU))
func NewCache[V any](opts ...Option) Cache[V] {
	cfg := cacheConfig{cleanupInterval: time.Minute}
	for _, o := range opts {
		o(&cfg)
	}
	c := &cache[V]{
		items:       make(map[string]*Entry[V]),
		stopChan:    make(chan struct{}),
		cacheConfig: cfg,
	}
	if c.cleanupInterval > 0 {
		go c.startCleanup()
	}
	return c
}

// cache is the concrete implementation of Cache[V].
type cache[V any] struct {
	mu        sync.RWMutex
	items     map[string]*Entry[V]
	stopChan  chan struct{}
	stopOnce  sync.Once
	statsRing statsRing
	cacheConfig
}

// Close stops the background cleanup goroutine (safe to call multiple times).
func (c *cache[V]) Close() {
	c.stopOnce.Do(func() {
		if c.cleanupInterval > 0 && c.stopChan != nil {
			close(c.stopChan)
		}
	})
}
