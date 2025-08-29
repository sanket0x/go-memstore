# go-memstore

A lightweight, concurrency-safe in-memory key-value store for Go with optional TTL and background cleanup.

## Features

- Simple, idiomatic Go API (interface-based)
- Per-key TTL support
- Optional background cleanup goroutine (configurable interval; default 1 minute)
- Pattern-based `Keys(pattern)` (supports `*` wildcard)
- Functional options for configuration
- Fully tested

---

## Installation

```bash
go get github.com/yourusername/go-memstore


## Example

```bash
go run example.go