// Package testutil provides testing utilities for MCP servers and transports.
package testutil

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/project/mcp-go-core/modules/transport"
)

// EchoServer is a test server that echoes back received messages.
type EchoServer struct {
	mu       sync.Mutex
	received []json.RawMessage
	sessions map[transport.SessionID]string
}

// NewEchoServer creates a new echo server.
func NewEchoServer() *EchoServer {
	return &EchoServer{
		sessions: make(map[transport.SessionID]string),
	}
}

// Handler returns a Handler that echoes back the received message.
func (s *EchoServer) Handler() transport.Handler {
	return func(ctx context.Context, msg json.RawMessage) (any, error) {
		s.mu.Lock()
		s.received = append(s.received, msg)
		s.mu.Unlock()

		return map[string]any{
			"jsonrpc": "2.0",
			"id":      nil,
			"echo":    string(msg),
		}, nil
	}
}

// Received returns all received messages.
func (s *EchoServer) Received() []json.RawMessage {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]json.RawMessage, len(s.received))
	copy(out, s.received)
	return out
}

// ReceivedCount returns the number of received messages.
func (s *EchoServer) ReceivedCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.received)
}

// GetSession returns the session data for a session ID.
func (s *EchoServer) GetSession(id transport.SessionID) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.sessions[id]
}

// SetSession sets session data.
func (s *EchoServer) SetSession(id transport.SessionID, data string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[id] = data
}
