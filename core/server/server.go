// Package server provides the MCP Server with lifecycle management.
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/project/mcp-go-core/core/prompt"
	"github.com/project/mcp-go-core/core/protocol"
	"github.com/project/mcp-go-core/core/resource"
	"github.com/project/mcp-go-core/core/router"
	"github.com/project/mcp-go-core/core/tool"
	"github.com/project/mcp-go-core/modules/transport"
)

// Server represents an MCP server with lifecycle management.
type Server struct {
	mu          sync.RWMutex
	name        string
	version     string
	router      *router.Router
	transport   Transport
	mw          []Middleware
	shutdownCh  chan struct{}
	wg          sync.WaitGroup
	timeout     time.Duration
	initialized bool
	running     bool
}

// Option configures a Server.
type Option func(*Server)

// WithName sets the server name and version.
func WithName(name string) Option {
	return func(s *Server) { s.name = name }
}

// WithVersion sets the server version.
func WithVersion(version string) Option {
	return func(s *Server) { s.version = version }
}

// WithShutdownTimeout sets the shutdown timeout.
func WithShutdownTimeout(t time.Duration) Option {
	return func(s *Server) { s.timeout = t }
}

// State represents the server lifecycle state.
type State int

const (
	StateCreated State = iota
	StateConfigured
	StateInitialized
	StateStarted
	StateRunning
	StateShuttingDown
	StateShutdown
)

// NewServer creates a new Server.
func NewServer(opts ...Option) *Server {
	s := &Server{
		router:     router.NewRouter(),
		shutdownCh: make(chan struct{}),
		timeout:    10 * time.Second,
		name:       "mcp-go-core",
		version:    "0.1.0",
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// AddTool registers a tool.
func (s *Server) AddTool(t tool.Tool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.router.RegisterTool(t)
}

// AddResource registers a resource.
func (s *Server) AddResource(r resource.Resource) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.router.RegisterResource(r)
}

// AddPrompt registers a prompt.
func (s *Server) AddPrompt(p prompt.Prompt) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.router.RegisterPrompt(p)
}

// handleMessage processes a single JSON-RPC message through the router.
func (s *Server) handleMessage(ctx context.Context, msg json.RawMessage) (any, error) {
	var req protocol.Request
	if err := json.Unmarshal(msg, &req); err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}

	resp, err := s.router.Dispatch(ctx, &req)
	if err != nil {
		return nil, err
	}
	return resp.Result, nil
}

// Run starts the server and blocks until context is cancelled.
func (s *Server) Run(ctx context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("server already running")
	}
	if s.transport == nil {
		s.mu.Unlock()
		return fmt.Errorf("no transport configured")
	}
	s.running = true
	s.initialized = true
	s.mu.Unlock()

	// Create the message handler that routes through the router
	handler := transport.Handler(s.handleMessage)

	// Start transport in a goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- s.transport.Serve(ctx, handler)
	}()

	// Wait for context cancellation or transport error
	select {
	case err := <-errCh:
		s.mu.Lock()
		s.running = false
		s.mu.Unlock()
		if err != nil {
			return fmt.Errorf("transport error: %w", err)
		}
		return nil
	case <-ctx.Done():
		s.mu.Lock()
		s.running = false
		s.mu.Unlock()
		return s.Shutdown(ctx)
	}
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown(ctx context.Context) error {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return nil
	}
	s.running = false
	s.mu.Unlock()

	close(s.shutdownCh)

	// Wait for in-flight requests with timeout
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-time.After(s.timeout):
		return fmt.Errorf("shutdown timed out after %s", s.timeout)
	case <-ctx.Done():
		return ctx.Err()
	}
}
