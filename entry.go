package memstore

import "time"

// Entry represents a single cache item.
type Entry[V any] struct {
	value  V
	expiry time.Time // zero means no expiry
}

func (e *Entry[V]) isExpired() bool {
	if e == nil {
		return true
	}
	if e.expiry.IsZero() {
		return false
	}
	return time.Now().After(e.expiry)
}
