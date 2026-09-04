// Package ratelimit provides token-bucket rate limiting middleware for MCP servers.
// Uses golang.org/x/time/rate for per-method rate limiting with configurable limits.
package ratelimit

import (
	"fmt"
	"sync"

	"github.com/project/mcp-go-core/core/mcperror"
	"golang.org/x/time/rate"
)

// DefaultLimits per-method rate limits (tokens/second, burst).
var DefaultLimits = map[string]struct {
	Rate  rate.Limit
	Burst int
}{
	"tools/call":     {Rate: 30, Burst: 30},
	"tools/list":     {Rate: 10, Burst: 10},
	"prompts/get":    {Rate: 10, Burst: 10},
	"prompts/list":   {Rate: 10, Burst: 10},
	"resources/read": {Rate: 10, Burst: 10},
	"resources/list": {Rate: 10, Burst: 10},
}

// Limiter is a token-bucket limiter with a name for status reporting.
type Limiter struct {
	name  string
	limit rate.Limit
	burst int
	mu    sync.Mutex
	count int64
	*rate.Limiter
}

// NewLimiter creates a new token-bucket limiter.
func NewLimiter(name string, r rate.Limit, b int) *Limiter {
	return &Limiter{
		name:    name,
		limit:   r,
		burst:   b,
		count:   0,
		Limiter: rate.NewLimiter(r, b),
	}
}

// Status returns the current bucket status.
type Status struct {
	Name     string  `json:"name"`
	Limit    float64 `json:"limit_per_second"`
	Burst    int     `json:"burst"`
	Requests int64   `json:"total_requests"`
}

// Manager holds per-method rate limiters.
// thread-safe via read/write mutex.
type Manager struct {
	mu       sync.RWMutex
	limiters map[string]*Limiter
	fallback bool // reject all if true (misconfigured)
}

// ErrRateLimit is a JSON-RPC 2.0 error for rate limiting.
// JSON-RPC code -32402 is reserved for rate limit exceeded.
var ErrRateLimit = mcperror.NewJSONRPCError(-32402, "rate limit exceeded")

// NewManager creates a Manager with default limits.
func NewManager() *Manager {
	m := &Manager{
		limiters: make(map[string]*Limiter),
		fallback: true,
	}
	m.Init(DefaultLimits)
	return m
}

// Init configures per-method limiters from a limits map.
func (m *Manager) Init(limits map[string]struct {
	Rate  rate.Limit
	Burst int
}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.limiters = make(map[string]*Limiter)
	for method, lim := range limits {
		m.limiters[method] = NewLimiter(method, lim.Rate, lim.Burst)
	}
	m.fallback = false
}

// Allow checks if a method is allowed under the rate limit.
// Returns nil if allowed, ErrRateLimit if exceeded.
func (m *Manager) Allow(method string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.fallback {
		return ErrRateLimit
	}

	lim, ok := m.limiters[method]
	if !ok {
		// No limiter for this method → allow
		return nil
	}

	lim.mu.Lock()
	lim.count++
	lim.mu.Unlock()

	if !lim.Allow() {
		return ErrRateLimit
	}
	return nil
}

// Status returns status of all limiters.
func (m *Manager) Status() []Status {
	m.mu.RLock()
	defer m.mu.RUnlock()
	statuses := make([]Status, 0, len(m.limiters))
	for name, lim := range m.limiters {
		lim.mu.Lock()
		s := Status{
			Name:     name,
			Limit:    float64(lim.limit),
			Burst:    lim.burst,
			Requests: lim.count,
		}
		lim.mu.Unlock()
		statuses = append(statuses, s)
	}
	return statuses
}

// AllowAll disables rate limiting (rejects all if misconfigured = false).
func (m *Manager) AllowAll() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.fallback = false
}

// RejectAll causes Allow() to reject every request.
func (m *Manager) RejectAll() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.fallback = true
}

// String returns a human-readable description of the limiter.
func (l *Limiter) String() string {
	return fmt.Sprintf("limiter[name=%s, rate=%v, burst=%d]",
		l.name, l.limit, l.burst)
}
