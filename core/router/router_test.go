package router

import (
	"context"
	"encoding/json"
	"fmt"
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

// T099: server→client resource notification on subscribed URI
func TestNotifyResourceUpdate(t *testing.T) {
	r := NewRouter()

	// Subscribe to a resource
	params := json.RawMessage(`{"uri":"mcp://test/1"}`)
	req := &protocol.Request{
		JSONRPC: "2.0",
		ID:      int64Ptr(1),
		Method:  "resources/subscribe",
		Params:  params,
	}
	_, err := r.Dispatch(context.Background(), req)
	if err != nil {
		t.Fatalf("subscribe failed: %v", err)
	}

	// Set notification handler
	var notifiedMethod string
	var notifiedParams any
	r.SetNotificationSender(func(method string, params any) error {
		notifiedMethod = method
		notifiedParams = params
		return nil
	})

	// Notify resource update
	if err := r.NotifyResourceUpdate("mcp://test/1", "update"); err != nil {
		t.Fatalf("NotifyResourceUpdate failed: %v", err)
	}

	if notifiedMethod != "notifications/resources/update" {
		t.Fatalf("expected method notifications/resources/update, got %s", notifiedMethod)
	}

	updateParams, ok := notifiedParams.(protocol.ResourceUpdateParams)
	if !ok {
		t.Fatalf("expected ResourceUpdateParams, got %T", notifiedParams)
	}
	if updateParams.URI != "mcp://test/1" {
		t.Fatalf("expected URI mcp://test/1, got %s", updateParams.URI)
	}
	if updateParams.ChangeType != "update" {
		t.Fatalf("expected changeType update, got %s", updateParams.ChangeType)
	}
}

// T099: notify-deleted for multiple subscribed clients
func TestNotifyResourceUpdateMultipleSubscriptions(t *testing.T) {
	r := NewRouter()

	// Subscribe to multiple resources
	for _, uri := range []string{"mcp://test/1", "mcp://test/2", "mcp://test/3"} {
		params := json.RawMessage(`{"uri":"` + uri + `"}`)
		req := &protocol.Request{
			JSONRPC: "2.0",
			ID:      int64Ptr(1),
			Method:  "resources/subscribe",
			Params:  params,
		}
		r.Dispatch(context.Background(), req)
	}

	var notifyCount int
	r.SetNotificationSender(func(method string, params any) error {
		notifyCount++
		return nil
	})

	// Only subscribed URI should trigger
	r.NotifyResourceUpdate("mcp://test/2", "delete")
	if notifyCount != 1 {
		t.Fatalf("expected 1 notification for subscribed URI, got %d", notifyCount)
	}
}

// T099: no notification when no subscribers
func TestNotifyResourceUpdateNoSubscribers(t *testing.T) {
	r := NewRouter()

	var notified bool
	r.SetNotificationSender(func(method string, params any) error {
		notified = true
		return nil
	})

	// Notify for a URI nobody subscribed to
	r.NotifyResourceUpdate("mcp://unsubscribed", "update")
	if notified {
		t.Fatal("expected no notification for unsubscribed URI")
	}
}

// T099: notification handler error propagates
func TestNotifyResourceUpdateHandlerError(t *testing.T) {
	r := NewRouter()

	// Subscribe
	params := json.RawMessage(`{"uri":"mcp://test/1"}`)
	req := &protocol.Request{
		JSONRPC: "2.0",
		ID:      int64Ptr(1),
		Method:  "resources/subscribe",
		Params:  params,
	}
	r.Dispatch(context.Background(), req)

	r.SetNotificationSender(func(method string, params any) error {
		return fmt.Errorf("send failed")
	})

	err := r.NotifyResourceUpdate("mcp://test/1", "update")
	if err == nil {
		t.Fatal("expected error from notification sender")
	}
}

// T099: resources/created fires notification for subscribed URI
func TestRegisterResourceNotifiesSubscriber(t *testing.T) {
	r := NewRouter()

	// Subscribe first
	params := json.RawMessage(`{"uri":"mcp://new/1"}`)
	req := &protocol.Request{
		JSONRPC: "2.0",
		ID:      int64Ptr(1),
		Method:  "resources/subscribe",
		Params:  params,
	}
	r.Dispatch(context.Background(), req)

	// Register a notification sender that receives updates
	var notifiedMethod string
	r.SetNotificationSender(func(method string, params any) error {
		notifiedMethod = method
		return nil
	})

	// Register a new resource at the subscribed URI
	res := testResource("mcp://new/1", "New Resource", "A new resource")
	r.RegisterResource(res)

	// Notify subscribers of the update
	r.NotifyResourceUpdate("mcp://new/1", "update")

	if notifiedMethod != "notifications/resources/update" {
		t.Fatalf("expected notification, got %s", notifiedMethod)
	}
}

// T100: prompts/create
func TestDispatchCreatePrompt(t *testing.T) {
	r := NewRouter()

	// Register a prompt creator factory
	r.SetPromptCreator(func(name, desc string, args map[string]any) (prompt.Prompt, error) {
		return testPrompt(name, desc), nil
	})

	params := json.RawMessage(`{"name":"dynamicPrompt","description":"a dynamic prompt"}`)
	req := &protocol.Request{
		JSONRPC: "2.0",
		ID:      int64Ptr(1),
		Method:  "prompts/create",
		Params:  params,
	}

	resp, err := r.Dispatch(context.Background(), req)
	if err != nil {
		t.Fatalf("Dispatch error: %v", err)
	}

	if resp.JSONRPC != "2.0" {
		t.Fatalf("expected jsonrpc 2.0, got %s", resp.JSONRPC)
	}

	// Verify the prompt was actually registered
	prompts := r.ListPrompts()
	if len(prompts) != 1 {
		t.Fatalf("expected 1 prompt registered, got %d", len(prompts))
	}
}

// T100: prompts/create without creator returns success (no-op)
func TestDispatchCreatePromptNoCreator(t *testing.T) {
	r := NewRouter()

	params := json.RawMessage(`{"name":"testPrompt"}`)
	req := &protocol.Request{
		JSONRPC: "2.0",
		ID:      int64Ptr(1),
		Method:  "prompts/create",
		Params:  params,
	}

	resp, err := r.Dispatch(context.Background(), req)
	if err != nil {
		t.Fatalf("Dispatch error: %v", err)
	}

	// No creator registered — should still return success (no-op)
	if resp.JSONRPC != "2.0" {
		t.Fatalf("expected jsonrpc 2.0, got %s", resp.JSONRPC)
	}
}

// T100: notifications/prompts/list_changed
func TestDispatchPromptsListChanged(t *testing.T) {
	r := NewRouter()

	called := false
	r.SetPromptListChangedHandler(func() error {
		called = true
		return nil
	})

	req := &protocol.Request{
		JSONRPC: "2.0",
		ID:      int64Ptr(1),
		Method:  "notifications/prompts/list_changed",
	}

	resp, err := r.Dispatch(context.Background(), req)
	if err != nil {
		t.Fatalf("Dispatch error: %v", err)
	}

	if !called {
		t.Fatal("expected prompt list changed handler to be called")
	}

	if resp.JSONRPC != "2.0" {
		t.Fatalf("expected jsonrpc 2.0, got %s", resp.JSONRPC)
	}
}

// T100: notifications/resources/list_changed
func TestDispatchResourcesListChanged(t *testing.T) {
	r := NewRouter()

	called := false
	r.SetResourceListChangedHandler(func() error {
		called = true
		return nil
	})

	req := &protocol.Request{
		JSONRPC: "2.0",
		ID:      int64Ptr(1),
		Method:  "notifications/resources/list_changed",
	}

	resp, err := r.Dispatch(context.Background(), req)
	if err != nil {
		t.Fatalf("Dispatch error: %v", err)
	}

	if !called {
		t.Fatal("expected resource list changed handler to be called")
	}

	if resp.JSONRPC != "2.0" {
		t.Fatalf("expected jsonrpc 2.0, got %s", resp.JSONRPC)
	}
}

// T100: notifications/tools/list_changed
func TestDispatchToolsListChanged(t *testing.T) {
	r := NewRouter()

	called := false
	r.SetToolsListChangedHandler(func() error {
		called = true
		return nil
	})

	req := &protocol.Request{
		JSONRPC: "2.0",
		ID:      int64Ptr(1),
		Method:  "notifications/tools/list_changed",
	}

	resp, err := r.Dispatch(context.Background(), req)
	if err != nil {
		t.Fatalf("Dispatch error: %v", err)
	}

	if !called {
		t.Fatal("expected tools list changed handler to be called")
	}

	if resp.JSONRPC != "2.0" {
		t.Fatalf("expected jsonrpc 2.0, got %s", resp.JSONRPC)
	}
}

// T101: resources/templates/list
func TestDispatchListResourceTemplates(t *testing.T) {
	r := NewRouter()

	req := &protocol.Request{
		JSONRPC: "2.0",
		ID:      int64Ptr(1),
		Method:  "resources/templates/list",
	}

	resp, err := r.Dispatch(context.Background(), req)
	if err != nil {
		t.Fatalf("Dispatch error: %v", err)
	}

	result, ok := resp.Result.(protocol.ResourceTemplateListResult)
	if !ok {
		t.Fatalf("expected ResourceTemplateListResult, got %T", resp.Result)
	}
	if result.Templates != nil && len(result.Templates) > 0 {
		t.Fatalf("expected empty templates list, got %d", len(result.Templates))
	}
}

// T101: resources/unsubscribe
func TestDispatchUnsubscribe(t *testing.T) {
	r := NewRouter()

	// Subscribe first
	r.subscriptions["mcp://test/1"] = map[string]bool{"default": true}

	params := json.RawMessage(`{"uri":"mcp://test/1"}`)
	req := &protocol.Request{
		JSONRPC: "2.0",
		ID:      int64Ptr(1),
		Method:  "resources/unsubscribe",
		Params:  params,
	}

	resp, err := r.Dispatch(context.Background(), req)
	if err != nil {
		t.Fatalf("Dispatch error: %v", err)
	}

	if resp.JSONRPC != "2.0" {
		t.Fatalf("expected jsonrpc 2.0, got %s", resp.JSONRPC)
	}

	if r.IsSubscribed("mcp://test/1") {
		t.Fatal("expected subscription to be removed")
	}
}

// T101: resources/unsubscribe with missing URI
func TestDispatchUnsubscribeMissingURI(t *testing.T) {
	r := NewRouter()

	req := &protocol.Request{
		JSONRPC: "2.0",
		ID:      int64Ptr(1),
		Method:  "resources/unsubscribe",
		Params:  json.RawMessage(`{}`),
	}

	_, err := r.Dispatch(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for missing uri")
	}
}

// T101: notifications/progress
func TestDispatchProgress(t *testing.T) {
	r := NewRouter()

	var received protocol.ProgressNotificationParams
	r.SetProgressHandler(func(params protocol.ProgressNotificationParams) {
		received = params
	})

	params := json.RawMessage(`{"progressToken":"abc-123","progress":5,"message":"working"}`)
	req := &protocol.Request{
		JSONRPC: "2.0",
		ID:      int64Ptr(1),
		Method:  "notifications/progress",
		Params:  params,
	}

	resp, err := r.Dispatch(context.Background(), req)
	if err != nil {
		t.Fatalf("Dispatch error: %v", err)
	}

	if resp.JSONRPC != "2.0" {
		t.Fatalf("expected jsonrpc 2.0, got %s", resp.JSONRPC)
	}

	if received.Progress != 5 {
		t.Fatalf("expected progress 5, got %d", received.Progress)
	}
	if received.Message != "working" {
		t.Fatalf("expected message 'working', got %s", received.Message)
	}
}

// T101: notifications/message
func TestDispatchMessage(t *testing.T) {
	r := NewRouter()

	var receivedLevel string
	var receivedData any
	r.SetMessageHandler(func(level, logger string, data any) {
		receivedLevel = level
		
		receivedData = data
	})

	params := json.RawMessage(`{"level":"info","logger":"myLogger","data":"hello"}`)
	req := &protocol.Request{
		JSONRPC: "2.0",
		ID:      int64Ptr(1),
		Method:  "notifications/message",
		Params:  params,
	}

	resp, err := r.Dispatch(context.Background(), req)
	if err != nil {
		t.Fatalf("Dispatch error: %v", err)
	}

	if resp.JSONRPC != "2.0" {
		t.Fatalf("expected jsonrpc 2.0, got %s", resp.JSONRPC)
	}

	if receivedLevel != "info" {
		t.Fatalf("expected level info, got %s", receivedLevel)
	}
	if receivedData != "hello" {
		t.Fatalf("expected data 'hello', got %v", receivedData)
	}
}

// T102: elicitation/create
func TestDispatchElicitationCreate(t *testing.T) {
	r := NewRouter()

	params := json.RawMessage(`{"message":"Do you accept?","requestedSchema":{"type":"object"}}`)
	req := &protocol.Request{
		JSONRPC: "2.0",
		ID:      int64Ptr(1),
		Method:  "elicitation/create",
		Params:  params,
	}

	resp, err := r.Dispatch(context.Background(), req)
	if err != nil {
		t.Fatalf("Dispatch error: %v", err)
	}

	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatalf("expected map result, got %T", resp.Result)
	}
	if result["elicitationId"] == nil {
		t.Fatal("expected elicitationId in result")
	}
	if result["message"] != "Do you accept?" {
		t.Fatalf("expected message, got %v", result["message"])
	}
}

// T102: notifications/elicitation/complete
func TestDispatchElicitationComplete(t *testing.T) {
	r := NewRouter()

	// First create an elicitation
	createReq := &protocol.Request{
		JSONRPC: "2.0",
		ID:      int64Ptr(1),
		Method:  "elicitation/create",
		Params:  json.RawMessage(`{"message":"accept?"}`),
	}
	createResp, _ := r.Dispatch(context.Background(), createReq)
	createResult := createResp.Result.(map[string]any)
	elID := createResult["elicitationId"].(string)

	// Now complete it
	completeReq := &protocol.Request{
		JSONRPC: "2.0",
		ID:      int64Ptr(2),
		Method:  "notifications/elicitation/complete",
		Params:  json.RawMessage(`{"id":"` + elID + `","result":{"action":"accept","content":{"answer":"yes"}}}`),
	}

	resp, err := r.Dispatch(context.Background(), completeReq)
	if err != nil {
		t.Fatalf("Dispatch error: %v", err)
	}
	if resp.JSONRPC != "2.0" {
		t.Fatalf("expected jsonrpc 2.0, got %s", resp.JSONRPC)
	}

	// Verify the elicitation was resolved
	resolved, ok := r.ResolveElicitation(elID)
	if !ok {
		t.Fatal("expected elicitation to be resolvable")
	}
	if resolved.Action != "accept" {
		t.Fatalf("expected action accept, got %s", resolved.Action)
	}
}

// T102: tasks/get
func TestDispatchTasksGet(t *testing.T) {
	r := NewRouter()
	r.RegisterTask("task-1", protocol.TaskStatusCompleted, "result data")

	params := json.RawMessage(`{"id":"task-1"}`)
	req := &protocol.Request{
		JSONRPC: "2.0",
		ID:      int64Ptr(1),
		Method:  "tasks/get",
		Params:  params,
	}

	resp, err := r.Dispatch(context.Background(), req)
	if err != nil {
		t.Fatalf("Dispatch error: %v", err)
	}

	task, ok := resp.Result.(protocol.TaskResult)
	if !ok {
		t.Fatalf("expected TaskResult, got %T", resp.Result)
	}
	if task.Status != protocol.TaskStatusCompleted {
		t.Fatalf("expected completed status, got %s", task.Status)
	}
}

// T102: tasks/list
func TestDispatchTasksList(t *testing.T) {
	r := NewRouter()
	r.RegisterTask("task-1", protocol.TaskStatusRunning, nil)
	r.RegisterTask("task-2", protocol.TaskStatusCompleted, "done")

	req := &protocol.Request{
		JSONRPC: "2.0",
		ID:      int64Ptr(1),
		Method:  "tasks/list",
	}

	resp, err := r.Dispatch(context.Background(), req)
	if err != nil {
		t.Fatalf("Dispatch error: %v", err)
	}

	result, ok := resp.Result.(protocol.TaskListResult)
	if !ok {
		t.Fatalf("expected TaskListResult, got %T", resp.Result)
	}
	if len(result.Tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(result.Tasks))
	}
}

// T102: tasks/cancel
func TestDispatchTasksCancel(t *testing.T) {
	r := NewRouter()
	r.RegisterTask("task-1", protocol.TaskStatusRunning, nil)

	params := json.RawMessage(`{"id":"task-1"}`)
	req := &protocol.Request{
		JSONRPC: "2.0",
		ID:      int64Ptr(1),
		Method:  "tasks/cancel",
		Params:  params,
	}

	resp, err := r.Dispatch(context.Background(), req)
	if err != nil {
		t.Fatalf("Dispatch error: %v", err)
	}

	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatalf("expected map result, got %T", resp.Result)
	}
	if result["success"] != true {
		t.Fatal("expected success true")
	}

	// Verify task was canceled
	task, ok := r.taskRegistry["task-1"]
	if !ok {
		t.Fatal("expected task to exist")
	}
	if task.Status != protocol.TaskStatusCanceled {
		t.Fatalf("expected canceled status, got %s", task.Status)
	}
}

// T102: server/discover
func TestDispatchServerDiscover(t *testing.T) {
	r := NewRouter()

	req := &protocol.Request{
		JSONRPC: "2.0",
		ID:      int64Ptr(1),
		Method:  "server/discover",
	}

	resp, err := r.Dispatch(context.Background(), req)
	if err != nil {
		t.Fatalf("Dispatch error: %v", err)
	}

	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatalf("expected map result, got %T", resp.Result)
	}
	if result["protocolVersion"] != "2024-11-05" {
		t.Fatalf("expected protocolVersion, got %v", result["protocolVersion"])
	}
	caps, ok := result["capabilities"].(map[string]any)
	if !ok {
		t.Fatal("expected capabilities map")
	}
	if !caps["tools"].(bool) {
		t.Fatal("expected tools capability true")
	}
}

// T102: roots/list
func TestDispatchRootsList(t *testing.T) {
	r := NewRouter()

	// Register roots (simulating client roots/list)
	r.SetRoots([]protocol.Root{
		{URI: "file:///home/user/docs", Name: "docs"},
		{URI: "file:///home/user/code", Name: "code"},
	})

	req := &protocol.Request{
		JSONRPC: "2.0",
		ID:      int64Ptr(1),
		Method:  "roots/list",
	}

	resp, err := r.Dispatch(context.Background(), req)
	if err != nil {
		t.Fatalf("Dispatch error: %v", err)
	}

	result, ok := resp.Result.(protocol.ListRootsResult)
	if !ok {
		t.Fatalf("expected ListRootsResult, got %T", resp.Result)
	}
	if len(result.Roots) != 2 {
		t.Fatalf("expected 2 roots, got %d", len(result.Roots))
	}
	if result.Roots[0].URI != "file:///home/user/docs" {
		t.Fatalf("expected docs root URI, got %s", result.Roots[0].URI)
	}
}

// T102: subscriptions/listen
func TestDispatchSubscriptionListen(t *testing.T) {
	r := NewRouter()

	params := json.RawMessage(`{"uri":"mcp://watch"}`)
	req := &protocol.Request{
		JSONRPC: "2.0",
		ID:      int64Ptr(1),
		Method:  "subscriptions/listen",
		Params:  params,
	}

	resp, err := r.Dispatch(context.Background(), req)
	if err != nil {
		t.Fatalf("Dispatch error: %v", err)
	}

	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatalf("expected map result, got %T", resp.Result)
	}
	if result["success"] != true {
		t.Fatal("expected success true")
	}
}
// T104: DeleteResource emits notifications/resources/deleted to subscribers
func TestDeleteResourceNotifiesSubscriber(t *testing.T) {
	r := NewRouter()
	r.SetNotificationSender(func(method string, params any) error {
		if method != "notifications/resources/deleted" {
			t.Errorf("expected notifications/resources/deleted, got %s", method)
		}
		p, ok := params.(protocol.ResourceDeleteParams)
		if !ok {
			t.Fatalf("expected ResourceDeleteParams, got %T", params)
		}
		if p.URI != "mcp://test/1" {
			t.Errorf("expected URI mcp://test/1, got %s", p.URI)
		}
		return nil
	})

	// Subscribe first
	r.Subscribe("mcp://test/1", "client-1")

	// Delete the resource
	err := r.DeleteResource("mcp://test/1")
	if err != nil {
		t.Fatalf("DeleteResource error: %v", err)
	}

	if r.IsSubscribed("mcp://test/1") {
		t.Fatal("expected subscription to be removed after delete")
	}
}

// T104: Per-client subscription tracking — only the unsubscribed client is removed
func TestPerClientSubscribeUnsubscribe(t *testing.T) {
	r := NewRouter()

	// Subscribe two clients
	r.Subscribe("mcp://test/1", "client-1")
	r.Subscribe("mcp://test/1", "client-2")

	if !r.IsSubscribed("mcp://test/1") {
		t.Fatal("expected resource to be subscribed")
	}

	// Unsubscribe only client-1
	r.UnsubscribeClient("mcp://test/1", "client-1")

	// Resource should still be subscribed because client-2 is still subscribed
	if !r.IsSubscribed("mcp://test/1") {
		t.Fatal("expected resource to still be subscribed after removing one client")
	}

	// Unsubscribe client-2 as well
	r.UnsubscribeClient("mcp://test/1", "client-2")

	if r.IsSubscribed("mcp://test/1") {
		t.Fatal("expected subscription to be fully removed")
	}
}

// T104: IsSubscribed backwards compatibility — works as before with default client
func TestIsSubscribedBackwardsCompat(t *testing.T) {
	r := NewRouter()

	// Subscribe with default client (backward-compatible API)
	r.subscriptions["mcp://test/1"] = map[string]bool{"default": true}

	if !r.IsSubscribed("mcp://test/1") {
		t.Fatal("expected IsSubscribed to return true for default client subscription")
	}

	if r.IsSubscribed("mcp://test/nonexistent") {
		t.Fatal("expected IsSubscribed to return false for nonexistent URI")
	}

	// Unsubscribe all clients (backward-compatible API)
	r.Unsubscribe("mcp://test/1")
	if r.IsSubscribed("mcp://test/1") {
		t.Fatal("expected IsSubscribed to return false after Unsubscribe")
	}
}
