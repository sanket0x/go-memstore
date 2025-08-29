package memstore

import "time"

// Set stores a key-value pair with TTL (ttl <= 0 means no expiry)
func (c *cache) Set(key string, value interface{}, ttl time.Duration) {
	var exp time.Time
	if ttl > 0 {
		exp = time.Now().Add(ttl)
	}

	c.mu.Lock()
	c.items[key] = &Entry{
		value:  value,
		expiry: exp,
	}
	c.mu.Unlock()
}

// SetWithDuration stores a value with expiry (alias to Set)
func (c *cache) SetWithDuration(key string, value interface{}, d time.Duration) {
	c.Set(key, value, d)
}

// Get retrieves a value by key
func (c *cache) Get(key string) (interface{}, bool) {
	// fast read lock
	c.mu.RLock()
	entry, exists := c.items[key]
	c.mu.RUnlock()

	if !exists {
		return nil, false
	}

	if entry.isExpired() {
		// lazy delete (double-check under write lock)
		c.mu.Lock()
		if cur, ok := c.items[key]; ok && cur == entry {
			delete(c.items, key)
		}
		c.mu.Unlock()
		return nil, false
	}
	return entry.value, true
}

// Delete removes a key
func (c *cache) Delete(key string) {
	c.mu.Lock()
	delete(c.items, key)
	c.mu.Unlock()
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
