package testutil

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"

	"github.com/project/mcp-go-core/modules/transport"
)

// TestSession is a test harness for session-based transport testing.
type TestSession struct {
	t        *testing.T
	smu      *transport.SessionManager
	results  []any
	mu       sync.Mutex
}

// NewTestSession creates a new TestSession.
func NewTestSession(t *testing.T) *TestSession {
	return &TestSession{
		t:   t,
		smu: transport.NewSessionManager(),
	}
}

// RegisterSession registers a new session.
func (s *TestSession) RegisterSession() transport.SessionID {
	return s.smu.RegisterSession()
}

// GetSession retrieves session info.
func (s *TestSession) GetSession(id transport.SessionID) chan struct{} {
	return s.smu.GetSession(id)
}

// UnregisterSession unregisters a session.
func (s *TestSession) UnregisterSession(id transport.SessionID) {
	s.smu.UnregisterSession(id)
}

// Send sends a message to a handler and records the result.
func (s *TestSession) Send(handler transport.Handler, msg json.RawMessage) (any, error) {
	ctx := context.Background()
	resp, err := handler(ctx, msg)
	s.mu.Lock()
	s.results = append(s.results, resp)
	s.mu.Unlock()
	return resp, err
}

// Results returns all results.
func (s *TestSession) Results() []any {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]any, len(s.results))
	copy(out, s.results)
	return out
}

// SessionCount returns the number of active sessions.
func (s *TestSession) SessionCount() int {
	return s.smu.Count()
}

// AssertConnected asserts that the session is connected.
func (s *TestSession) AssertConnected(id transport.SessionID) {
	if s.smu.GetSession(id) == nil {
		s.t.Fatalf("session %s not found", id)
	}
}

// AssertDisconnected asserts that the session is disconnected.
func (s *TestSession) AssertDisconnected(id transport.SessionID) {
	if s.smu.GetSession(id) != nil {
		s.t.Fatalf("session %s still exists", id)
	}
}

// CloseAll closes all sessions.
func (s *TestSession) CloseAll() {
	s.smu.CloseAll()// This will fail, let me fix below
}

// SessionInfo returns session count for testing.
func (s *TestSession) SessionInfo() string {
	return fmt.Sprintf("sessions: %d", s.smu.Count())
}
