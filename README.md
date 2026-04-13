# go-memstore

A lightweight, concurrency-safe, in-memory key-value store for Go with per-key TTL, pattern-based key lookup, eviction policies, and 24-hour rolling stats.

[![Go Reference](https://pkg.go.dev/badge/github.com/San-B-09/go-memstore.svg)](https://pkg.go.dev/github.com/San-B-09/go-memstore)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

---

## Features

- **Generic** — declare the value type once; no type assertions at call sites
- Per-key TTL with lazy and/or background expiry
- `Keys(pattern)` with `*` wildcard support
- Eviction policies: `PolicyNone`, `PolicyLRU`, `PolicyLFU`
- 24-hour rolling stats (hits, misses, evictions) via `StatsProvider`
- Safe for concurrent use by any number of goroutines
- Zero dependencies beyond the standard library (testify for tests only)

---

## Installation

```bash
go get github.com/San-B-09/go-memstore
```

Requires Go 1.21+.

---

## Quick Start

```go
import memstore "github.com/San-B-09/go-memstore"

// Declare the value type once at construction — no type assertions later
c := memstore.NewCache[string]()
defer c.Close()

c.Set("user:1", "alice")
c.SetWithDuration("session:abc", "token", 30*time.Minute)

val, ok := c.Get("user:1")  // val is string, not interface{}
_, ok = c.Get("missing")    // ("", false)

c.Exists("user:1")          // true
c.Delete("user:1")          // true (existed)

keys := c.Keys("session:*") // pattern match with * wildcard
fmt.Println(c.Len())        // number of live keys
```

---

## TTL / Expiry

Keys set via `Set` never expire. Keys set via `SetWithDuration` expire after the given duration.

| Mode | How to configure | Behaviour |
|---|---|---|
| **Background cleanup** (default) | `WithCleanupInterval(d)` | A goroutine sweeps expired keys every `d`. Default: 1 minute. |
| **Lazy expiry** | `WithCleanupInterval(0)` | Keys are removed on the next access instead of proactively. |

```go
// Lazy expiry only (no goroutine)
c := memstore.NewCache[string](memstore.WithCleanupInterval(0))

// Fast cleanup for short-lived sessions
c := memstore.NewCache[string](memstore.WithCleanupInterval(10 * time.Second))
```

---

## Eviction Policies

Limit the maximum number of keys and choose what happens when the limit is reached.

```go
// Reject new keys once 10 000 are stored
c := memstore.NewCache[string](memstore.WithMaxKeys(10_000, memstore.PolicyNone))

// Evict the least-recently-used key to make room
c := memstore.NewCache[string](memstore.WithMaxKeys(10_000, memstore.PolicyLRU))

// Evict the least-frequently-used key to make room
c := memstore.NewCache[string](memstore.WithMaxKeys(10_000, memstore.PolicyLFU))
```

| Policy | Constant | Behaviour |
|---|---|---|
| No eviction | `PolicyNone` | New keys are silently rejected when the limit is reached. Overwrites always succeed. |
| Least Recently Used | `PolicyLRU` | Evicts the key that has not been accessed for the longest time. O(1) via doubly-linked list. |
| Least Frequently Used | `PolicyLFU` | Evicts the key with the fewest total accesses. Ties broken by recency. O(1) via frequency buckets. |

> **Note:** With `PolicyLRU` or `PolicyLFU`, `Get` uses a write lock internally to update access order. This maintains correctness but reduces concurrent read throughput compared to `PolicyNone`. See [Benchmarks](#benchmarks) below.

---

## Stats (24-hour rolling window)

Stats are **not** part of the core `Cache` interface. Access them via a type assertion to `StatsProvider`:

```go
c := memstore.NewCache[string]()

c.Set("k", "v")
c.Get("k")       // hit
c.Get("missing") // miss

if sp, ok := c.(memstore.StatsProvider); ok {
    s := sp.Stats()
    _ = s.Hits      // uint64
    _ = s.Misses    // uint64
    _ = s.Evictions // uint64
    _ = s.Keys      // int — live key count at time of call
}
```

Stats are accumulated in 5-minute buckets over a 24-hour sliding window. Buckets older than 24 hours are automatically discarded.

---

## API Reference

### `Cache[V]` interface

| Method | Description |
|---|---|
| `Set(key, value)` | Store a value with no expiry |
| `SetWithDuration(key, value, d)` | Store a value that expires after `d` |
| `Get(key) (V, bool)` | Retrieve a value; `false` if missing or expired |
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

## Benchmarks

Measured on Apple M-series (amd64 via Rosetta), Go 1.22, `PolicyNone` (no eviction tracking):

| Benchmark | ops/sec | ns/op | B/op | allocs/op |
|---|---|---|---|---|
| `BenchmarkSet` | 2 274 292 | 536 | 188 | 4 |
| `BenchmarkSetWithDuration` | 1 990 232 | 669 | 204 | 4 |
| `BenchmarkGet` | 10 494 510 | 113 | 0 | 0 |
| `BenchmarkConcurrentSet` | 2 352 498 | 472 | 90 | 4 |
| `BenchmarkConcurrentGet` | 3 220 123 | 351 | 13 | 1 |

**Key observations:**

- `Get` is zero-allocation and fast (~113 ns) under `PolicyNone` because it uses a read lock that scales with concurrent readers.
- With `PolicyLRU` or `PolicyLFU`, `Get` switches to a write lock to maintain access order — expect ~3–4× higher latency under concurrent reads.
- The gap between sequential Get (113 ns) and concurrent Get (351 ns) is caused by the `statsRing` mutex, which is acquired on every `Get`. A future optimisation would switch to atomic counters within a bucket and only lock on bucket rotation (every 5 minutes).

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
