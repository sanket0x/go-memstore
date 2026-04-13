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
