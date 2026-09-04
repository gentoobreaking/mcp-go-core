package smoke

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/project/mcp-go-core/core/middleware"
	"github.com/project/mcp-go-core/core/protocol"
	"github.com/project/mcp-go-core/core/router"
	"github.com/project/mcp-go-core/core/server"
	"github.com/project/mcp-go-core/core/tool"
	"github.com/project/mcp-go-core/testutil"
)

// protocolMockTool implements tool.Tool
type protocolMockTool struct{}

func (protocolMockTool) Name() string         { return "test_tool" }
func (protocolMockTool) Description() string  { return "Test tool for integration" }
func (protocolMockTool) InputSchema() tool.Schema {
	return tool.Schema{"type": "object"}
}

func (protocolMockTool) Handler() tool.ToolHandler {
	return func(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
		return &protocol.Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  map[string]any{"result": "my-result"},
		}, nil
	}
}

func int64Ptr(i int64) *int64 {
	return &i
}

// RT-001: server starts and can be shut down gracefully
func TestRT001ServerStarts(t *testing.T) {
	srv := server.NewBuilder().
		WithName("rt-test-server").
		WithTransport(testutil.NewMockTransport()).
		WithTimeout(5 * time.Second).
		MustBuild()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	err := srv.Run(ctx)
	if err != nil && err != context.Canceled {
		t.Fatalf("unexpected error: %v", err)
	}
}

// RT-002: valid initialize response
func TestRT002InitializeResponse(t *testing.T) {
	r := router.NewRouter()

	req := &protocol.Request{
		JSONRPC: "2.0",
		ID:      int64Ptr(1),
		Method:  "initialize",
		Params: json.RawMessage(`{
			"protocolVersion": "2024-11-05",
			"capabilities": {},
			"clientInfo": {"name": "test-client", "version": "1.0.0"}
		}`),
	}

	resp, err := r.Dispatch(context.Background(), req)
	if err != nil {
		t.Fatalf("Dispatch error: %v", err)
	}

	if resp.JSONRPC != "2.0" {
		t.Fatalf("expected jsonrpc 2.0, got %s", resp.JSONRPC)
	}

	result, ok := resp.Result.(protocol.InitializeResult)
	if !ok {
		t.Fatalf("expected InitializeResult, got %T", resp.Result)
	}

	if result.ProtocolVersion != "2024-11-05" {
		t.Fatalf("expected protocol version 2024-11-05, got %s", result.ProtocolVersion)
	}

	if result.ServerInfo.Name != "mcp-go-core" {
		t.Fatalf("expected server name mcp-go-core, got %s", result.ServerInfo.Name)
	}

	if result.ServerInfo.Version != "0.1.0" {
		t.Fatalf("expected server version 0.1.0, got %s", result.ServerInfo.Version)
	}

	if result.Capabilities.Tools == nil || !result.Capabilities.Tools.ListAvailable {
		t.Fatal("expected tools capability with listAvailable")
	}

	if result.Capabilities.Resources == nil || !result.Capabilities.Resources.ListAvailable {
		t.Fatal("expected resources capability with listAvailable")
	}

	if result.Capabilities.Prompts == nil || !result.Capabilities.Prompts.ListAvailable {
		t.Fatal("expected prompts capability with listAvailable")
	}
}

// RT-003: tool list correct
func TestRT003ToolList(t *testing.T) {
	r := router.NewRouter()
	r.RegisterTool(protocolMockTool{})

	req := &protocol.Request{
		JSONRPC: "2.0",
		ID:      int64Ptr(1),
		Method:  "tools/list",
	}

	resp, err := r.Dispatch(context.Background(), req)
	if err != nil {
		t.Fatalf("Dispatch error: %v", err)
	}

	result, ok := resp.Result.(protocol.ToolListResult)
	if !ok {
		t.Fatalf("expected ToolListResult, got %T", resp.Result)
	}

	if len(result.Tools) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(result.Tools))
	}

	if result.Tools[0].Name != "test_tool" {
		t.Fatalf("expected tool name test_tool, got %s", result.Tools[0].Name)
	}
}

// RT-004: tool call returns correct result
func TestRT004ToolCall(t *testing.T) {
	r := router.NewRouter()
	r.RegisterTool(protocolMockTool{})

	req := &protocol.Request{
		JSONRPC: "2.0",
		ID:      int64Ptr(1),
		Method:  "tools/call",
		Params:  json.RawMessage(`{"name":"test_tool","arguments":{}}`),
	}

	resp, err := r.Dispatch(context.Background(), req)
	if err != nil {
		t.Fatalf("Dispatch error: %v", err)
	}

	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatalf("expected map result, got %T", resp.Result)
	}

	if result["result"] != "my-result" {
		t.Fatalf("expected result 'my-result', got %v", result["result"])
	}
}

// RT-005: graceful shutdown
func TestRT005GracefulShutdown(t *testing.T) {
	srv := server.NewBuilder().
		WithName("rt-shutdown-server").
		WithTransport(testutil.NewMockTransport()).
		WithTimeout(5 * time.Second).
		MustBuild()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	err := srv.Run(ctx)
	if err != nil && err != context.Canceled {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestFullMCPRoundTrip tests routing through MockTransport + router.
func TestFullMCPRoundTrip(t *testing.T) {
	mt := testutil.NewMockTransport()
	r := router.NewRouter()
	r.RegisterTool(protocolMockTool{})

	mt.Intercept(func(msg json.RawMessage) (any, error) {
		var req protocol.Request
		if err := json.Unmarshal(msg, &req); err != nil {
			return nil, err
		}
		resp, err := r.Dispatch(context.Background(), &req)
		if err != nil {
			return nil, err
		}
		return resp.Result, nil
	})

	ctx := context.Background()

	msg := json.RawMessage(`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"test_tool","arguments":{}}}`)

	resp, err := mt.Send(ctx, msg)
	if err != nil {
		t.Fatalf("Send error: %v", err)
	}

	result, ok := resp.(map[string]any)
	if !ok {
		t.Fatalf("expected map result, got %T", resp)
	}

	if result["result"] != "my-result" {
		t.Fatalf("expected 'my-result', got %v", result["result"])
	}
}

// TestResourcesList tests listing registered resources returns empty list.
func TestResourcesList(t *testing.T) {
	r := router.NewRouter()

	req := &protocol.Request{
		JSONRPC: "2.0",
		ID:      int64Ptr(1),
		Method:  "resources/list",
	}

	resp, err := r.Dispatch(context.Background(), req)
	if err != nil {
		t.Fatalf("Dispatch error: %v", err)
	}

	result, ok := resp.Result.(protocol.ResourceListResult)
	if !ok {
		t.Fatalf("expected ResourceListResult, got %T", resp.Result)
	}

	if len(result.Resources) != 0 {
		t.Fatalf("expected 0 resources, got %d", len(result.Resources))
	}
}

// TestPromptsList tests listing registered prompts returns empty list.
func TestPromptsList(t *testing.T) {
	r := router.NewRouter()

	req := &protocol.Request{
		JSONRPC: "2.0",
		ID:      int64Ptr(1),
		Method:  "prompts/list",
	}

	resp, err := r.Dispatch(context.Background(), req)
	if err != nil {
		t.Fatalf("Dispatch error: %v", err)
	}

	result, ok := resp.Result.(protocol.PromptListResult)
	if !ok {
		t.Fatalf("expected PromptListResult, got %T", resp.Result)
	}

	if len(result.Prompts) != 0 {
		t.Fatalf("expected 0 prompts, got %d", len(result.Prompts))
	}
}

// TestMethodNotFound tests unknown method handling.
func TestMethodNotFound(t *testing.T) {
	r := router.NewRouter()

	req := &protocol.Request{
		JSONRPC: "2.0",
		ID:      int64Ptr(1),
		Method:  "unknown/method",
	}

	_, err := r.Dispatch(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for unknown method")
	}
}

// TestMiddlewareChain tests middleware + router integration.
func TestMiddlewareChain(t *testing.T) {
	r := router.NewRouter()
	r.RegisterTool(protocolMockTool{})

	mwHandler := middleware.HandlerFunc(func(method string, params []byte) ([]byte, error) {
		resp, err := r.Dispatch(context.Background(), &protocol.Request{
			JSONRPC: "2.0",
			ID:      int64Ptr(1),
			Method:  method,
			Params:  json.RawMessage(params),
		})
		if err != nil {
			return nil, err
		}
		return json.Marshal(resp.Result)
	})

	chained := middleware.Chain(mwHandler, middleware.Recovery())

	resp, err := chained.HandleRequest("tools/call", json.RawMessage(`{"name":"test_tool","arguments":{}}`))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(resp) == 0 {
		t.Fatal("expected non-empty response")
	}
}
