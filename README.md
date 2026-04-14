# go-memstore

A lightweight, concurrency-safe, in-memory key-value store for Go with per-key TTL, pattern-based key lookup, and eviction policies.

[![CI](https://github.com/sanket0x/go-memstore/actions/workflows/ci.yml/badge.svg)](https://github.com/sanket0x/go-memstore/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/sanket0x/go-memstore/branch/main/graph/badge.svg)](https://codecov.io/gh/sanket0x/go-memstore)
[![Release](https://img.shields.io/github/v/tag/sanket0x/go-memstore?label=release)](https://github.com/sanket0x/go-memstore/tags)
[![Go Report Card](https://goreportcard.com/badge/github.com/sanket0x/go-memstore)](https://goreportcard.com/report/github.com/sanket0x/go-memstore)
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
| `BenchmarkSet` | 661 | 202 | 4 |
| `BenchmarkConcurrentSet` | 575 | 93 | 4 |
| `BenchmarkSetWithDuration` | 699 | 222 | 4 |
| `BenchmarkConcurrentSetWithDuration` | 664 | 95 | 4 |
| `BenchmarkGet` | 14 | 0 | 0 |
| `BenchmarkConcurrentGet` | 162 | 13 | 1 |

**LRU eviction policy** (cache pre-filled to capacity; every Set triggers an eviction)

| Benchmark | ns/op | B/op | allocs/op |
|---|---|---|---|
| `BenchmarkSetLRU` | 490 | 146 | 6 |
| `BenchmarkConcurrentSetLRU` | 716 | 135 | 5 |
| `BenchmarkSetWithDurationLRU` | 603 | 146 | 6 |
| `BenchmarkConcurrentSetWithDurationLRU` | 832 | 137 | 5 |
| `BenchmarkGetLRU` | 158 | 13 | 1 |
| `BenchmarkConcurrentGetLRU` | 379 | 13 | 1 |

**LFU eviction policy** (cache pre-filled to capacity; every Set triggers an eviction)

| Benchmark | ns/op | B/op | allocs/op |
|---|---|---|---|
| `BenchmarkSetLFU` | 598 | 147 | 6 |
| `BenchmarkConcurrentSetLFU` | 845 | 189 | 6 |
| `BenchmarkSetWithDurationLFU` | 733 | 147 | 6 |
| `BenchmarkConcurrentSetWithDurationLFU` | 1031 | 189 | 6 |
| `BenchmarkGetLFU` | 322 | 77 | 3 |
| `BenchmarkConcurrentGetLFU` | 630 | 78 | 3 |

**Notes:**
- `Get` with no policy is 14 ns (read lock + map lookup, zero allocation).
- LRU/LFU `Set` benchmarks are run against a full cache so every operation evicts — this reflects the worst-case steady-state throughput.
- LRU `Get` (158 ns / 1 alloc) — updates the access-order list via a separate tracker lock; concurrent map reads proceed in parallel.
- LFU `Get` (322 ns / 3 allocs) — each access promotes a key between frequency buckets, which may allocate a new list node.

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
