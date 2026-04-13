package memstore

import "time"

// liveCountLocked counts non-expired keys. Must be called with c.mu held.
func (c *cache[V]) liveCountLocked() int {
	n := 0
	for _, v := range c.items {
		if v != nil && !v.isExpired() {
			n++
		}
	}
	return n
}

// enforceCapacity makes room for a new key when the cache is full.
// Returns false if the insert should be rejected (PolicyNone). Must be called with c.mu held.
func (c *cache[V]) enforceCapacity(key string) bool {
	if c.maxKeys <= 0 {
		return true
	}
	if _, exists := c.items[key]; exists {
		return true // overwrite always allowed
	}
	if c.liveCountLocked() < c.maxKeys {
		return true
	}
	if c.tracker == nil {
		return false // PolicyNone: reject
	}
	// LRU / LFU: evict one key to make room
	if evictKey := c.tracker.evict(); evictKey != "" {
		delete(c.items, evictKey)
		c.recordEviction()
	}
	return true
}

func (c *cache[V]) recordHit() {
	if c.statsRing != nil {
		c.statsRing.recordHit()
	}
}

func (c *cache[V]) recordMiss() {
	if c.statsRing != nil {
		c.statsRing.recordMiss()
	}
}

func (c *cache[V]) recordEviction() {
	if c.statsRing != nil {
		c.statsRing.recordEviction()
	}
}

// Set stores a key-value pair without expiry.
// Returns ErrCacheFull if the cache is at capacity and the policy is PolicyNone.
func (c *cache[V]) Set(key string, value V) error {
	c.mu.Lock()
	if !c.enforceCapacity(key) {
		c.mu.Unlock()
		c.recordEviction()
		return ErrCacheFull
	}
	_, isOverwrite := c.items[key]
	c.items[key] = &Entry[V]{value: value}
	if c.tracker != nil {
		if isOverwrite {
			c.tracker.onAccess(key)
		} else {
			c.tracker.onInsert(key)
		}
	}
	c.mu.Unlock()
	return nil
}

// SetWithDuration stores a value that expires after d.
// Returns ErrCacheFull if the cache is at capacity and the policy is PolicyNone.
func (c *cache[V]) SetWithDuration(key string, value V, d time.Duration) error {
	c.mu.Lock()
	if !c.enforceCapacity(key) {
		c.mu.Unlock()
		c.recordEviction()
		return ErrCacheFull
	}
	_, isOverwrite := c.items[key]
	c.items[key] = &Entry[V]{value: value, expiry: time.Now().Add(d)}
	if c.tracker != nil {
		if isOverwrite {
			c.tracker.onAccess(key)
		} else {
			c.tracker.onInsert(key)
		}
	}
	c.mu.Unlock()
	return nil
}

// Get retrieves a value by key.
// With LRU/LFU active, a write lock is used to update access order atomically.
func (c *cache[V]) Get(key string) (V, bool) {
	if c.tracker != nil {
		return c.getTracked(key)
	}

	c.mu.RLock()
	entry, exists := c.items[key]
	c.mu.RUnlock()

	if !exists {
		c.recordMiss()
		var zero V
		return zero, false
	}
	if entry.isExpired() {
		if c.cleanupInterval == 0 {
			c.mu.Lock()
			if cur, ok := c.items[key]; ok && cur == entry {
				delete(c.items, key)
			}
			c.mu.Unlock()
		}
		c.recordMiss()
		var zero V
		return zero, false
	}
	c.recordHit()
	return entry.value, true
}

func (c *cache[V]) getTracked(key string) (V, bool) {
	c.mu.Lock()
	entry, exists := c.items[key]
	if exists && !entry.isExpired() {
		c.tracker.onAccess(key)
		c.mu.Unlock()
		c.recordHit()
		return entry.value, true
	}
	if exists && entry.isExpired() {
		delete(c.items, key)
		c.tracker.onDelete(key)
	}
	c.mu.Unlock()
	c.recordMiss()
	var zero V
	return zero, false
}

// Delete removes a key; returns true if it existed.
func (c *cache[V]) Delete(key string) bool {
	c.mu.Lock()
	_, exists := c.items[key]
	if exists {
		delete(c.items, key)
		if c.tracker != nil {
			c.tracker.onDelete(key)
		}
	}
	c.mu.Unlock()
	return exists
}

// Exists checks if a key exists and has not expired.
func (c *cache[V]) Exists(key string) bool {
	_, ok := c.Get(key)
	return ok
}

// Keys returns all live keys matching pattern (supports '*' wildcard).
func (c *cache[V]) Keys(pattern string) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var keys []string
	for k, v := range c.items {
		if v == nil || v.isExpired() {
			continue
		}
		if matchPattern(k, pattern) {
			keys = append(keys, k)
		}
	}
	return keys
}

// Len returns the number of live (non-expired) keys.
func (c *cache[V]) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.liveCountLocked()
}
