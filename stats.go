package memstore

import (
	"sync"
	"time"
)

const (
	statsBucketGranularity = 5 * time.Minute
	statsWindow            = 24 * time.Hour
	statsBucketCount       = int(statsWindow / statsBucketGranularity)
)

// Stats holds aggregated cache metrics over a rolling 24-hour window.
type Stats struct {
	Hits      uint64
	Misses    uint64
	Evictions uint64
}

// StatsHandle provides access to rolling stats for a cache created with WithStats.
type StatsHandle struct {
	ring *statsRing
}

// Snapshot returns the aggregated metrics for the last 24 hours.
func (h *StatsHandle) Snapshot() Stats {
	return h.ring.snapshot()
}

// WithStats enables 24-hour rolling stats collection and returns the option
// alongside a handle for reading snapshots. Stats are disabled by default.
//
//	statsOpt, stats := memstore.WithStats()
//	c := memstore.NewCache[string](statsOpt)
//	s := stats.Snapshot()
func WithStats() (Option, *StatsHandle) {
	ring := &statsRing{}
	handle := &StatsHandle{ring: ring}
	opt := func(cfg *cacheConfig) {
		cfg.statsRing = ring
	}
	return opt, handle
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

func (r *statsRing) snapshot() Stats {
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
	return s
}
