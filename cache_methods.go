package memstore

import "time"

func (c *cache[V]) liveCountLocked() int {
	n := 0
	for _, v := range c.items {
		if v != nil && !v.isExpired() {
			n++
		}
	}
	return n
}

// enforceCapacity must be called with c.mu held.
func (c *cache[V]) enforceCapacity(key string) bool {
	if c.maxKeys <= 0 {
		return true
	}
	if _, exists := c.items[key]; exists {
		return true
	}
	if c.mapSize < c.maxKeys {
		return true
	}
	if c.tracker == nil {
		return false
	}
	c.trackerMu.Lock()
	evictKey := c.tracker.evict()
	c.trackerMu.Unlock()
	if evictKey != "" {
		delete(c.items, evictKey)
		c.mapSize--
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
	if !isOverwrite {
		c.mapSize++
	}
	if c.tracker != nil {
		c.trackerMu.Lock()
		if isOverwrite {
			c.tracker.onAccess(key)
		} else {
			c.tracker.onInsert(key)
		}
		c.trackerMu.Unlock()
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
	if !isOverwrite {
		c.mapSize++
	}
	if c.tracker != nil {
		c.trackerMu.Lock()
		if isOverwrite {
			c.tracker.onAccess(key)
		} else {
			c.tracker.onInsert(key)
		}
		c.trackerMu.Unlock()
	}
	c.mu.Unlock()
	return nil
}

// Get retrieves a value by key.
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
				c.mapSize--
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
	c.mu.RLock()
	entry, exists := c.items[key]
	c.mu.RUnlock()

	if exists && !entry.isExpired() {
		c.trackerMu.Lock()
		c.tracker.onAccess(key) // no-op if key was concurrently deleted
		c.trackerMu.Unlock()
		c.recordHit()
		return entry.value, true
	}

	if exists && entry.isExpired() {
		c.mu.Lock()
		// Double-check: another goroutine may have already cleaned this up or replaced it.
		if cur, ok := c.items[key]; ok && cur == entry {
			delete(c.items, key)
			c.mapSize--
			c.trackerMu.Lock()
			c.tracker.onDelete(key)
			c.trackerMu.Unlock()
		}
		c.mu.Unlock()
	}

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
		c.mapSize--
		if c.tracker != nil {
			c.trackerMu.Lock()
			c.tracker.onDelete(key)
			c.trackerMu.Unlock()
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
