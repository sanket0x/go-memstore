package memstore

import "time"

// startCleanup runs background cleanup if enabled
func (c *cache) startCleanup() {
	if c.cleanupInterval <= 0 {
		return
	}
	ticker := time.NewTicker(c.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.deleteExpired()
		case <-c.stopChan:
			return
		}
	}
}

// deleteExpired removes expired keys (protected by lock)
func (c *cache) deleteExpired() {
	now := time.Now()
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, v := range c.items {
		if v == nil {
			delete(c.items, k)
			continue
		}
		if !v.expiry.IsZero() && now.After(v.expiry) {
			delete(c.items, k)
		}
	}
}
