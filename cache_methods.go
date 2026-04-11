package memstore

import "time"

// atCapacity reports whether the cache is at its key limit for a new (non-overwrite) insertion.
// Must be called with c.mu held.
func (c *cache) atCapacity(key string) bool {
	if c.maxKeys <= 0 {
		return false
	}
	if _, exists := c.items[key]; exists {
		return false // overwrite is always allowed
	}
	// Count non-expired live keys
	live := 0
	for _, v := range c.items {
		if v != nil && !v.isExpired() {
			live++
		}
	}
	return live >= c.maxKeys
}

// Set stores a key-value pair without expiry
func (c *cache) Set(key string, value interface{}) {
	c.mu.Lock()
	if c.atCapacity(key) {
		c.mu.Unlock()
		c.statsRing.recordEviction()
		return
	}
	c.items[key] = &Entry{value: value}
	c.mu.Unlock()
}

// SetWithDuration stores a value with expiry
func (c *cache) SetWithDuration(key string, value interface{}, d time.Duration) {
	c.mu.Lock()
	if c.atCapacity(key) {
		c.mu.Unlock()
		c.statsRing.recordEviction()
		return
	}
	c.items[key] = &Entry{
		value:  value,
		expiry: time.Now().Add(d),
	}
	c.mu.Unlock()
}

// Get retrieves a value by key
func (c *cache) Get(key string) (interface{}, bool) {
	// fast read lock
	c.mu.RLock()
	entry, exists := c.items[key]
	c.mu.RUnlock()

	if !exists {
		c.statsRing.recordMiss()
		return nil, false
	}

	if entry.isExpired() {
		if c.cleanupInterval == 0 {
			// lazy delete only when no background cleanup
			c.mu.Lock()
			if cur, ok := c.items[key]; ok && cur == entry {
				delete(c.items, key)
			}
			c.mu.Unlock()
		}
		c.statsRing.recordMiss()
		return nil, false
	}
	c.statsRing.recordHit()
	return entry.value, true
}

// Delete removes a key; returns true if it existed
func (c *cache) Delete(key string) bool {
	c.mu.Lock()
	_, exists := c.items[key]
	delete(c.items, key)
	c.mu.Unlock()
	return exists
}

// Exists checks if a key exists (and not expired)
func (c *cache) Exists(key string) bool {
	_, ok := c.Get(key)
	return ok
}

// Keys returns keys matching a simple pattern (supports '*' wildcard anywhere)
func (c *cache) Keys(pattern string) []string {
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

// Len returns number of non-expired items
func (c *cache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	count := 0
	for _, v := range c.items {
		if v != nil && !v.isExpired() {
			count++
		}
	}
	return count
}
