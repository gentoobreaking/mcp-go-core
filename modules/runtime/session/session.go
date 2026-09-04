// Package session provides session management for MCP connections.
// Sessions track client connections, lifecycle state, and metadata.
package session

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/project/mcp-go-core/core/lifecycle"
)

var (
	ErrSessionNotFound = errors.New("session not found")
	ErrSessionActive  = errors.New("session is still active")
)

// Session represents an MCP client session.
type Session struct {
	ID        string
	Name      string
	Info      map[string]any
	State     lifecycle.State
	CreatedAt time.Time
	UpdatedAt time.Time
	ClosedAt  time.Time
	lm        *lifecycle.Manager
}

// Manager manages sessions with lifecycle tracking.
type Manager struct {
	mu       sync.RWMutex
	sessions map[string]*Session
	lm       *lifecycle.Manager
	counter  int64
}

// NewManager creates a new session Manager.
func NewManager() *Manager {
	return &Manager{
		sessions: make(map[string]*Session),
		lm:       lifecycle.NewManager(),
	}
}

// Create creates a new session with the given name and metadata.
func (m *Manager) Create(name string, info map[string]any) (*Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	id := m.generateID()
	s := &Session{
		ID:        id,
		Name:      name,
		Info:      info,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		lm:        lifecycle.NewManager(), // each session gets its own lifecycle manager
	}

	// Initialize lifecycle: Created -> Configured -> Initialized -> Started -> Running
	if err := s.lm.Transition(lifecycle.Configured); err != nil {
		return nil, fmt.Errorf("session lifecycle configure: %w", err)
	}
	if err := s.lm.Transition(lifecycle.Initialized); err != nil {
		return nil, fmt.Errorf("session lifecycle init: %w", err)
	}
	if err := s.lm.Transition(lifecycle.Started); err != nil {
		return nil, fmt.Errorf("session lifecycle start: %w", err)
	}
	if err := s.lm.Transition(lifecycle.Running); err != nil {
		return nil, fmt.Errorf("session lifecycle running: %w", err)
	}

	s.State = lifecycle.Running
	m.sessions[id] = s

	return s, nil
}

// Get retrieves a session by ID.
func (m *Manager) Get(id string) (*Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	s, ok := m.sessions[id]
	if !ok {
		return nil, ErrSessionNotFound
	}
	return s, nil
}
// Destroy removes a session by ID.
func (m *Manager) Destroy(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	s, ok := m.sessions[id]
	if !ok {
		return ErrSessionNotFound
	}

	if s.State == lifecycle.Running {
		_ = s.lm.Transition(lifecycle.ShuttingDown)
	}
	_ = s.lm.Transition(lifecycle.Shutdown)

	s.ClosedAt = time.Now()
	s.State = lifecycle.Shutdown
	delete(m.sessions, id)

	return nil
}
func (m *Manager) DestroyAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, s := range m.sessions {
		if s.State == lifecycle.Running {
			_ = s.lm.Transition(lifecycle.ShuttingDown)
		}
		_ = s.lm.Transition(lifecycle.Shutdown)
		s.ClosedAt = time.Now()
		s.State = lifecycle.Shutdown
		delete(m.sessions, id)
	}
}

// Count returns the number of active sessions.
func (m *Manager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.sessions)
}

// ActiveSessions returns all currently active sessions.
func (m *Manager) ActiveSessions() []*Session {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var active []*Session
	for _, s := range m.sessions {
		if s.State == lifecycle.Running {
			active = append(active, s)
		}
	}
	return active
}

// CloseAll shuts down the session manager, destroying all sessions.
func (m *Manager) CloseAll() {
	m.DestroyAll()
}

// Close gracefully closes a session by ID, transitioning through lifecycle states.
func (m *Manager) Close(ctx context.Context, id string) error {
	s, err := m.Get(id)
	if err != nil {
		return err
	}

	if err := s.lm.Transition(lifecycle.ShuttingDown); err != nil {
		return err
	}

	done := make(chan error, 1)
	go func() {
		// In a real implementation, this would wait for in-flight requests
		done <- s.lm.Transition(lifecycle.Shutdown)
	}()

	select {
	case err := <-done:
		if err != nil {
			return err
		}
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(10 * time.Second):
		return fmt.Errorf("session %s shutdown timed out", id)
	}

	// Remove session from manager after successful shutdown
	m.mu.Lock()
	defer m.mu.Unlock()
	s.ClosedAt = time.Now()
	s.State = lifecycle.Shutdown
	delete(m.sessions, id)

	return nil
}

// generateID creates a unique session ID using a monotonic counter.
func (m *Manager) generateID() string {
	m.counter++
	return fmt.Sprintf("sess_%d_%d", time.Now().UnixNano(), m.counter)
}
