package memstore

import (
	"sync"
	"time"
)

const (
	statsBucketCount       = 288 // 24h / 5min
	statsBucketGranularity = 5 * time.Minute
	statsWindow            = 24 * time.Hour
)

// Stats holds aggregated cache metrics over a rolling 24-hour window.
type Stats struct {
	Hits      uint64
	Misses    uint64
	Evictions uint64
	Keys      int
}

// StatsProvider is an optional interface implemented by the cache.
// It is not part of the core Cache interface; access it via type assertion:
//
//	if sp, ok := c.(memstore.StatsProvider); ok {
//	    s := sp.Stats()
//	}
type StatsProvider interface {
	Stats() Stats
}

type statsBucket struct {
	hits, misses, evictions uint64
	ts                      time.Time
}

type statsRing struct {
	mu      sync.Mutex
	buckets [statsBucketCount]statsBucket
	current int
}

func (r *statsRing) advance(now time.Time) {
	cur := &r.buckets[r.current]
	if cur.ts.IsZero() {
		cur.ts = now.Truncate(statsBucketGranularity)
		return
	}
	if now.Before(cur.ts.Add(statsBucketGranularity)) {
		return
	}
	r.current = (r.current + 1) % statsBucketCount
	r.buckets[r.current] = statsBucket{ts: now.Truncate(statsBucketGranularity)}
}

func (r *statsRing) recordHit() {
	r.mu.Lock()
	r.advance(time.Now())
	r.buckets[r.current].hits++
	r.mu.Unlock()
}

func (r *statsRing) recordMiss() {
	r.mu.Lock()
	r.advance(time.Now())
	r.buckets[r.current].misses++
	r.mu.Unlock()
}

func (r *statsRing) recordEviction() {
	r.mu.Lock()
	r.advance(time.Now())
	r.buckets[r.current].evictions++
	r.mu.Unlock()
}

func (r *statsRing) snapshot(liveKeys int) Stats {
	cutoff := time.Now().Add(-statsWindow)
	r.mu.Lock()
	defer r.mu.Unlock()

	var s Stats
	for i := range r.buckets {
		b := &r.buckets[i]
		if b.ts.IsZero() || b.ts.Before(cutoff) {
			continue
		}
		s.Hits += b.hits
		s.Misses += b.misses
		s.Evictions += b.evictions
	}
	s.Keys = liveKeys
	return s
}

// Stats returns aggregated metrics for the last 24 hours.
func (c *cache[V]) Stats() Stats {
	return c.statsRing.snapshot(c.Len())
}
