package metrics

import (
	"testing"

	"github.com/project/mcp-go-core/core/middleware"
)

func TestNewMetrics(t *testing.T) {
	m := NewMetrics()
	if m == nil {
		t.Fatal("expected non-nil metrics")
	}
}

func TestMetricsConfigure(t *testing.T) {
	m := NewMetrics()
	if err := m.Configure(); err != nil {
		t.Fatal(err)
	}
}

func TestMetricsMiddleware(t *testing.T) {
	m := NewMetrics()
	h := middleware.HandlerFunc(func(method string, params []byte) ([]byte, error) {
		return []byte("ok"), nil
	})

	chained := middleware.Chain(h, m.Middleware())
	result, err := chained.HandleRequest("tools/call", nil)
	if err != nil {
		t.Fatal(err)
	}
	if string(result) != "ok" {
		t.Fatal("result mismatch")
	}
}

func TestMetricsHandler(t *testing.T) {
	m := NewMetrics()
	handler := m.Handler()
	if handler == nil {
		t.Fatal("expected handler")
	}
}
