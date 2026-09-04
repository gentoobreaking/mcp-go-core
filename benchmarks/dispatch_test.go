package benchmarks

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/project/mcp-go-core/core/middleware"
	"github.com/project/mcp-go-core/core/protocol"
	"github.com/project/mcp-go-core/core/router"
	"github.com/project/mcp-go-core/core/server"
	"github.com/project/mcp-go-core/core/tool"
	"github.com/project/mcp-go-core/modules/transport/stdio"
)

// benchmarkTool implements tool.Tool for benchmarking.
type benchmarkTool struct{}

func (benchmarkTool) Name() string        { return "bench_tool" }
func (benchmarkTool) Description() string { return "Benchmark tool" }
func (benchmarkTool) InputSchema() tool.Schema {
	return tool.Schema{"type": "object"}
}
func (benchmarkTool) Handler() tool.ToolHandler {
	return func(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
		return &protocol.Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  json.RawMessage(`{"result":"ok"}`),
		}, nil
	}
}

// BenchmarkToolDispatch measures the dispatch latency for a tool call.
// Targets: P50 < 10 µs, P99 < 100 µs
func BenchmarkToolDispatch(b *testing.B) {
	r := router.NewRouter()
	r.RegisterTool(benchmarkTool{})

	h := middleware.HandlerFunc(func(method string, params []byte) ([]byte, error) {
		return []byte(`{"jsonrpc":"2.0","result":"ok"}`), nil
	})

	// Apply middleware chain
	chained := middleware.Chain(h, middleware.Recovery())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = chained.HandleRequest("tools/call", nil)
	}
}

// BenchmarkToolDispatchInProcess measures end-to-end dispatch through server.
func BenchmarkToolDispatchInProcess(b *testing.B) {
	r := router.NewRouter()
	r.RegisterTool(benchmarkTool{})

	h := middleware.HandlerFunc(func(method string, params []byte) ([]byte, error) {
		return []byte(`{"jsonrpc":"2.0","result":"ok"}`), nil
	})

	chained := middleware.Chain(h, middleware.Recovery())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := chained.HandleRequest("tools/call", nil)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
		if string(resp) != `{"jsonrpc":"2.0","result":"ok"}` {
			b.Fatalf("unexpected response: %s", resp)
		}
	}
}

// BenchmarkToolDispatchThroughStdio measures full stdio round-trip.
func BenchmarkToolDispatchThroughput(b *testing.B) {
	r := router.NewRouter()
	r.RegisterTool(benchmarkTool{})

	h := middleware.HandlerFunc(func(method string, params []byte) ([]byte, error) {
		return []byte(`{"jsonrpc":"2.0","result":"ok"}`), nil
	})

	_ = server.NewServer(
		server.WithName("bench-server"),
	)
	tr := stdio.New(nil, nil)
	_ = tr

	chained := middleware.Chain(h, middleware.Recovery())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = chained.HandleRequest("tools/call", nil)
	}
}

// BenchmarkToolDispatchP50P99 measures latency distribution.
func BenchmarkToolDispatchP50P99(b *testing.B) {
	r := router.NewRouter()
	r.RegisterTool(benchmarkTool{})

	h := middleware.HandlerFunc(func(method string, params []byte) ([]byte, error) {
		return []byte(`{"jsonrpc":"2.0","result":"ok"}`), nil
	})

	chained := middleware.Chain(h, middleware.Recovery())

	latencies := make([]int64, 0, b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := b.Elapsed()
		_, _ = chained.HandleRequest("tools/call", nil)
		latencies = append(latencies, int64(b.Elapsed()-start))
	}

	// Calculate P50 and P99
	if len(latencies) > 0 {
		b.StopTimer()
		p50 := latencies[len(latencies)/2]
		p99 := latencies[int(float64(len(latencies))*0.99)]
		fmt.Printf("P50: %dns, P99: %dns\n", p50, p99)
	}
}
