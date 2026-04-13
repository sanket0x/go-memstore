package memstore

import "container/list"

// lfuTracker implements evictionTracker using the O(1) LFU algorithm.
// Keys at the same frequency are ordered by recency (front = most recent).
// All methods must be called with c.mu held.
type lfuTracker struct {
	keyFreq    map[string]int
	freqBucket map[int]*list.List
	keyEl      map[string]*list.Element
	minFreq    int
}

func newLFUTracker() *lfuTracker {
	return &lfuTracker{
		keyFreq:    make(map[string]int),
		freqBucket: make(map[int]*list.List),
		keyEl:      make(map[string]*list.Element),
	}
}

func (t *lfuTracker) onInsert(key string) {
	t.keyFreq[key] = 1
	if t.freqBucket[1] == nil {
		t.freqBucket[1] = list.New()
	}
	el := t.freqBucket[1].PushFront(key)
	t.keyEl[key] = el
	t.minFreq = 1
}

func (t *lfuTracker) onAccess(key string) {
	freq := t.keyFreq[key]

	// remove from current frequency bucket
	el := t.keyEl[key]
	t.freqBucket[freq].Remove(el)
	if t.freqBucket[freq].Len() == 0 {
		delete(t.freqBucket, freq)
		if t.minFreq == freq {
			t.minFreq++
		}
	}

	// add to next frequency bucket
	freq++
	t.keyFreq[key] = freq
	if t.freqBucket[freq] == nil {
		t.freqBucket[freq] = list.New()
	}
	el = t.freqBucket[freq].PushFront(key)
	t.keyEl[key] = el
}

func (t *lfuTracker) onDelete(key string) {
	freq := t.keyFreq[key]
	el := t.keyEl[key]
	t.freqBucket[freq].Remove(el)
	if t.freqBucket[freq].Len() == 0 {
		delete(t.freqBucket, freq)
	}
	delete(t.keyFreq, key)
	delete(t.keyEl, key)
}

// evict removes and returns the least-frequently-used key.
// Among keys at the same frequency, the least-recently-used is chosen.
func (t *lfuTracker) evict() string {
	bucket := t.freqBucket[t.minFreq]
	if bucket == nil || bucket.Len() == 0 {
		return ""
	}
	el := bucket.Back()
	key := el.Value.(string)
	bucket.Remove(el)
	if bucket.Len() == 0 {
		delete(t.freqBucket, t.minFreq)
	}
	delete(t.keyFreq, key)
	delete(t.keyEl, key)
	return key
}
