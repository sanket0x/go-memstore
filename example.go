//go:build ignore

// Standalone example — run with: go run example.go

package main

import (
	"fmt"
	"time"

	memstore "github.com/San-B-09/go-memstore"
)

func main() {
	basicExample()
	ttlExample()
	lruExample()
}

// basicExample shows simple get/set on a string-typed cache.
func basicExample() {
	fmt.Println("=== Basic ===")

	c := memstore.NewCache[string]()
	defer c.Close()

	c.Set("lang", "Go")
	if val, ok := c.Get("lang"); ok {
		fmt.Println("lang:", val) // Go — no type assertion needed
	}
}

// ttlExample shows lazy expiry (no background goroutine).
func ttlExample() {
	fmt.Println("=== TTL ===")

	c := memstore.NewCache[string](memstore.WithCleanupInterval(0))
	defer c.Close()

	c.SetWithDuration("token", "abc123", 50*time.Millisecond)
	if _, ok := c.Get("token"); ok {
		fmt.Println("token present before expiry")
	}

	time.Sleep(100 * time.Millisecond)

	if _, ok := c.Get("token"); !ok {
		fmt.Println("token expired")
	}
}

// lruExample shows a capacity-bounded cache with LRU eviction.
func lruExample() {
	fmt.Println("=== LRU eviction (max 3 keys) ===")

	c := memstore.NewCache[int](memstore.WithMaxKeys(3, memstore.PolicyLRU))
	defer c.Close()

	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("c", 3)

	// Touch "a" and "b" — "c" becomes the least recently used
	c.Get("a")
	c.Get("b")

	// Inserting "d" evicts "c"
	c.Set("d", 4)

	for _, key := range []string{"a", "b", "c", "d"} {
		val, ok := c.Get(key)
		if ok {
			fmt.Printf("  %s = %d\n", key, val)
		} else {
			fmt.Printf("  %s = (evicted)\n", key)
		}
	}
	// Output:
	//   a = 1
	//   b = 2
	//   c = (evicted)
	//   d = 4
}
