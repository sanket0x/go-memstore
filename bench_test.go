package memstore

import (
	"fmt"
	"testing"
	"time"
)

// ── Default (no eviction policy) ─────────────────────────────────────────────

func BenchmarkSet(b *testing.B) {
	c := NewCache[any](WithCleanupInterval(0))
	defer c.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Set(fmt.Sprintf("key%d", i), i)
	}
}

func BenchmarkConcurrentSet(b *testing.B) {
	c := NewCache[any](WithCleanupInterval(0))
	defer c.Close()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			c.Set(fmt.Sprintf("key%d", i), i)
			i++
		}
	})
}

func BenchmarkSetWithDuration(b *testing.B) {
	c := NewCache[any](WithCleanupInterval(0))
	defer c.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.SetWithDuration(fmt.Sprintf("key%d", i), i, time.Minute)
	}
}

func BenchmarkConcurrentSetWithDuration(b *testing.B) {
	c := NewCache[any](WithCleanupInterval(0))
	defer c.Close()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			c.SetWithDuration(fmt.Sprintf("key%d", i), i, time.Minute)
			i++
		}
	})
}

func BenchmarkGet(b *testing.B) {
	c := NewCache[any](WithCleanupInterval(0))
	defer c.Close()
	c.Set("key", "value")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Get("key")
	}
}

func BenchmarkConcurrentGet(b *testing.B) {
	c := NewCache[any](WithCleanupInterval(0))
	defer c.Close()
	for i := 0; i < 1000; i++ {
		c.Set(fmt.Sprintf("key%d", i), i)
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			c.Get(fmt.Sprintf("key%d", i%1000))
			i++
		}
	})
}

// ── LRU eviction ─────────────────────────────────────────────────────────────

const lruCap = 1000

func BenchmarkSetLRU(b *testing.B) {
	c := NewCache[any](WithCleanupInterval(0), WithMaxKeys(lruCap, PolicyLRU))
	defer c.Close()
	for i := 0; i < lruCap; i++ {
		c.Set(fmt.Sprintf("seed%d", i), i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Set(fmt.Sprintf("key%d", i), i)
	}
}

func BenchmarkConcurrentSetLRU(b *testing.B) {
	c := NewCache[any](WithCleanupInterval(0), WithMaxKeys(lruCap, PolicyLRU))
	defer c.Close()
	for i := 0; i < lruCap; i++ {
		c.Set(fmt.Sprintf("seed%d", i), i)
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			c.Set(fmt.Sprintf("key%d", i), i)
			i++
		}
	})
}

func BenchmarkSetWithDurationLRU(b *testing.B) {
	c := NewCache[any](WithCleanupInterval(0), WithMaxKeys(lruCap, PolicyLRU))
	defer c.Close()
	for i := 0; i < lruCap; i++ {
		c.Set(fmt.Sprintf("seed%d", i), i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.SetWithDuration(fmt.Sprintf("key%d", i), i, time.Minute)
	}
}

func BenchmarkConcurrentSetWithDurationLRU(b *testing.B) {
	c := NewCache[any](WithCleanupInterval(0), WithMaxKeys(lruCap, PolicyLRU))
	defer c.Close()
	for i := 0; i < lruCap; i++ {
		c.Set(fmt.Sprintf("seed%d", i), i)
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			c.SetWithDuration(fmt.Sprintf("key%d", i), i, time.Minute)
			i++
		}
	})
}

func BenchmarkGetLRU(b *testing.B) {
	c := NewCache[any](WithCleanupInterval(0), WithMaxKeys(lruCap, PolicyLRU))
	defer c.Close()
	for i := 0; i < lruCap; i++ {
		c.Set(fmt.Sprintf("key%d", i), i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Get(fmt.Sprintf("key%d", i%lruCap))
	}
}

func BenchmarkConcurrentGetLRU(b *testing.B) {
	c := NewCache[any](WithCleanupInterval(0), WithMaxKeys(lruCap, PolicyLRU))
	defer c.Close()
	for i := 0; i < lruCap; i++ {
		c.Set(fmt.Sprintf("key%d", i), i)
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			c.Get(fmt.Sprintf("key%d", i%lruCap))
			i++
		}
	})
}

// ── LFU eviction ─────────────────────────────────────────────────────────────

const lfuCap = 1000

func BenchmarkSetLFU(b *testing.B) {
	c := NewCache[any](WithCleanupInterval(0), WithMaxKeys(lfuCap, PolicyLFU))
	defer c.Close()
	for i := 0; i < lfuCap; i++ {
		c.Set(fmt.Sprintf("seed%d", i), i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Set(fmt.Sprintf("key%d", i), i)
	}
}

func BenchmarkConcurrentSetLFU(b *testing.B) {
	c := NewCache[any](WithCleanupInterval(0), WithMaxKeys(lfuCap, PolicyLFU))
	defer c.Close()
	for i := 0; i < lfuCap; i++ {
		c.Set(fmt.Sprintf("seed%d", i), i)
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			c.Set(fmt.Sprintf("key%d", i), i)
			i++
		}
	})
}

func BenchmarkSetWithDurationLFU(b *testing.B) {
	c := NewCache[any](WithCleanupInterval(0), WithMaxKeys(lfuCap, PolicyLFU))
	defer c.Close()
	for i := 0; i < lfuCap; i++ {
		c.Set(fmt.Sprintf("seed%d", i), i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.SetWithDuration(fmt.Sprintf("key%d", i), i, time.Minute)
	}
}

func BenchmarkConcurrentSetWithDurationLFU(b *testing.B) {
	c := NewCache[any](WithCleanupInterval(0), WithMaxKeys(lfuCap, PolicyLFU))
	defer c.Close()
	for i := 0; i < lfuCap; i++ {
		c.Set(fmt.Sprintf("seed%d", i), i)
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			c.SetWithDuration(fmt.Sprintf("key%d", i), i, time.Minute)
			i++
		}
	})
}

func BenchmarkGetLFU(b *testing.B) {
	c := NewCache[any](WithCleanupInterval(0), WithMaxKeys(lfuCap, PolicyLFU))
	defer c.Close()
	for i := 0; i < lfuCap; i++ {
		c.Set(fmt.Sprintf("key%d", i), i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Get(fmt.Sprintf("key%d", i%lfuCap))
	}
}

func BenchmarkConcurrentGetLFU(b *testing.B) {
	c := NewCache[any](WithCleanupInterval(0), WithMaxKeys(lfuCap, PolicyLFU))
	defer c.Close()
	for i := 0; i < lfuCap; i++ {
		c.Set(fmt.Sprintf("key%d", i), i)
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			c.Get(fmt.Sprintf("key%d", i%lfuCap))
			i++
		}
	})
}
