package server

import (
	"fmt"
	"sync"
	"time"

	"github.com/project/mcp-go-core/core/feature"
	"github.com/project/mcp-go-core/core/middleware"
	"github.com/project/mcp-go-core/core/prompt"
	"github.com/project/mcp-go-core/core/resource"
	"github.com/project/mcp-go-core/core/router"
	"github.com/project/mcp-go-core/core/tool"
	"github.com/project/mcp-go-core/modules/transport"
)

// Transport is the interface for server transports.
type Transport = transport.Transport

// Middleware wraps a Handler to add cross-cutting behavior.
type Middleware = middleware.Middleware
// Handler is the interface for processing requests.
type Handler = middleware.Handler

// Builder provides a fluent API for constructing a Server.
// Lifecycle order is enforced: WithName → WithTool/WithResource/WithPrompt → WithTransport → Build.
type Builder struct {
	mu        sync.Mutex
	name      string
	timeout   time.Duration
	tools     []tool.Tool
	resources []resource.Resource
	prompts   []prompt.Prompt
	flags     *feature.Flags
	transport Transport
	mw        []Middleware
	built     bool
}

// NewBuilder creates a new Server Builder.
func NewBuilder() *Builder {
	return &Builder{
		name:    "mcp-go-core",
		timeout: 10 * time.Second,
	}
}

// WithName sets the server name.
func (b *Builder) WithName(name string) *Builder {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.name = name
	return b
}

// WithTimeout sets the shutdown timeout.
func (b *Builder) WithTimeout(t time.Duration) *Builder {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.timeout = t
	return b
}

// WithTool registers a tool.
func (b *Builder) WithTool(t tool.Tool) *Builder {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.tools = append(b.tools, t)
	return b
}

// WithResource registers a resource.
func (b *Builder) WithResource(r resource.Resource) *Builder {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.resources = append(b.resources, r)
	return b
}

// WithPrompt registers a prompt.
func (b *Builder) WithPrompt(p prompt.Prompt) *Builder {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.prompts = append(b.prompts, p)
	return b
}
// WithFlags configures feature flags for the server.
func (b *Builder) WithFlags(f *feature.Flags) *Builder {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.flags = f
	return b
}

// WithTransport sets the transport.
func (b *Builder) WithTransport(t Transport) *Builder {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.transport = t
	return b
}

// WithMiddleware adds middleware to the chain.
func (b *Builder) WithMiddleware(mw ...Middleware) *Builder {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.mw = append(b.mw, mw...)
	return b
}

// Build constructs the Server with all configured options.
// Returns an error if no transport is set.
func (b *Builder) Build() (*Server, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.transport == nil {
		return nil, fmt.Errorf("transport is required: call WithTransport first")
	}

	s := &Server{
		name:       b.name,
		router:     router.NewRouter(),
		shutdownCh: make(chan struct{}),
		timeout:    b.timeout,
		flags:      b.flags,
		transport:  b.transport,
		mw:         b.mw,
	}

	for _, t := range b.tools {
		s.router.RegisterTool(t)
	}
	for _, r := range b.resources {
		s.router.RegisterResource(r)
	}
	for _, p := range b.prompts {
		s.router.RegisterPrompt(p)
	}

	b.built = true
	return s, nil
}

// MustBuild constructs the Server, panicking on error.
func (b *Builder) MustBuild() *Server {
	s, err := b.Build()
	if err != nil {
		panic(err)
	}
	return s
}
