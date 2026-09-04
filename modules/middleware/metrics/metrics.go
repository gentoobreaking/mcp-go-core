// Package metrics provides metrics collection middleware.
package metrics

import (
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/project/mcp-go-core/core/middleware"
)

var (
	httpRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mcp_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)
	requestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mcp_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
)

// Metrics holds metrics collection state.
type Metrics struct {
	enabled bool
	mu      sync.RWMutex
}

// NewMetrics creates a new Metrics middleware instance.
func NewMetrics() *Metrics {
	return &Metrics{enabled: true}
}

// Configure sets up metrics collection.
func (m *Metrics) Configure() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.enabled = true
	return nil
}

// Middleware returns a middleware that collects Prometheus metrics.
func (m *Metrics) Middleware() middleware.Middleware {
	return func(next middleware.Handler) middleware.Handler {
		return middleware.HandlerFunc(func(method string, params []byte) ([]byte, error) {
			start := time.Now()
			result, err := next.HandleRequest(method, params)
			duration := time.Since(start).Seconds()

			httpRequests.WithLabelValues(method, "mcp", "200").Inc()
			requestDuration.WithLabelValues(method, "mcp").Observe(duration)

			if err != nil {
				httpRequests.WithLabelValues(method, "mcp", "500").Inc()
			}

			return result, err
		})
	}
}

// Handler returns an HTTP handler for Prometheus metrics endpoint.
func (m *Metrics) Handler() http.Handler {
	return promhttp.Handler()
}
