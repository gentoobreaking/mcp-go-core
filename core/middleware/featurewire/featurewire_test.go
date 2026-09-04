package featurewire

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/project/mcp-go-core/core/feature"
	"github.com/project/mcp-go-core/core/middleware"
	"github.com/project/mcp-go-core/core/mcperror"
)

func TestDefaultFlagMapper(t *testing.T) {
	tests := []struct {
		method string
		want   string
	}{
		{"tools/call:mytool", "call:mytool"},
		{"resources/list", "list"},
		{"prompts/list", "list"},
		{"initialize", ""},
		{"", ""},
	}
	for _, tt := range tests {
		got := DefaultFlagMapper(tt.method)
		if got != tt.want {
			t.Errorf("DefaultFlagMapper(%q) = %q, want %q", tt.method, got, tt.want)
		}
	}
}

func TestMiddlewareAllowsEnabled(t *testing.T) {
	flags := feature.NewFlags(map[string]bool{"mytool": true})
	mw := Middleware(flags, func(method string) string { return "mytool" })

	called := false
	handler := mw(middleware.HandlerFunc(func(method string, params []byte) ([]byte, error) {
		called = true
		return []byte(`{"result":"ok"}`), nil
	}))

	_, err := handler.HandleRequest("tools/call:mytool", nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !called {
		t.Fatal("expected handler to be called for enabled flag")
	}
}

func TestMiddlewareBlocksDisabled(t *testing.T) {
	flags := feature.NewFlags(map[string]bool{"mytool": false})
	mw := Middleware(flags, func(method string) string { return "mytool" })

	handler := mw(middleware.HandlerFunc(func(method string, params []byte) ([]byte, error) {
		t.Fatal("handler should not be called for disabled flag")
		return nil, nil
	}))

	_, err := handler.HandleRequest("tools/call:mytool", nil)
	if err == nil {
		t.Fatal("expected error for disabled flag")
	}

	// Verify it's a Validation error
	var mErr *mcperror.Error
	if !errors.As(err, &mErr) {
		t.Fatalf("expected mcperror.Error, got %T", err)
	}
	if mErr.Code != mcperror.CodeValidation {
		t.Fatalf("expected validation error code, got %s", mErr.Code)
	}
}

func TestMiddlewareSkipsUnknownFlags(t *testing.T) {
	// Empty flag name → no gating, always allowed
	flags := feature.NewFlags(map[string]bool{})
	mw := Middleware(flags, func(method string) string { return "" })

	called := false
	handler := mw(middleware.HandlerFunc(func(method string, params []byte) ([]byte, error) {
		called = true
		return []byte(`"ok"`), nil
	}))

	_, err := handler.HandleRequest("initialize", nil)
	if err != nil {
		t.Fatalf("expected no error for empty flag name, got %v", err)
	}
	if !called {
		t.Fatal("expected handler to be called when flag name is empty")
	}
}

func TestHealthStatus(t *testing.T) {
	flags := feature.NewFlags(map[string]bool{"a": true, "b": false})
	status := HealthStatus(flags)

	if status["a"] != true {
		t.Error("expected flag a to be enabled")
	}
	if status["b"] != false {
		t.Error("expected flag b to be disabled")
	}
}

func TestMarshalFlagStatus(t *testing.T) {
	flags := feature.NewFlags(map[string]bool{"myflag": true})
	data, err := MarshalFlagStatus(flags, "tools/call:myflag", "myflag")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("expected valid JSON, got error: %v", err)
	}

	if !result["enabled"].(bool) {
		t.Fatal("expected enabled to be true")
	}
	if result["name"] != "myflag" {
		t.Fatalf("expected name myflag, got %v", result["name"])
	}
}

func TestParseFlagParams(t *testing.T) {
	params := json.RawMessage(`{"method":"tools/call:foo"}`)
	got := ParseFlagParams(params)
	if got != "tools/call:foo" {
		t.Fatalf("expected tools/call:foo, got %s", got)
	}

	empty := ParseFlagParams(nil)
	if empty != "" {
		t.Fatalf("expected empty string for nil params, got %s", empty)
	}
}

// helper used in TestMiddlewareBlocksDisabled for fallback string check
func findSubstring(s, sub string) bool {
	return strings.Contains(s, sub)
}
