package storage

import (
	"github.com/google/uuid"
	"sync"
)

type itemWithID interface {
	ID() uuid.UUID
}

// ConcurrentSlice is a slice which is safe for usage in multiple goroutines.
// It can only hold items which have function ID() uuid.UUID
type ConcurrentSlice struct {
	sync.RWMutex
	items []itemWithID
}

// New returns an empty ConcurrentSlice
func New() ConcurrentSlice {
	items := make([]itemWithID, 0)
	return ConcurrentSlice{sync.RWMutex{}, items}
}

// Append appends item to ConcurrentSlice
func (cs *ConcurrentSlice) Append(item itemWithID) {
	cs.Lock()
	defer cs.Unlock()

	cs.items = append(cs.items, item)
}

// DeleteByID deletes item with correspondind ID from slice
func (cs *ConcurrentSlice) DeleteByID(id uuid.UUID) bool {
	cs.Lock()
	defer cs.Unlock()
	for index, value := range cs.items {
		if value.ID() == id {
			cs.items[len(cs.items)-1], cs.items[index] = cs.items[index], cs.items[len(cs.items)-1]
			cs.items = cs.items[:len(cs.items)-1]
			return true
		}
	}
	return false
}

// Get returns item from ConcurrentSlice with given index
func (cs *ConcurrentSlice) Get(idx int) (interface{}, bool) {
	cs.Lock()
	defer cs.Unlock()
	if idx >= len(cs.items) || idx < 0 {
		return nil, false
	}
	//TODO: think about race condition here
	return cs.items[idx], true
}

// Len returns length of ConcurrentSlice
func (cs *ConcurrentSlice) Len() int {
	cs.Lock()
	defer cs.Unlock()
	return len(cs.items)
}

// CurrentItems returns a normal slice of items in ConcurrentSlice
func (cs *ConcurrentSlice) CurrentItems() []itemWithID {
	cs.Lock()
	defer cs.Unlock()
	currItems := make([]itemWithID, len(cs.items))
	copy(currItems, cs.items)
	return currItems
}
