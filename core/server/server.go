// Package server provides the MCP Server with lifecycle management.
package server

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/project/mcp-go-core/core/prompt"
	"github.com/project/mcp-go-core/core/resource"
	"github.com/project/mcp-go-core/core/router"
	"github.com/project/mcp-go-core/core/tool"
)

// Server represents an MCP server with lifecycle management.
type Server struct {
	mu          sync.RWMutex
	name        string
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

// WithName sets the server name.
func WithName(name string) Option {
	return func(s *Server) { s.name = name }
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

// Run starts the server loop.
func (s *Server) Run(ctx context.Context) error {
	s.mu.Lock()
	s.initialized = true
	s.running = true
	s.mu.Unlock()

	<-ctx.Done()
	return s.Shutdown(ctx)
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
