package store

import (
	"context"
	"sync"
)

type MemoryStore struct {
	mu    sync.RWMutex
	links map[string]Link
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		links: make(map[string]Link),
	}
}

// Save matches the LinkStore interface, so *MemoryStore satisfies LinkStore.
// Pointer receivers mean *MemoryStore (not MemoryStore) implements the interface.
func (m *MemoryStore) Save(_ context.Context, link Link) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.links[link.Code] = link
	return nil
}

// Get matches the LinkStore interface, so *MemoryStore satisfies LinkStore.
// This is how the service can accept any store implementation.
func (m *MemoryStore) Get(_ context.Context, code string) (Link, bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	link, ok := m.links[code]
	return link, ok, nil
}
