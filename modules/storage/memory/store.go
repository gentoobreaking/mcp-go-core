// Package memory provides in-memory storage for MCP servers.
package memory

import (
	"context"
	"errors"
	"sync"
)

var ErrNotFound = errors.New("key not found")

// Store implements in-memory storage.
type Store struct {
	mu   sync.RWMutex
	data map[string][]byte
}

// New creates a new in-memory store.
func New() *Store {
	return &Store{
		data: make(map[string][]byte),
	}
}

// Get retrieves a value by key.
func (s *Store) Get(ctx context.Context, key string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.data[key]
	if !ok {
		return nil, ErrNotFound
	}
	return val, nil
}

// Set stores a value by key.
func (s *Store) Set(ctx context.Context, key string, value []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
	return nil
}

// Delete removes a key.
func (s *Store) Delete(ctx context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
	return nil
}

// Keys returns all keys.
func (s *Store) Keys(ctx context.Context) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	keys := make([]string, 0, len(s.data))
	for k := range s.data {
		keys = append(keys, k)
	}
	return keys, nil
}
