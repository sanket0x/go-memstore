# go-memstore

A lightweight, concurrency-safe, in-memory key-value store for Go with per-key TTL, pattern-based key lookup, eviction policies, and 24-hour rolling stats.

[![Go Reference](https://pkg.go.dev/badge/github.com/San-B-09/go-memstore.svg)](https://pkg.go.dev/github.com/San-B-09/go-memstore)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

---

## Features

- Simple, idiomatic Go API (interface-based)
- Per-key TTL with lazy and/or background expiry
- `Keys(pattern)` with `*` wildcard support
- Eviction policies: `PolicyNone` (v1.0), `PolicyLRU` / `PolicyLFU` (v1.1)
- 24-hour rolling stats (hits, misses, evictions) via `StatsProvider`
- Safe for concurrent use by any number of goroutines
- Zero dependencies beyond the standard library (testify for tests only)

---

## Installation

```bash
go get github.com/San-B-09/go-memstore
```

---

## Quick Start

```go
import memstore "github.com/San-B-09/go-memstore"

// Create a cache (background cleanup every 1 minute by default)
c := memstore.NewCache()
defer c.Close()

// Store values
c.Set("user:1", "alice")
c.SetWithDuration("session:abc", token, 30*time.Minute)

// Retrieve
val, ok := c.Get("user:1")   // ("alice", true)
_, ok  = c.Get("missing")    // (nil, false)

// Check existence / delete
c.Exists("user:1")            // true
c.Delete("user:1")            // true (existed); false if missing

// Pattern matching (supports * wildcard)
keys := c.Keys("session:*")   // all session keys
fmt.Println(c.Len())          // number of live (non-expired) keys
```

---

## TTL / Expiry

Keys set via `Set` never expire. Keys set via `SetWithDuration` expire after the given duration.

Two eviction modes are available:

| Mode | How to configure | Behaviour |
|---|---|---|
| **Background cleanup** (default) | `WithCleanupInterval(d)` | A goroutine sweeps expired keys every `d`. Default: 1 minute. |
| **Lazy expiry** | `WithCleanupInterval(0)` | Keys are evicted on the next access instead of proactively. |

```go
// Lazy expiry only (no goroutine)
c := memstore.NewCache(memstore.WithCleanupInterval(0))

// Fast cleanup for short-lived sessions
c := memstore.NewCache(memstore.WithCleanupInterval(10 * time.Second))
```

---

## Eviction Policies

Limit the maximum number of keys and choose what happens when the limit is reached.

```go
c := memstore.NewCache(
    memstore.WithMaxKeys(10_000, memstore.PolicyNone),
)
```

| Policy | Constant | Behaviour |
|---|---|---|
| No eviction | `PolicyNone` | New keys are silently rejected when the limit is reached. Overwrites always succeed. |
| Least Recently Used | `PolicyLRU` | *(v1.1)* |
| Least Frequently Used | `PolicyLFU` | *(v1.1)* |

---

## Stats (24-hour rolling window)

Stats are **not** part of the core `Cache` interface. Access them via a type assertion to `StatsProvider`:

```go
c := memstore.NewCache()

c.Set("k", "v")
c.Get("k")       // hit
c.Get("missing") // miss

if sp, ok := c.(memstore.StatsProvider); ok {
    s := sp.Stats()
    fmt.Printf("hits=%d misses=%d evictions=%d keys=%d\n",
        s.Hits, s.Misses, s.Evictions, s.Keys)
}
```

Stats are accumulated in 5-minute buckets over a 24-hour sliding window. Buckets older than 24 hours are automatically discarded.

---

## API Reference

### `Cache` interface

| Method | Description |
|---|---|
| `Set(key, value)` | Store a value with no expiry |
| `SetWithDuration(key, value, d)` | Store a value that expires after `d` |
| `Get(key) (value, bool)` | Retrieve a value; `false` if missing or expired |
| `Delete(key) bool` | Remove a key; returns `true` if it existed |
| `Exists(key) bool` | `true` if the key exists and has not expired |
| `Keys(pattern) []string` | All live keys matching the pattern (`*` wildcard) |
| `Len() int` | Number of live (non-expired) keys |
| `Close()` | Stop the background cleanup goroutine (idempotent) |

### Options

| Option | Description |
|---|---|
| `WithCleanupInterval(d)` | Background sweep interval; `0` disables the goroutine |
| `WithMaxKeys(n, policy)` | Cap the number of live keys and choose eviction behaviour |

### `StatsProvider` interface

```go
type StatsProvider interface {
    Stats() Stats
}

type Stats struct {
    Hits      uint64
    Misses    uint64
    Evictions uint64
    Keys      int
}
```

---

## Running Tests & Benchmarks

```bash
# All tests with race detector
go test ./... -race

# Benchmarks
go test -bench=. -benchmem ./...
```

---

## License

MIT — see [LICENSE](LICENSE).
