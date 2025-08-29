//go:build ignore
// +build ignore

// This is a standalone example program, not part of the memstore package build.
// Run it directly with: go run example.go

package main

import (
	"fmt"
	"time"

	"github.com/yourusername/go-memstore"
)

func main() {
	// Create cache with default cleanup interval (1 minute)
	c := memstore.NewCache()
	defer c.Close()

	c.Set("foo", "bar", 0) // ttl = 0 → no expiry
	val, ok := c.Get("foo")
	fmt.Println("foo:", val, ok) // foo: bar true

	// Create cache with lazy expiry (no background cleanup)
	lazy := memstore.NewCache(memstore.WithCleanupInterval(0))
	defer lazy.Close()

	lazy.Set("temp", "gone soon", 2*time.Second)
	time.Sleep(3 * time.Second)
	_, ok = lazy.Get("temp")
	fmt.Println("temp exists after 3s?", ok) // false
}
