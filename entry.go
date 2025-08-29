package memstore

import "time"

// Entry represents a single cache item
type Entry struct {
	value  interface{}
	expiry time.Time // zero means no expiry
}

// isExpired checks if entry is expired
func (e *Entry) isExpired() bool {
	if e == nil {
		return true
	}
	if e.expiry.IsZero() {
		return false
	}
	return time.Now().After(e.expiry)
}
