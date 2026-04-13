package memstore

import (
	"fmt"
	"testing"
	"time"
)

func BenchmarkSet(b *testing.B) {
	c := NewCache[any](WithCleanupInterval(0))
	defer c.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Set(fmt.Sprintf("key%d", i), i)
	}
}

func BenchmarkSetWithDuration(b *testing.B) {
	c := NewCache[any](WithCleanupInterval(0))
	defer c.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.SetWithDuration(fmt.Sprintf("key%d", i), i, time.Minute)
	}
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

func BenchmarkConcurrentGet(b *testing.B) {
	c := NewCache[any](WithCleanupInterval(0))
	defer c.Close()
	// Pre-populate
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

// --- Eviction policy benchmarks (Get only — most impacted by policy) ---

func BenchmarkGetLRU(b *testing.B) {
	const cap = 1000
	c := NewCache[any](WithCleanupInterval(0), WithMaxKeys(cap, PolicyLRU))
	defer c.Close()
	for i := 0; i < cap; i++ {
		c.Set(fmt.Sprintf("key%d", i), i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Get(fmt.Sprintf("key%d", i%cap))
	}
}

func BenchmarkGetLFU(b *testing.B) {
	const cap = 1000
	c := NewCache[any](WithCleanupInterval(0), WithMaxKeys(cap, PolicyLFU))
	defer c.Close()
	for i := 0; i < cap; i++ {
		c.Set(fmt.Sprintf("key%d", i), i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Get(fmt.Sprintf("key%d", i%cap))
	}
}

func BenchmarkConcurrentGetLRU(b *testing.B) {
	const cap = 1000
	c := NewCache[any](WithCleanupInterval(0), WithMaxKeys(cap, PolicyLRU))
	defer c.Close()
	for i := 0; i < cap; i++ {
		c.Set(fmt.Sprintf("key%d", i), i)
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			c.Get(fmt.Sprintf("key%d", i%cap))
			i++
		}
	})
}

func BenchmarkConcurrentGetLFU(b *testing.B) {
	const cap = 1000
	c := NewCache[any](WithCleanupInterval(0), WithMaxKeys(cap, PolicyLFU))
	defer c.Close()
	for i := 0; i < cap; i++ {
		c.Set(fmt.Sprintf("key%d", i), i)
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			c.Get(fmt.Sprintf("key%d", i%cap))
			i++
		}
	})
}
