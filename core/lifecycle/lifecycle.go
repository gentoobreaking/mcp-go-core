// Package lifecycle provides server lifecycle state management.
package lifecycle

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// State represents a lifecycle phase.
type State int

const (
	Created State = iota
	Configured
	Initialized
	Started
	Running
	ShuttingDown
	Shutdown
)

// String returns the state name.
func (s State) String() string {
	states := []string{"Created", "Configured", "Initialized", "Started", "Running", "ShuttingDown", "Shutdown"}
	if int(s) < len(states) {
		return states[s]
	}
	return "Unknown"
}

// Manager manages lifecycle transitions.
type Manager struct {
	mu      sync.Mutex
	state   State
	timeout time.Duration
}

// NewManager creates a new lifecycle Manager.
func NewManager() *Manager {
	return &Manager{state: Created, timeout: 10 * time.Second}
}

// SetTimeout configures the shutdown timeout.
func (m *Manager) SetTimeout(t time.Duration) {
	m.timeout = t
}

// Transition moves to the next state if valid.
func (m *Manager) Transition(to State) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.canTransition(m.state, to) {
		return fmt.Errorf("invalid transition from %s to %s", m.state, to)
	}
	m.state = to
	return nil
}

func (m *Manager) canTransition(from, to State) bool {
	transitions := map[State][]State{
		Created:      {Configured},
		Configured:   {Initialized},
		Initialized:  {Started},
		Started:      {Running},
		Running:      {ShuttingDown},
		ShuttingDown: {Shutdown},
	}
	for _, s := range transitions[from] {
		if s == to {
			return true
		}
	}
	return false
}

// CurrentState returns the current lifecycle state.
func (m *Manager) CurrentState() State {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.state
}

// Shutdown performs graceful shutdown with context support.
func (m *Manager) Shutdown(ctx context.Context) error {
	if err := m.Transition(ShuttingDown); err != nil {
		return err
	}
	done := make(chan struct{})
	go func() {
		m.Transition(Shutdown)
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(m.timeout):
		return fmt.Errorf("shutdown timeout")
	}
}
