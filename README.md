# go-memstore

A lightweight, concurrency-safe, in-memory key-value store for Go with per-key TTL, pattern-based key lookup, and eviction policies.

[![Go Reference](https://pkg.go.dev/badge/github.com/sanket0x/go-memstore.svg)](https://pkg.go.dev/github.com/sanket0x/go-memstore)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

---

## Features

- **Generic** - declare the value type once; no type assertions at call sites
- Per-key TTL with lazy and/or background expiry
- `Keys(pattern)` with `*` wildcard support
- Eviction policies: `PolicyNone`, `PolicyLRU`, `PolicyLFU`
- Opt-in 24-hour rolling stats (zero overhead when disabled)
- Safe for concurrent use by any number of goroutines
- No runtime dependencies

---

## Installation

```bash
go get github.com/sanket0x/go-memstore
```

Requires Go 1.21+.

---

## Quick Start

```go
import memstore "github.com/sanket0x/go-memstore"

// Declare the value type once - no type assertions later
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
// Reject new keys once 10,000 are stored
c := memstore.NewCache[string](memstore.WithMaxKeys(10_000, memstore.PolicyNone))

// Evict the least-recently-used key to make room
c := memstore.NewCache[string](memstore.WithMaxKeys(10_000, memstore.PolicyLRU))

// Evict the least-frequently-used key to make room
c := memstore.NewCache[string](memstore.WithMaxKeys(10_000, memstore.PolicyLFU))
```

| Policy | Constant | Behaviour |
|---|---|---|
| No eviction | `PolicyNone` | New keys are rejected when the limit is reached - `Set` returns `ErrCacheFull`. Overwrites always succeed. |
| Least Recently Used | `PolicyLRU` | Evicts the key that has not been accessed for the longest time. O(1) via doubly-linked list. |
| Least Frequently Used | `PolicyLFU` | Evicts the key with the fewest total accesses. Ties broken by recency. O(1) via frequency buckets. |

> `Set` and `SetWithDuration` return `error`. With `PolicyNone`, a full cache returns `ErrCacheFull` which can be checked with `errors.Is`. With `PolicyLRU`/`PolicyLFU`, a key is always evicted to make room so the error is always `nil`.

---

## Stats (opt-in, 24-hour rolling window)

Stats are **disabled by default** - no overhead unless you enable them. Use `WithStats()` which returns both an option and a handle:

```go
statsOpt, stats := memstore.WithStats()
c := memstore.NewCache[string](statsOpt)

c.Set("k", "v")
c.Get("k")       // hit
c.Get("missing") // miss

s := stats.Snapshot()
_ = s.Hits      // uint64
_ = s.Misses    // uint64
_ = s.Evictions // uint64
```

Stats are accumulated in 5-minute buckets over a 24-hour sliding window.

---

## Reference

### `Cache[V]` methods

| Method | Returns | Description |
|---|---|---|
| `Set(key, value)` | `error` | Store a value with no expiry; `ErrCacheFull` if rejected by `PolicyNone` |
| `SetWithDuration(key, value, d)` | `error` | Store a value expiring after `d`; `ErrCacheFull` if rejected by `PolicyNone` |
| `Get(key)` | `(V, bool)` | Retrieve a value; zero value + `false` if missing or expired |
| `Delete(key)` | `bool` | Remove a key; `true` if it existed |
| `Exists(key)` | `bool` | `true` if the key exists and has not expired |
| `Keys(pattern)` | `[]string` | All live keys matching the pattern (`*` wildcard) |
| `Len()` | `int` | Number of live (non-expired) keys |
| `Close()` | - | Stop the background cleanup goroutine (idempotent) |

### Options

| Option | Returns | Description |
|---|---|---|
| `WithCleanupInterval(d)` | `Option` | Background sweep interval; `0` disables the goroutine |
| `WithMaxKeys(n, policy)` | `Option` | Cap the number of live keys and choose eviction behaviour |
| `WithStats()` | `(Option, *StatsHandle)` | Enable rolling stats collection |

---

## Benchmarks

Measured on Apple M-series (amd64 via Rosetta), Go 1.22. Stats disabled in all runs.

**No eviction policy**

| Benchmark | ns/op | B/op | allocs/op |
|---|---|---|---|
| `BenchmarkSet` | 642 | 194 | 4 |
| `BenchmarkSetWithDuration` | 733 | 207 | 4 |
| `BenchmarkGet` | 14 | 0 | 0 |
| `BenchmarkConcurrentSet` | 502 | 93 | 4 |
| `BenchmarkConcurrentGet` | 164 | 13 | 1 |

**With eviction policies (`Get` only - most impacted operation)**

| Benchmark | ns/op | B/op | allocs/op |
|---|---|---|---|
| `BenchmarkGetLRU` | 170 | 13 | 1 |
| `BenchmarkGetLFU` | 326 | 77 | 3 |
| `BenchmarkConcurrentGetLRU` | 404 | 13 | 1 |
| `BenchmarkConcurrentGetLFU` | 553 | 78 | 3 |

**Notes:**
- `Get` with no policy is 14 ns (read lock + map lookup, zero allocation).
- LRU `Get` (170 ns) is only ~12× slower than no-policy because it uses a write lock but only moves a list pointer - no allocation in steady state.
- LFU `Get` (326 ns / 3 allocs) - each access promotes a key between frequency buckets, which may allocate a new list node.
- Concurrent LRU/LFU gets serialise on a write lock; throughput scales with single-goroutine latency.

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

MIT - see [LICENSE](LICENSE).
