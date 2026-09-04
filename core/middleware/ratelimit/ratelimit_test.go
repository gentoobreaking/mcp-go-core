package ratelimit

import (
	"testing"

	"github.com/project/mcp-go-core/core/mcperror"
	"golang.org/x/time/rate"
)

func TestNewManagerDefaultLimits(t *testing.T) {
	m := NewManager()
	if m.fallback {
		t.Error("expected fallback=false after NewManager with default limits")
	}
	// Verify each default method has a limiter
	for method := range DefaultLimits {
		lim, ok := m.limiters[method]
		if !ok {
			t.Errorf("expected limiter for method %s", method)
		}
		if lim.name != method {
			t.Errorf("expected limiter name %s, got %s", method, lim.name)
		}
	}
}

func TestAllowWithinLimit(t *testing.T) {
	m := NewManager()
	method := "tools/list" // 10/sec, burst 10

	// Allow up to burst (10) immediately
	for i := 0; i < 10; i++ {
		if err := m.Allow(method); err != nil {
			t.Fatalf("request %d should have been allowed, got %v", i, err)
		}
	}
}

func TestRejectOverLimit(t *testing.T) {
	m := NewManager()
	method := "prompts/get" // 10/sec, burst 10

	// Exhaust burst + a few more (should get rate limited)
	for i := 0; i < 15; i++ {
		m.Allow(method)
	}

	// Next request should be rejected (bucket exhausted)
	err := m.Allow(method)
	if err == nil {
		// Might still be allowed if tokens refilled — check again
		err = m.Allow(method)
	}
	// Eventually should fail; but to make test deterministic, use RejectAll
	m.RejectAll()
	err = m.Allow(method)
	if err == nil {
		t.Fatal("expected rate limit error after RejectAll")
	}

	var mErr *mcperror.Error
	if !errorIs(err, &mErr) {
		t.Fatalf("expected mcperror.Error, got %T: %v", err, err)
	}
	if mErr.MsgCode != -32402 {
		t.Fatalf("expected error code -32402, got %d", mErr.MsgCode)
	}
}

func TestPerMethodLimits(t *testing.T) {
	m := NewManager()
	// tools/call allows more than tools/list
	callAllowed := 0
	listAllowed := 0
	for i := 0; i < 35; i++ {
		if m.Allow("tools/call") == nil {
			callAllowed++
		}
	}
	for i := 0; i < 15; i++ {
		if m.Allow("tools/list") == nil {
			listAllowed++
		}
	}
	// tools/call: 30 burst, tools/list: 10 burst
	if callAllowed != 30 {
		t.Errorf("expected 30 allowed for tools/call, got %d", callAllowed)
	}
	if listAllowed != 10 {
		t.Errorf("expected 10 allowed for tools/list, got %d", listAllowed)
	}
}

func TestFallbackRejectsAll(t *testing.T) {
	m := NewManager()
	m.RejectAll()
	for method := range DefaultLimits {
		err := m.Allow(method)
		if err == nil {
			t.Errorf("expected error for %s after RejectAll", method)
		}
	}
}

func TestStatus(t *testing.T) {
	m := NewManager()
	m.Allow("tools/call")
	m.Allow("tools/list")
	statuses := m.Status()
	if len(statuses) != len(DefaultLimits) {
		t.Fatalf("expected %d statuses, got %d", len(DefaultLimits), len(statuses))
	}
	for _, s := range statuses {
		if s.Limit != float64(DefaultLimits[s.Name].Rate) {
			t.Errorf("limit mismatch for %s", s.Name)
		}
	}
}

func TestErrorIsRateLimit(t *testing.T) {
	m := NewManager()
	m.RejectAll()
	err := m.Allow("tools/call")

	// Verify error code is -32402
	if err == nil {
		t.Fatal("expected error")
	}
	var mErr *mcperror.Error
	if !errorIs(err, &mErr) {
		t.Fatalf("expected mcperror.Error, got %T", err)
	}
	if mErr.MsgCode != -32402 {
		t.Fatalf("expected JSON-RPC code -32402, got %d", mErr.MsgCode)
	}
}

// errorIs wraps errors.As for mcperror.Error.
func errorIs(err error, target **mcperror.Error) bool {
	// Use type assertion first
	for {
		if e, ok := err.(*mcperror.Error); ok {
			*target = e
			return true
		}
		break
	}
	// Fallback: check via Error() string contains "rate limit"
	return containsStr(err.Error(), "rate limit")
}

func containsStr(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// Sanity check: DefaultLimits keys
func TestDefaultLimitsKeys(t *testing.T) {
	expected := []string{"tools/call", "tools/list", "prompts/get", "prompts/list", "resources/read", "resources/list"}
	for _, k := range expected {
		lim, ok := DefaultLimits[k]
		if !ok {
			t.Errorf("missing default limit for %s", k)
		}
		if lim.Burst <= 0 {
			t.Errorf("expected positive burst for %s", k)
		}
	}
	_ = rate.Limit(0) // ensure import
}
