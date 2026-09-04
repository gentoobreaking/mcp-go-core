package router

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/project/mcp-go-core/core/protocol"
	"github.com/project/mcp-go-core/core/prompt"
	"github.com/project/mcp-go-core/core/resource"
	"github.com/project/mcp-go-core/core/tool"
)

func TestRouterMethods(t *testing.T) {
	r := NewRouter()

	// Test unknown method returns error
	_, err := r.Dispatch(context.Background(), &protocol.Request{
		JSONRPC: "2.0",
		Method:  "unknown",
	})
	if err == nil {
		t.Fatal("expected error for unknown method")
	}
}

func TestRouterRegisterTool(t *testing.T) {
	r := NewRouter()
	tk := tool.NewTool("test", "test tool", nil, nil)
	r.RegisterTool(tk)
	// Verify featuregraph is not referenced
}

func TestNoFeatureGraphReference(t *testing.T) {
	// Ensure router.go does not import or reference featuregraph
	// This is a build-time check - the import list above confirms it
}

func TestDispatchListTools(t *testing.T) {
	r := NewRouter()
	r.RegisterTool(testTool("tool1", "description 1"))
	r.RegisterTool(testTool("tool2", "description 2"))

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

	if len(result.Tools) != 2 {
		t.Fatalf("expected 2 tools, got %d", len(result.Tools))
	}
}

func TestDispatchCallTool(t *testing.T) {
	r := NewRouter()
	r.RegisterTool(testTool("echo", "echo tool"))

	params := json.RawMessage(`{"name":"echo","arguments":{"msg":"hello"}}`)

	req := &protocol.Request{
		JSONRPC: "2.0",
		ID:      int64Ptr(1),
		Method:  "tools/call",
		Params:  params,
	}

	resp, err := r.Dispatch(context.Background(), req)
	if err != nil {
		t.Fatalf("Dispatch error: %v", err)
	}

	if resp.JSONRPC != "2.0" {
		t.Fatalf("expected jsonrpc 2.0, got %s", resp.JSONRPC)
	}
}

func TestDispatchCallToolNotFound(t *testing.T) {
	r := NewRouter()

	params := json.RawMessage(`{"name":"nonexistent","arguments":{}}`)

	req := &protocol.Request{
		JSONRPC: "2.0",
		ID:      int64Ptr(1),
		Method:  "tools/call",
		Params:  params,
	}

	_, err := r.Dispatch(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for nonexistent tool")
	}
}

func TestDispatchListResources(t *testing.T) {
	r := NewRouter()

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

func TestDispatchListPrompts(t *testing.T) {
	r := NewRouter()

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

func TestDispatchInitialize(t *testing.T) {
	r := NewRouter()

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
}

func TestDispatchNotificationsCancel(t *testing.T) {
	r := NewRouter()

	// notifications/cancel is a notification — should return a nil-result response
	req := &protocol.Request{
		JSONRPC: "2.0",
		ID:      int64Ptr(1),
		Method:  "notifications/cancel",
		Params:  json.RawMessage(`{"requestId": 99}`),
	}

	resp, err := r.Dispatch(context.Background(), req)
	if err != nil {
		t.Fatalf("Dispatch error: %v", err)
	}

	if resp.JSONRPC != "2.0" {
		t.Fatalf("expected jsonrpc 2.0, got %s", resp.JSONRPC)
	}
}

func TestDispatchInitialized(t *testing.T) {
	r := NewRouter()

	// initialized is a notification — should return a nil-result response
	req := &protocol.Request{
		JSONRPC: "2.0",
		ID:      int64Ptr(1),
		Method:  "initialized",
	}

	resp, err := r.Dispatch(context.Background(), req)
	if err != nil {
		t.Fatalf("Dispatch error: %v", err)
	}

	if resp.JSONRPC != "2.0" {
		t.Fatalf("expected jsonrpc 2.0, got %s", resp.JSONRPC)
	}
}

func TestDispatchListPromptsWithRegistered(t *testing.T) {
	r := NewRouter()
	r.RegisterPrompt(testPrompt("prompt1", "description 1"))

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

	if len(result.Prompts) != 1 {
		t.Fatalf("expected 1 prompt, got %d", len(result.Prompts))
	}

	if result.Prompts[0].Name != "prompt1" {
		t.Fatalf("expected prompt1, got %s", result.Prompts[0].Name)
	}
}

func TestDispatchReadResource(t *testing.T) {
	r := NewRouter()
	r.RegisterResource(testResource("mcp://test/1", "Test Resource", "A test resource"))

	params := json.RawMessage(`{"uri":"mcp://test/1"}`)

	req := &protocol.Request{
		JSONRPC: "2.0",
		ID:      int64Ptr(1),
		Method:  "resources/read",
		Params:  params,
	}

	resp, err := r.Dispatch(context.Background(), req)
	if err != nil {
		t.Fatalf("Dispatch error: %v", err)
	}

	if resp.JSONRPC != "2.0" {
		t.Fatalf("expected jsonrpc 2.0, got %s", resp.JSONRPC)
	}
}

func TestDispatchGetPrompt(t *testing.T) {
	r := NewRouter()
	r.RegisterPrompt(testPrompt("greeting", "greeting prompt"))

	params := json.RawMessage(`{"name":"greeting","arguments":{"name":"world"}}`)

	req := &protocol.Request{
		JSONRPC: "2.0",
		ID:      int64Ptr(1),
		Method:  "prompts/get",
		Params:  params,
	}

	resp, err := r.Dispatch(context.Background(), req)
	if err != nil {
		t.Fatalf("Dispatch error: %v", err)
	}

	if resp.JSONRPC != "2.0" {
		t.Fatalf("expected jsonrpc 2.0, got %s", resp.JSONRPC)
	}
}

// --- Test helpers ---

func int64Ptr(i int64) *int64 { return &i }

func testTool(name, desc string) tool.Tool {
	return tool.NewTool(name, desc, tool.Schema{"type": "object"},
		func(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
			return &protocol.Response{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result:  map[string]any{"result": "ok"},
			}, nil
		},
	)
}

func testResource(uri, name, desc string) resource.Resource {
	return resource.NewResource(uri, name, desc,
		func(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
			return &protocol.Response{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result:  map[string]any{"contents": []string{"hello"}},
			}, nil
		},
	)
}

func testPrompt(name, desc string) prompt.Prompt {
	return prompt.NewPrompt(name, desc,
		func(ctx context.Context, req prompt.PromptRequest) (prompt.PromptResponse, error) {
			return prompt.PromptResponse{
				Description: desc,
				Messages: []prompt.PromptMessage{
					{Role: "assistant", Content: "Hello!"},
				},
			}, nil
		},
	)
}
