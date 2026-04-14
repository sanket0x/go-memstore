package memstore

import (
	"errors"
	"sync"
	"time"
)

// ErrCacheFull is returned by Set and SetWithDuration when the cache has reached
// its MaxKeys limit and the eviction policy is PolicyNone.
var ErrCacheFull = errors.New("cache is full")

// Cache is an in-memory key-value store.
type Cache[V any] interface {
	// Set stores a key-value pair with no expiry.
	// Returns ErrCacheFull if the cache is at capacity and the policy is PolicyNone.
	Set(key string, value V) error

	// SetWithDuration stores a value that expires after d.
	// Returns ErrCacheFull if the cache is at capacity and the policy is PolicyNone.
	SetWithDuration(key string, value V, d time.Duration) error

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

type cacheConfig struct {
	cleanupInterval time.Duration
	maxKeys         int
	evictionPolicy  EvictionPolicy
	tracker         evictionTracker
	statsRing       *statsRing // nil when stats are disabled (default)
}

// NewCache returns a new Cache[V].
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

type cache[V any] struct {
	mu        sync.RWMutex  // protects items and mapSize
	trackerMu sync.Mutex    // protects tracker; always acquired after mu, never before
	items     map[string]*Entry[V]
	mapSize   int // total entries in items; maintained on every insert/delete
	stopChan  chan struct{}
	stopOnce  sync.Once
	cacheConfig
}

func (c *cache[V]) Close() {
	c.stopOnce.Do(func() {
		if c.cleanupInterval > 0 && c.stopChan != nil {
			close(c.stopChan)
		}
	})
}
