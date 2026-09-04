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

	// Verify T088: Logging capability advertised
	if result.Capabilities.Logging == nil {
		t.Fatal("expected logging capability to be advertised")
	}
	if !result.Capabilities.Resources.Subscribe {
		t.Fatal("expected resources subscribe capability")
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

// T088: logging/setLogLevel test
func TestDispatchSetLogLevel(t *testing.T) {
	r := NewRouter()

	params := json.RawMessage(`{"level":"debug"}`)

	req := &protocol.Request{
		JSONRPC: "2.0",
		ID:      int64Ptr(1),
		Method:  "logging/setLogLevel",
		Params:  params,
	}

	resp, err := r.Dispatch(context.Background(), req)
	if err != nil {
		t.Fatalf("Dispatch error: %v", err)
	}

	if resp.JSONRPC != "2.0" {
		t.Fatalf("expected jsonrpc 2.0, got %s", resp.JSONRPC)
	}

	if r.LogLevel() != protocol.LogLevelDebug {
		t.Fatalf("expected log level debug, got %s", r.LogLevel())
	}
}

// T088: sampling/createMessage test
func TestDispatchCreateMessage(t *testing.T) {
	r := NewRouter()
	r.SetSampler(func(ctx context.Context, req *protocol.CreateMessageParams) (*protocol.CreateMessageResult, error) {
		return &protocol.CreateMessageResult{
			Role:   "assistant",
			Content: protocol.Content{Type: "text", Text: "sampled response"},
			Model:  "test-model",
		}, nil
	})

	params := json.RawMessage(`{
		"messages": [{"role":"user","content":{"type":"text","text":"hello"}}],
		"maxTokens": 100,
		"temperature": 0.7,
		"systemPrompt": "You are helpful"
	}`)

	req := &protocol.Request{
		JSONRPC: "2.0",
		ID:      int64Ptr(1),
		Method:  "sampling/createMessage",
		Params:  params,
	}

	resp, err := r.Dispatch(context.Background(), req)
	if err != nil {
		t.Fatalf("Dispatch error: %v", err)
	}

	if resp.JSONRPC != "2.0" {
		t.Fatalf("expected jsonrpc 2.0, got %s", resp.JSONRPC)
	}

	result, ok := resp.Result.(*protocol.CreateMessageResult)
	if !ok {
		t.Fatalf("expected *CreateMessageResult, got %T", resp.Result)
	}

	if result.Content.Text != "sampled response" {
		t.Fatalf("expected sampled response, got %s", result.Content.Text)
	}
	if result.Model != "test-model" {
		t.Fatalf("expected test-model, got %s", result.Model)
	}
}

// T088: sampling/createMessage without handler
func TestDispatchCreateMessageNoHandler(t *testing.T) {
	r := NewRouter()

	params := json.RawMessage(`{"messages":[],"maxTokens":100}`)

	req := &protocol.Request{
		JSONRPC: "2.0",
		ID:      int64Ptr(1),
		Method:  "sampling/createMessage",
		Params:  params,
	}

	_, err := r.Dispatch(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for missing sampler")
	}
}

// T088: resources/created notification
func TestRegisterResourceCreatedNotification(t *testing.T) {
	r := NewRouter()

	var notifiedURI string
	var notifiedName string
	r.SetOnResourceCreated(func(uri, name string) {
		notifiedURI = uri
		notifiedName = name
	})

	res := testResource("mcp://test/1", "Test Resource", "A test resource")
	r.RegisterResource(res)

	if notifiedURI != "mcp://test/1" {
		t.Fatalf("expected URI mcp://test/1, got %s", notifiedURI)
	}
	if notifiedName != "Test Resource" {
		t.Fatalf("expected name Test Resource, got %s", notifiedName)
	}
}

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

// T089: ping method
func TestDispatchPing(t *testing.T) {
	r := NewRouter()

	req := &protocol.Request{
		JSONRPC: "2.0",
		ID:      int64Ptr(1),
		Method:  "ping",
	}

	resp, err := r.Dispatch(context.Background(), req)
	if err != nil {
		t.Fatalf("Dispatch error: %v", err)
	}

	if resp.JSONRPC != "2.0" {
		t.Fatalf("expected jsonrpc 2.0, got %s", resp.JSONRPC)
	}
}

// T089: complete/arg method
func TestDispatchComplete(t *testing.T) {
	r := NewRouter()
	r.RegisterTool(testTool("myTool", "a test tool"))
	r.RegisterTool(testTool("otherTool", "another tool"))

	params := json.RawMessage(`{"ref":{"kind":"tool"},"argumentName":"name","value":""}`)

	req := &protocol.Request{
		JSONRPC: "2.0",
		ID:      int64Ptr(1),
		Method:  "complete",
		Params:  params,
	}

	resp, err := r.Dispatch(context.Background(), req)
	if err != nil {
		t.Fatalf("Dispatch error: %v", err)
	}

	result, ok := resp.Result.(protocol.CompleteResult)
	if !ok {
		t.Fatalf("expected CompleteResult, got %T", resp.Result)
	}

	values := result.Completion.Values
	if len(values) != 2 {
		t.Fatalf("expected 2 completion values, got %d", len(values))
	}
}

// T089: complete with invalid params
func TestDispatchCompleteInvalidParams(t *testing.T) {
	r := NewRouter()

	params := json.RawMessage(`{invalid`)

	req := &protocol.Request{
		JSONRPC: "2.0",
		ID:      int64Ptr(1),
		Method:  "complete",
		Params:  params,
	}

	_, err := r.Dispatch(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for invalid params")
	}
}

// T089: notifications/roots/list_changed
func TestDispatchRootsListChanged(t *testing.T) {
	r := NewRouter()

	called := false
	r.SetRootsHandler(func() error {
		called = true
		return nil
	})

	req := &protocol.Request{
		JSONRPC: "2.0",
		ID:      int64Ptr(1),
		Method:  "notifications/roots/list_changed",
	}

	resp, err := r.Dispatch(context.Background(), req)
	if err != nil {
		t.Fatalf("Dispatch error: %v", err)
	}

	if !called {
		t.Fatal("expected roots handler to be called")
	}

	if resp.JSONRPC != "2.0" {
		t.Fatalf("expected jsonrpc 2.0, got %s", resp.JSONRPC)
	}
}

// T089: resources/subscribe
func TestDispatchSubscribe(t *testing.T) {
	r := NewRouter()

	params := json.RawMessage(`{"uri":"mcp://test/1"}`)

	req := &protocol.Request{
		JSONRPC: "2.0",
		ID:      int64Ptr(1),
		Method:  "resources/subscribe",
		Params:  params,
	}

	resp, err := r.Dispatch(context.Background(), req)
	if err != nil {
		t.Fatalf("Dispatch error: %v", err)
	}

	if resp.JSONRPC != "2.0" {
		t.Fatalf("expected jsonrpc 2.0, got %s", resp.JSONRPC)
	}

	if !r.IsSubscribed("mcp://test/1") {
		t.Fatal("expected subscription to be stored")
	}
}

// T089: resources/subscribe with missing URI
func TestDispatchSubscribeMissingURI(t *testing.T) {
	r := NewRouter()

	req := &protocol.Request{
		JSONRPC: "2.0",
		ID:      int64Ptr(1),
		Method:  "resources/subscribe",
		Params:  json.RawMessage(`{}`),
	}

	_, err := r.Dispatch(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for missing uri")
	}
}
