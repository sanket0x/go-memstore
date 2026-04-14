package memstore

import "time"

func (c *cache[V]) startCleanup() {
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

func (c *cache[V]) deleteExpired() {
	now := time.Now()
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, v := range c.items {
		if v == nil || (!v.expiry.IsZero() && now.After(v.expiry)) {
			delete(c.items, k)
			c.mapSize--
			if c.tracker != nil {
				c.trackerMu.Lock()
				c.tracker.onDelete(k)
				c.trackerMu.Unlock()
			}
		}
	}
}
