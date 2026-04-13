package memstore

import "container/list"

// lruTracker implements evictionTracker using a doubly-linked list.
// Front = most recently used, Back = least recently used.
// All methods must be called with c.mu held.
type lruTracker struct {
	list  *list.List
	index map[string]*list.Element
}

func newLRUTracker() *lruTracker {
	return &lruTracker{
		list:  list.New(),
		index: make(map[string]*list.Element),
	}
}

func (t *lruTracker) onInsert(key string) {
	el := t.list.PushFront(key)
	t.index[key] = el
}

func (t *lruTracker) onAccess(key string) {
	if el, ok := t.index[key]; ok {
		t.list.MoveToFront(el)
	}
}

func (t *lruTracker) onDelete(key string) {
	if el, ok := t.index[key]; ok {
		t.list.Remove(el)
		delete(t.index, key)
	}
}

// evict removes and returns the least-recently-used key.
func (t *lruTracker) evict() string {
	el := t.list.Back()
	if el == nil {
		return ""
	}
	key := el.Value.(string)
	t.list.Remove(el)
	delete(t.index, key)
	return key
}
