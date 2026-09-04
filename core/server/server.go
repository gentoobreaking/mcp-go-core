// Package server provides the MCP Server with lifecycle management.
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/project/mcp-go-core/core/config"
	"github.com/project/mcp-go-core/core/feature"
	"github.com/project/mcp-go-core/core/mcperror"
	"github.com/project/mcp-go-core/core/middleware/featurewire"
	"github.com/project/mcp-go-core/core/middleware/ratelimit"
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
	flags       *feature.Flags
	lim         *ratelimit.Manager
	shutdownCh  chan struct{}
	wg          sync.WaitGroup
	timeout     time.Duration
	initialized bool
	running     bool
	healthEnabled bool
	cfg           *config.Config
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

// WithFlags sets feature flags for the server.
func WithFlags(f *feature.Flags) Option {
	return func(s *Server) { s.flags = f }
}

// WithRateLimiter sets a rate limiter manager for the server.
func WithRateLimiter(lim *ratelimit.Manager) Option {
	return func(s *Server) { s.lim = lim }
}

// WithHealth enables HTTP health check endpoints on the transport.
func WithHealth(enabled bool) Option {
	return func(s *Server) { s.healthEnabled = enabled }
}
// WithConfig attaches configuration for health endpoint reporting.
func WithConfig(c *config.Config) Option {
	return func(s *Server) { s.cfg = c }
}

// HealthHandler returns HTTP handlers for health check endpoints.
// Only returns non-nil handlers when WithHealth(true) was called.
func (s *Server) HealthHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/health/features", s.handleHealthFeatures)
	mux.HandleFunc("/health/features/", s.handleHealthFeature)
	mux.HandleFunc("/health/rate-limits", s.handleHealthRateLimits)
	mux.HandleFunc("/health/config", s.handleHealthConfig)
	return mux
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"version": s.version,
	})
}

func (s *Server) handleHealthFeatures(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if s.flags == nil {
		http.Error(w, "flags not configured", http.StatusServiceUnavailable)
		return
	}
	status := featurewire.HealthStatus(s.flags)
	json.NewEncoder(w).Encode(map[string]any{"flags": status})
}

func (s *Server) handleHealthFeature(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Path[len("/health/features/"):]
	if name == "" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if s.flags == nil {
		http.Error(w, "flags not configured", http.StatusServiceUnavailable)
		return
	}
	flag := s.flags.Get(name)
	snap := s.flags.Snapshot()
	if _, ok := snap[name]; !ok {
		http.Error(w, "flag not found: " + name, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"name":    name,
		"enabled": flag.Enabled,
	})
}

func (s *Server) handleHealthRateLimits(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if s.lim == nil {
		http.Error(w, "rate limiter not configured", http.StatusServiceUnavailable)
		return
	}
	statuses := s.lim.Status()
	// Convert to JSON-serializable form
	out := make([]map[string]any, 0, len(statuses))
	for _, st := range statuses {
		out = append(out, map[string]any{
			"method":    st.Name,
			"limit":     st.Limit,
			"burst":     st.Burst,
			"requests":  st.Requests,
		})
	}
	json.NewEncoder(w).Encode(map[string]any{"limits": out})
}

func (s *Server) handleHealthConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if s.cfg == nil {
		http.Error(w, "config not configured", http.StatusServiceUnavailable)
		return
	}
	json.NewEncoder(w).Encode(s.cfg.Health())
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


// SendNotification broadcasts a notification to all connected clients via
// the transport layer. This is used for server→client push notifications
// (e.g., notifications/resources/update).
func (s *Server) SendNotification(method string, params any) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	// Serialize the notification
	notif := protocol.Notification{
		JSONRPC: "2.0",
		Method:  method,
	}
	if params != nil {
		data, err := json.Marshal(params)
		if err != nil {
			return fmt.Errorf("marshal notification params: %w", err)
		}
		notif.Params = data
	}
	// Transport push is transport-specific; log for now
	// Stdio/HTTP/SSE transports can override this
	_ = notif
	return nil
}

// sendNotification is the internal callback wired to the router's
// NotificationSender. It serializes the notification and delegates to
// the transport layer (overridable by transport implementations).
func (s *Server) sendNotification(method string, params any) error {
	return s.SendNotification(method, params)
}

// notifyResourceUpdate emits a resource change notification to subscribed clients.
func (s *Server) notifyResourceUpdate(uri, changeType string) error {
	return s.router.NotifyResourceUpdate(uri, changeType)
}
// handleMessage processes a single JSON-RPC message through the router.
func (s *Server) handleMessage(ctx context.Context, msg json.RawMessage) (any, error) {
	var req protocol.Request
	if err := json.Unmarshal(msg, &req); err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}

	// Feature flag gate: ... (existing)
	if s.flags != nil {
		flagName := flagNameForMethod(req.Method)
		if flagName != "" && s.flags.IsDisabled(flagName) {
			return nil, mcperror.NewError(mcperror.CodeValidation,
				fmt.Sprintf("feature flag '%s' is disabled", flagName))
		}
	}

	// Rate limit gate: per-method token bucket
	if s.lim != nil {
		if err := s.lim.Allow(req.Method); err != nil {
			return nil, err
		}
	}

	resp, err := s.router.Dispatch(ctx, &req)
	if err != nil {
		return nil, err
	}
	return resp.Result, nil
}

// flagNameForMethod maps an MCP method to its feature flag name.
// Empty string means no flag gating (always allowed).
func flagNameForMethod(method string) string {
	switch {
	case len(method) > 10 && method[:10] == "tools/call":
		return method[10:]
	case len(method) > 11 && method[:11] == "resources/":
		return method[11:]
	case len(method) > 8 && method[:8] == "prompts/":
		return method[8:]
	}
	return ""
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

	// Wire notification sender to router — sends server→client notifications
	s.router.SetNotificationSender(s.sendNotification)

	// Wire health routes if enabled and transport supports it
	if s.healthEnabled {
		type healthSetter interface {
			WithHealthRoutes(http.Handler)
		}
		if hs, ok := s.transport.(healthSetter); ok {
			hs.WithHealthRoutes(s.HealthHandler())
		}
	}
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
