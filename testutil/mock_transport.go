package testutil

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/project/mcp-go-core/modules/transport"
)

// MockTransport is a mock Transport implementation for testing.
type MockTransport struct {
	mu          sync.Mutex
	handler     transport.Handler
	requests    []json.RawMessage
	responses   []any
	closed      bool
	intercept   func(json.RawMessage) (any, error)
}

// NewMockTransport creates a new MockTransport.
func NewMockTransport() *MockTransport {
	return &MockTransport{}
}

// Serve implements Transport by processing messages through the handler.
func (m *MockTransport) Serve(ctx context.Context, handler transport.Handler) error {
	m.mu.Lock()
	m.handler = handler
	m.mu.Unlock()

	<-ctx.Done()
	return ctx.Err()
}

// Close shuts down the mock transport.
func (m *MockTransport) Close(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return nil
}

// Send sends a message to the transport for processing.
func (m *MockTransport) Send(ctx context.Context, msg json.RawMessage) (any, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.requests = append(m.requests, msg)

	if m.intercept != nil {
		resp, err := m.intercept(msg)
		m.responses = append(m.responses, resp)
		return resp, err
	}

	if m.handler == nil {
		return nil, nil
	}

	resp, err := m.handler(ctx, msg)
	m.responses = append(m.responses, resp)
	return resp, err
}

// Intercept sets an intercept function that overrides the handler.
func (m *MockTransport) Intercept(fn func(json.RawMessage) (any, error)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.intercept = fn
}

// Requests returns all received requests.
func (m *MockTransport) Requests() []json.RawMessage {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]json.RawMessage, len(m.requests))
	copy(out, m.requests)
	return out
}

// Responses returns all recorded responses.
func (m *MockTransport) Responses() []any {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]any, len(m.responses))
	copy(out, m.responses)
	return out
}

// IsClosed returns whether the transport is closed.
func (m *MockTransport) IsClosed() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.closed
}
