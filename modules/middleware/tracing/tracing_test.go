package tracing

import (
	"testing"

	"github.com/project/mcp-go-core/core/middleware"
)

func TestNewTracing(t *testing.T) {
	tr := NewTracing()
	if tr == nil {
		t.Fatal("expected non-nil tracing")
	}
}

func TestTracingConfigure(t *testing.T) {
	tr := NewTracing()
	if err := tr.Configure("test-service"); err != nil {
		t.Fatal(err)
	}
}

func TestTracingMiddleware(t *testing.T) {
	tr := NewTracing()
	h := middleware.HandlerFunc(func(method string, params []byte) ([]byte, error) {
		return []byte("ok"), nil
	})

	chained := middleware.Chain(h, tr.Middleware())
	result, err := chained.HandleRequest("tools/list", nil)
	if err != nil {
		t.Fatal(err)
	}
	if string(result) != "ok" {
		t.Fatal("result mismatch")
	}
}
