// Package tool provides the Tool interface for MCP tools.
package tool

import (
	"context"

	"github.com/project/mcp-go-core/core/protocol"
)

// Schema represents a JSON Schema for tool input.
type Schema map[string]any

// ToolHandler is the function signature for handling a tool call.
type ToolHandler func(ctx context.Context, req *protocol.Request) (*protocol.Response, error)

// Tool is the interface for an MCP tool.
type Tool interface {
	Name() string
	Description() string
	InputSchema() Schema
	Handler() ToolHandler
}

// BaseTool provides a partial implementation of Tool.
type BaseTool struct {
	name        string
	description string
	schema      Schema
	handler     ToolHandler
}

// NewTool creates a new Tool.
func NewTool(name, description string, schema Schema, handler ToolHandler) Tool {
	if name == "" {
		panic("tool name cannot be empty")
	}
	return &BaseTool{
		name:        name,
		description: description,
		schema:      schema,
		handler:     handler,
	}
}

func (t *BaseTool) Name() string         { return t.name }
func (t *BaseTool) Description() string  { return t.description }
func (t *BaseTool) InputSchema() Schema  { return t.schema }
func (t *BaseTool) Handler() ToolHandler { return t.handler }
