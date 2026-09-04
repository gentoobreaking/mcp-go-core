// Package tracing provides distributed tracing middleware using OpenTelemetry.
package tracing

import (
	"context"
	"sync"
	"time"

	"github.com/project/mcp-go-core/core/middleware"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// Tracing holds tracing configuration.
type Tracing struct {
	enabled bool
	mu      sync.RWMutex
	tracer  trace.Tracer
	meter   metric.Meter
}

// NewTracing creates a new Tracing middleware instance.
func NewTracing() *Tracing {
	return &Tracing{
		enabled: true,
		tracer:  otel.Tracer("mcp-go-core"),
		meter:   otel.Meter("mcp-go-core"),
	}
}

// Configure sets up tracing with the given service name.
func (t *Tracing) Configure(serviceName string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.enabled = true
	t.tracer = otel.Tracer(serviceName)
	t.meter = otel.Meter(serviceName)
	return nil
}

// Middleware returns a middleware that adds distributed tracing.
func (t *Tracing) Middleware() middleware.Middleware {
	return func(next middleware.Handler) middleware.Handler {
		return middleware.HandlerFunc(func(method string, params []byte) ([]byte, error) {
			ctx, span := t.tracer.Start(context.Background(), method)
			defer span.End()

			span.SetAttributes(attribute.String("mcp.method", method))
			_ = ctx

			start := time.Now()
			result, err := next.HandleRequest(method, params)
			duration := time.Since(start)

			span.SetAttributes(attribute.Int64("mcp.duration_ms", duration.Milliseconds()))
			if err != nil {
				span.RecordError(err)
			}

			return result, err
		})
	}
}
