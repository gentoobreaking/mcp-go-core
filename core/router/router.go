// Package router provides tool/resource/prompt dispatch.
package router

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/project/mcp-go-core/core/mcperror"
	"github.com/project/mcp-go-core/core/protocol"
	"github.com/project/mcp-go-core/core/prompt"
	"github.com/project/mcp-go-core/core/resource"
	"github.com/project/mcp-go-core/core/tool"
)

// Router dispatches MCP methods to registered handlers.
type Router struct {
	tools     map[string]tool.Tool
	resources map[string]resource.Resource
	prompts   map[string]prompt.Prompt
	logLevel  protocol.LogLevel
	sampler   SamplingHandler
	onResourceCreated CreatedNotifier
	subscriptions map[string]bool
	rootsHandler    func() error
}

// SamplingHandler is called when the client sends sampling/createMessage.
type SamplingHandler func(ctx context.Context, req *protocol.CreateMessageParams) (*protocol.CreateMessageResult, error)

// CreatedNotifier sends a resources/created notification when a resource is registered.
type CreatedNotifier func(uri, name string)

// NewRouter creates a new Router.
func NewRouter() *Router {
	return &Router{
		tools:     make(map[string]tool.Tool),
		resources: make(map[string]resource.Resource),
		prompts:   make(map[string]prompt.Prompt),
		logLevel:  protocol.LogLevelInfo,
		subscriptions: make(map[string]bool),
	}
}

// SetLogLevel sets the server log level.
func (r *Router) SetLogLevel(level protocol.LogLevel) {
	r.logLevel = level
}

// LogLevel returns the current log level.
func (r *Router) LogLevel() protocol.LogLevel {
	return r.logLevel
}

// SetSampler registers a handler for sampling/createMessage.
func (r *Router) SetSampler(h SamplingHandler) {
	r.sampler = h
}

// SetOnResourceCreated registers a notification callback for resources/created.
func (r *Router) SetOnResourceCreated(fn CreatedNotifier) {
	r.onResourceCreated = fn
}

// RegisterTool adds a tool to the router.
func (r *Router) RegisterTool(t tool.Tool) {
	r.tools[t.Name()] = t
}

// RegisterResource adds a resource to the router and emits a
// notifications/resources/created notification to any registered handler.
func (r *Router) RegisterResource(res resource.Resource) {
	r.resources[res.URI()] = res
	if r.onResourceCreated != nil {
		r.onResourceCreated(res.URI(), res.Name())
	}
}

// RegisterPrompt adds a prompt to the router.
func (r *Router) RegisterPrompt(p prompt.Prompt) {
	r.prompts[p.Name()] = p
}

// Dispatch routes a request to the appropriate handler.
func (r *Router) Dispatch(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	switch req.Method {
	case "initialize":
		return r.handleInitialize(ctx, req)
	case "tools/list":
		return r.dispatchListTools(ctx, req)
	case "tools/call":
		return r.dispatchCallTool(ctx, req)
	case "resources/list":
		return r.dispatchListResources(ctx, req)
	case "resources/read":
		return r.dispatchReadResource(ctx, req)
	case "prompts/list":
		return r.dispatchListPrompts(ctx, req)
	case "prompts/get":
		return r.dispatchGetPrompt(ctx, req)
	case "logging/setLogLevel":
		return r.dispatchSetLogLevel(ctx, req)
	case "sampling/createMessage":
		return r.dispatchCreateMessage(ctx, req)
	case "notifications/cancel":
		// Notification — client requesting cancellation of in-flight request
		return &protocol.Response{JSONRPC: "2.0", ID: req.ID, Result: nil}, nil
	case "initialized":
		// Notification — no response needed
		return &protocol.Response{JSONRPC: "2.0", ID: req.ID, Result: nil}, nil
	case "ping":
		return r.dispatchPing(ctx, req)
	case "complete":
		return r.dispatchComplete(ctx, req)
	case "notifications/roots/list_changed":
		return r.dispatchRootsListChanged(ctx, req)
	case "resources/subscribe":
		return r.dispatchSubscribe(ctx, req)
	default:
		return nil, mcperror.NewError(mcperror.CodeProtocol, "method not found: "+req.Method)
	}
}


// dispatchPing handles the ping method — returns "pong".
func (r *Router) dispatchPing(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	return &protocol.Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  protocol.PingResult{},
	}, nil
}

// dispatchComplete handles complete/arg and complete/prompt requests.
func (r *Router) dispatchComplete(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	var params protocol.CompletionParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return nil, mcperror.NewInvalidParamsError("invalid complete params: " + err.Error())
	}

	// Provide completion values based on context
	values := r.completeValues(params)

	return &protocol.Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: protocol.CompleteResult{
			Completion: protocol.CompletionResult{
				Values: values,
			},
		},
	}, nil
}

// completeValues returns completion candidates for the given params.
func (r *Router) completeValues(params protocol.CompletionParams) []string {
	// Argument value "true"/"false" completion for booleans
	if params.Value == "" || strings.HasPrefix(strings.ToLower(params.Value), "true") || strings.HasPrefix(strings.ToLower(params.Value), "false") {
		if params.ArgumentName == "enabled" || params.ArgumentName == "verbose" {
			return []string{"true", "false"}
		}
	}

	// Tool/prompt reference completion
	if params.Ref.Kind == "tool" || params.Ref.Kind == "prompt" {
		var names []string
		if params.Ref.Kind == "tool" {
			for name := range r.tools {
				names = append(names, name)
			}
		} else {
			for name := range r.prompts {
				names = append(names, name)
			}
		}
		return names
	}

	return nil
}

// dispatchRootsListChanged handles notifications/roots/list_changed.
func (r *Router) dispatchRootsListChanged(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	// Trigger roots refresh hook if registered
	if r.rootsHandler != nil {
		_ = r.rootsHandler()
	}
	// Notification — no response body needed, but return JSON-RPC ack
	return &protocol.Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  nil,
	}, nil
}

// dispatchSubscribe handles resources/subscribe requests.
func (r *Router) dispatchSubscribe(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	var params protocol.SubscribeParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return nil, mcperror.NewInvalidParamsError("invalid subscribe params: " + err.Error())
	}

	if params.URI == "" {
		return nil, mcperror.NewInvalidParamsError("uri is required")
	}

	r.subscriptions[params.URI] = true

	return &protocol.Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: struct{}{},
	}, nil
}

// SetRootsHandler registers a callback invoked when the client sends
// notifications/roots/list_changed.
func (r *Router) SetRootsHandler(handler func() error) {
	r.rootsHandler = handler
}

// IsSubscribed returns whether a resource URI has an active subscription.
func (r *Router) IsSubscribed(uri string) bool {
	return r.subscriptions[uri]
}

// handleInitialize processes the initialize request handshake.
func (r *Router) handleInitialize(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	var initReq protocol.InitializeRequest
	if err := json.Unmarshal(req.Params, &initReq); err != nil {
		return nil, mcperror.NewInvalidParamsError("invalid initialize params: " + err.Error())
	}

	result := protocol.InitializeResult{
		ProtocolVersion: "2024-11-05",
		Capabilities: protocol.ServerCapabilities{
			Tools:     &protocol.ToolsCapability{ListAvailable: true, Call: true},
			Resources: &protocol.ResourcesCapability{ListAvailable: true, Get: true, Subscribe: true},
			Prompts:   &protocol.PromptsCapability{ListAvailable: true, Get: true},
			Logging:   &protocol.LoggingCapability{Log: true},
		},
		ServerInfo: protocol.Implementation{
			Name:    "mcp-go-core",
			Version: "0.1.0",
		},
	}

	_ = initReq // client info available if needed for future extensions

	return &protocol.Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}, nil
}

// dispatchListTools returns all registered tools.
func (r *Router) dispatchListTools(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	tools := make([]protocol.Tool, 0, len(r.tools))
	for _, t := range r.tools {
		tools = append(tools, protocol.Tool{
			Name:        t.Name(),
			Description: t.Description(),
			InputSchema: t.InputSchema(),
		})
	}
	return &protocol.Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  protocol.ToolListResult{Tools: tools},
	}, nil
}

// dispatchCallTool invokes a registered tool by name.
func (r *Router) dispatchCallTool(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	var callReq struct {
		Name      string         `json:"name"`
		Arguments map[string]any `json:"arguments"`
	}
	if err := json.Unmarshal(req.Params, &callReq); err != nil {
		return nil, mcperror.NewInvalidParamsError("invalid tool call params: " + err.Error())
	}

	t, ok := r.tools[callReq.Name]
	if !ok {
		return nil, mcperror.NewError(mcperror.CodeTool, "tool not found: "+callReq.Name)
	}

	return t.Handler()(ctx, req)
}

// dispatchListResources returns all registered resources.
func (r *Router) dispatchListResources(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	resources := make([]protocol.Resource, 0, len(r.resources))
	for _, res := range r.resources {
		resources = append(resources, protocol.Resource{
			URI:         res.URI(),
			Name:        res.Name(),
			Description: res.Description(),
		})
	}
	return &protocol.Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  protocol.ResourceListResult{Resources: resources},
	}, nil
}

// dispatchReadResource reads a resource by URI.
func (r *Router) dispatchReadResource(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	var readReq struct {
		URI string `json:"uri"`
	}
	if err := json.Unmarshal(req.Params, &readReq); err != nil {
		return nil, mcperror.NewInvalidParamsError("invalid resource read params: " + err.Error())
	}

	res, ok := r.resources[readReq.URI]
	if !ok {
		return nil, mcperror.NewError(mcperror.CodeProtocol, "resource not found: "+readReq.URI)
	}

	return res.Read(ctx, req)
}

// dispatchListPrompts returns all registered prompts.
func (r *Router) dispatchListPrompts(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	prompts := make([]protocol.Prompt, 0, len(r.prompts))
	for _, p := range r.prompts {
		prompts = append(prompts, protocol.Prompt{
			Name:        p.Name(),
			Description: p.Description(),
		})
	}
	return &protocol.Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  protocol.PromptListResult{Prompts: prompts},
	}, nil
}

// dispatchGetPrompt retrieves a prompt by name.
func (r *Router) dispatchGetPrompt(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	var getReq struct {
		Name      string         `json:"name"`
		Arguments map[string]any `json:"arguments"`
	}
	if err := json.Unmarshal(req.Params, &getReq); err != nil {
		return nil, mcperror.NewInvalidParamsError("invalid prompt get params: " + err.Error())
	}

	p, ok := r.prompts[getReq.Name]
	if !ok {
		return nil, mcperror.NewError(mcperror.CodeProtocol, "prompt not found: "+getReq.Name)
	}

	resp, err := p.Get(ctx, prompt.PromptRequest{
		Name:      getReq.Name,
		Arguments: getReq.Arguments,
	})
	if err != nil {
		return nil, err
	}

	return &protocol.Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  resp,
}, nil
}

// dispatchSetLogLevel sets the server log level.
func (r *Router) dispatchSetLogLevel(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	var params protocol.SetLogLevelParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return nil, mcperror.NewInvalidParamsError("invalid setLogLevel params: " + err.Error())
	}

	r.SetLogLevel(params.Level)

	return &protocol.Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  nil,
	}, nil
}

// dispatchCreateMessage handles sampling/createMessage by delegating to the
// registered SamplingHandler.
func (r *Router) dispatchCreateMessage(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	if r.sampler == nil {
		return nil, mcperror.NewError(mcperror.CodeProtocol, "sampling not supported")
	}

	var params protocol.CreateMessageParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return nil, mcperror.NewInvalidParamsError("invalid createMessage params: " + err.Error())
	}

	result, err := r.sampler(ctx, &params)
	if err != nil {
		return nil, err
	}

	return &protocol.Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}, nil
}
