// Package router provides tool/resource/prompt dispatch.
package router

import (
	"context"

	"github.com/project/mcp-go-core/core/mcperror"
	"github.com/project/mcp-go-core/core/protocol"
	"github.com/project/mcp-go-core/core/prompt"
	"github.com/project/mcp-go-core/core/resource"
	"github.com/project/mcp-go-core/core/tool"
)

// Router dispatches MCP methods to registered handlers.
type Router struct {
	tools    map[string]tool.Tool
	resources map[string]resource.Resource
	prompts  map[string]prompt.Prompt
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
	case "tools/call":
		return r.dispatchTool(ctx, req)
	case "resources/read":
		return r.dispatchResource(ctx, req)
	case "prompts/get":
		return r.dispatchPrompt(ctx, req)
	default:
		return nil, mcperror.NewError(mcperror.CodeProtocol, "method not found: "+req.Method)
	}
}

func (r *Router) dispatchTool(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	// Extract tool name from params - simplified for now
	// Full implementation would parse ToolRequest
	return nil, mcperror.NewError(mcperror.CodeValidation, "tool dispatch not fully implemented")
}

func (r *Router) dispatchResource(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	return nil, mcperror.NewError(mcperror.CodeValidation, "resource dispatch not fully implemented")
}

func (r *Router) dispatchPrompt(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	return nil, mcperror.NewError(mcperror.CodeValidation, "prompt dispatch not fully implemented")
}
