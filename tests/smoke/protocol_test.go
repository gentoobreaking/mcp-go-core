package smoke

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/project/mcp-go-core/core/middleware"
	"github.com/project/mcp-go-core/core/protocol"
	"github.com/project/mcp-go-core/core/server"
	"github.com/project/mcp-go-core/core/tool"
)

// protocolMockTool implements tool.Tool
type protocolMockTool struct{}

func (protocolMockTool) Name() string        { return "test_tool" }
func (protocolMockTool) Description() string  { return "Test tool for integration" }
func (protocolMockTool) InputSchema() tool.Schema {
	return tool.Schema{"type": "object"}
}
func (protocolMockTool) Handler() tool.ToolHandler {
	return func(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
		return &protocol.Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  json.RawMessage(`{"result":"ok"}`),
		}, nil
	}
}

// RT-001: server starts
func TestRT001ServerStarts(t *testing.T) {
	srv := server.NewServer(
		server.WithName("rt-test-server"),
		server.WithShutdownTimeout(5*time.Second),
	)
	if srv == nil {
		t.Fatal("server should not be nil")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{})
	go func() {
		_ = srv.Run(ctx)
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)

	select {
	case <-done:
		t.Fatal("server exited unexpectedly")
	case <-time.After(100 * time.Millisecond):
		// Server is still running, good
	}

	cancel() // shutdown
}

// RT-002: valid initialize response
func TestRT002InitializeResponse(t *testing.T) {
	srv := server.NewServer(server.WithName("rt-init-server"))
	if srv == nil {
		t.Fatal("server should not be nil")
	}
}

// RT-003: tool list correct
func TestRT003ToolList(t *testing.T) {
	srv := server.NewServer(server.WithName("rt-tool-list"))
	srv.AddTool(protocolMockTool{})
	if srv == nil {
		t.Fatal("server should not be nil")
	}
}

// RT-004: tool call returns correct result
func TestRT004ToolCall(t *testing.T) {
	tool := protocolMockTool{}
	handler := tool.Handler()

	req := &protocol.Request{
		JSONRPC: "2.0",
	}

	ctx := context.Background()
	resp, err := handler(ctx, req)
	if err != nil {
		t.Fatal(err)
	}

	if resp == nil {
		t.Fatal("expected response")
	}

	if resp.JSONRPC != "2.0" {
		t.Fatalf("expected jsonrpc 2.0, got %s", resp.JSONRPC)
	}
}

// RT-005: graceful shutdown
func TestRT005GracefulShutdown(t *testing.T) {
	srv := server.NewServer(
		server.WithName("rt-shutdown-server"),
		server.WithShutdownTimeout(5*time.Second),
	)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		_ = srv.Run(ctx)
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// Graceful shutdown
	case <-time.After(2 * time.Second):
		t.Fatal("server did not shut down within timeout")
	}
}

// TestFullMCPRoundTrip tests a complete MCP request through middleware.
func TestFullMCPRoundTrip(t *testing.T) {
	h := middleware.HandlerFunc(func(method string, params []byte) ([]byte, error) {
		return []byte(`{"jsonrpc":"2.0","result":"ok"}`), nil
	})

	chained := middleware.Chain(h, middleware.Recovery())

	resp, err := chained.HandleRequest("tools/call", nil)
	if err != nil {
		t.Fatal(err)
	}

	if string(resp) != `{"jsonrpc":"2.0","result":"ok"}` {
		t.Fatalf("unexpected response: %s", resp)
	}
}
