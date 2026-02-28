package store

import (
	"context"
	"sync"
	"time"
)

type MemoryStore struct {
	mu           sync.RWMutex
	links        map[string]Link
	intentToCode map[string]string
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		links:        make(map[string]Link),
		intentToCode: make(map[string]string),
	}
}

// Save matches the LinkStore interface, so *MemoryStore satisfies LinkStore.
// Pointer receivers mean *MemoryStore (not MemoryStore) implements the interface.
func (m *MemoryStore) Save(_ context.Context, link Link) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.links[link.Code]; exists {
		return ErrCodeExists
	}

	key := BuildIntentKey(link.LongURL, link.ExpiresAt)
	if existingCode, exists := m.intentToCode[key]; exists {
		if _, exists := m.links[existingCode]; exists {
			return ErrIntentExists
		}
		// Defensive cleanup if maps become inconsistent.
		delete(m.intentToCode, key)
	}

	m.links[link.Code] = cloneLink(link)
	m.intentToCode[key] = link.Code
	return nil
}

// Get matches the LinkStore interface, so *MemoryStore satisfies LinkStore.
// This is how the service can accept any store implementation.
func (m *MemoryStore) Get(_ context.Context, code string) (Link, bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	link, ok := m.links[code]
	if !ok {
		return Link{}, false, nil
	}
	return cloneLink(link), true, nil
}

func (m *MemoryStore) FindByIntent(_ context.Context, longURL string, expiresAt *time.Time) (Link, bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	code, ok := m.intentToCode[BuildIntentKey(longURL, expiresAt)]
	if !ok {
		return Link{}, false, nil
	}
	link, ok := m.links[code]
	if !ok {
		return Link{}, false, nil
	}
	return cloneLink(link), true, nil
}

func cloneLink(link Link) Link {
	linkCopy := link
	linkCopy.ExpiresAt = cloneTime(link.ExpiresAt)
	return linkCopy
}

func cloneTime(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	v := t.UTC()
	return &v
}
