package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/project/mcp-go-core/core/feature"
	"github.com/project/mcp-go-core/core/middleware/ratelimit"
)

func TestHandleHealth(t *testing.T) {
	s := &Server{
		name:    "test-server",
		version: "1.0.0",
	}
	rr := httptest.NewRecorder()
	s.handleHealth(rr, &http.Request{})

	var resp map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["status"] != "ok" {
		t.Errorf("expected status=ok, got %s", resp["status"])
	}
	if resp["version"] != "1.0.0" {
		t.Errorf("expected version=1.0.0, got %s", resp["version"])
	}
}

func TestHandleHealthFeatures(t *testing.T) {
	s := &Server{
		flags: feature.NewFlags(map[string]bool{
			"tools/call":  true,
			"prompts/get": false,
		}),
	}
	rr := httptest.NewRecorder()
	s.handleHealthFeatures(rr, &http.Request{})

	var resp map[string]map[string]bool
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	flags, ok := resp["flags"]
	if !ok {
		t.Fatal("expected flags key in response")
	}
	if !flags["tools/call"] {
		t.Error("expected tools/call to be enabled")
	}
	if flags["prompts/get"] {
		t.Error("expected prompts/get to be disabled")
	}
}

func TestHandleHealthFeatureNotFound(t *testing.T) {
	s := &Server{
		flags: feature.NewFlags(map[string]bool{}),
	}
	req := httptest.NewRequest(http.MethodGet, "/health/features/nonexistent", nil)
	rr := httptest.NewRecorder()
	s.handleHealthFeature(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestHandleHealthFeatureFound(t *testing.T) {
	s := &Server{
		flags: feature.NewFlags(map[string]bool{
			"test": true,
		}),
	}
	req := httptest.NewRequest(http.MethodGet, "/health/features/test", nil)
	rr := httptest.NewRecorder()
	s.handleHealthFeature(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	var resp map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["name"] != "test" {
		t.Errorf("expected name=test, got %v", resp["name"])
	}
	if resp["enabled"] != true {
		t.Errorf("expected enabled=true, got %v", resp["enabled"])
	}
}

func TestHandleHealthRateLimits(t *testing.T) {
	s := &Server{
		lim: ratelimit.NewManager(),
	}
	rr := httptest.NewRecorder()
	s.handleHealthRateLimits(rr, &http.Request{})

	var resp map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	limits, ok := resp["limits"].([]any)
	if !ok {
		t.Fatal("expected limits array")
	}
	if len(limits) == 0 {
		t.Error("expected non-empty limits")
	}
}

func TestHandleHealthConfigNotConfigured(t *testing.T) {
	s := &Server{}
	rr := httptest.NewRecorder()
	s.handleHealthConfig(rr, &http.Request{})

	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", rr.Code)
	}
}

func TestHandleHealthFeaturesNilFlags(t *testing.T) {
	s := &Server{}
	rr := httptest.NewRecorder()
	s.handleHealthFeatures(rr, &http.Request{})

	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", rr.Code)
	}
}

func TestHandleHealthRateLimitsNilLim(t *testing.T) {
	s := &Server{}
	rr := httptest.NewRecorder()
	s.handleHealthRateLimits(rr, &http.Request{})

	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", rr.Code)
	}
}

func TestHealthHandlerRoutes(t *testing.T) {
	s := &Server{
		name:    "test",
		version: "1.0.0",
		flags:   feature.NewFlags(map[string]bool{"test": true}),
		lim:     ratelimit.NewManager(),
	}
	mux := s.HealthHandler()

	tests := []struct {
		path     string
		wantCode int
	}{
		{"/health", 200},
		{"/health/features", 200},
		{"/health/features/test", 200},
		{"/health/features/nonexistent", 404},
		{"/health/rate-limits", 200},
		{"/health/config", 503}, // not configured
	}

	for _, tt := range tests {
		req := httptest.NewRequest(http.MethodGet, tt.path, nil)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)
		if rr.Code != tt.wantCode {
			t.Errorf("%s: expected %d, got %d", tt.path, tt.wantCode, rr.Code)
		}
	}
}

func TestHealthHandlerWithHTTPServer(t *testing.T) {
	flags := feature.NewFlags(map[string]bool{"tools/call": true})
	lim := ratelimit.NewManager()

	s := NewServer(
		WithHealth(true),
		WithFlags(flags),
		WithRateLimiter(lim),
	)
	s.flags = flags
	s.lim = lim

	handler := s.HealthHandler()
	if handler == nil {
		t.Fatal("expected non-nil HealthHandler")
	}

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestHealthRoutesNotRegisteredByDefault(t *testing.T) {
	s := NewServer()
	if s.healthEnabled {
		t.Error("expected healthEnabled=false by default")
	}

	s2 := NewServer(WithHealth(true))
	if !s2.healthEnabled {
		t.Error("expected healthEnabled=true after WithHealth(true)")
	}
}
