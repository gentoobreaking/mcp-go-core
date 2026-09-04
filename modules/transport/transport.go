package transport

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"sync"
)

// Handler processes a JSON-RPC message and returns a response.
type Handler func(ctx context.Context, msg json.RawMessage) (any, error)

// SessionID is a unique identifier for a transport session.
type SessionID string

// NewSessionID generates a new random session ID.
func NewSessionID() SessionID {
	b := make([]byte, 16)
	rand.Read(b)
	return SessionID(hex.EncodeToString(b))
}

// Transport is the unified interface for all MCP transports.
type Transport interface {
	// Serve starts the transport server and processes messages via handler.
	Serve(ctx context.Context, handler Handler) error
	// Close shuts down the transport gracefully.
	Close(ctx context.Context) error
}

// SessionManager manages transport sessions.
type SessionManager struct {
	mu       sync.RWMutex
	sessions map[SessionID]chan struct{}
}

// NewSessionManager creates a new session manager.
func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[SessionID]chan struct{}),
	}
}

// RegisterSession creates a new session and returns its ID.
func (sm *SessionManager) RegisterSession() SessionID {
	id := NewSessionID()
	done := make(chan struct{})
	sm.mu.Lock()
	sm.sessions[id] = done
	sm.mu.Unlock()
	return id
}

// GetSession returns the done channel for a session, or nil if not found.
func (sm *SessionManager) GetSession(id SessionID) chan struct{} {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.sessions[id]
}

// UnregisterSession removes a session.
func (sm *SessionManager) UnregisterSession(id SessionID) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if done, ok := sm.sessions[id]; ok {
		close(done)
		delete(sm.sessions, id)
	}
}

// CloseAll closes all sessions.
func (sm *SessionManager) CloseAll() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	for id, done := range sm.sessions {
		close(done)
		delete(sm.sessions, id)
	}
}

// Count returns the number of active sessions.
func (sm *SessionManager) Count() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.sessions)
}
