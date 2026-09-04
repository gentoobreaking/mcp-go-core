// Package router provides tool/resource/prompt dispatch.
package router

import (
	"context"
	"encoding/json"

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
}

// NewRouter creates a new Router.
func NewRouter() *Router {
	return &Router{
		tools:     make(map[string]tool.Tool),
		resources: make(map[string]resource.Resource),
		prompts:   make(map[string]prompt.Prompt),
	}
}

// RegisterTool adds a tool to the router.
func (r *Router) RegisterTool(t tool.Tool) {
	r.tools[t.Name()] = t
}

// RegisterResource adds a resource to the router.
func (r *Router) RegisterResource(res resource.Resource) {
	r.resources[res.URI()] = res
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
	case "initialized":
		// Notification — no response needed
		return &protocol.Response{JSONRPC: "2.0", ID: req.ID, Result: nil}, nil
	default:
		return nil, mcperror.NewError(mcperror.CodeProtocol, "method not found: "+req.Method)
	}
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
			Resources: &protocol.ResourcesCapability{ListAvailable: true, Get: true},
			Prompts:   &protocol.PromptsCapability{ListAvailable: true, Get: true},
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
