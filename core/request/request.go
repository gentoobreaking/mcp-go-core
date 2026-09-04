// Package request provides MCP request types for tools, resources, and prompts.
package request

import "encoding/json"

// ToolRequest is the request params for tools/call.
type ToolRequest struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments,omitempty"`
}

// ResourceRequest is the request params for resources/read.
type ResourceRequest struct {
	URI string `json:"uri"`
}

// PromptRequest is the request params for prompts/get.
type PromptRequest struct {
	Name string `json:"name"`
	Arguments json.RawMessage `json:"arguments,omitempty"`
}
