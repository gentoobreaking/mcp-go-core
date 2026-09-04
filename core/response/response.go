// Package response provides MCP response types for tools, resources, and prompts.
package response

import (
	"encoding/json"

	"github.com/project/mcp-go-core/core/prompt"
)

// ToolResponse is the result of a tools/call.
type ToolResponse struct {
	Content []TypedContent `json:"content"`
}

// TypedContent represents a single content block in a tool response.
type TypedContent struct {
	Type     string `json:"type"`      // "text", "image", "resource"
	Text     string `json:"text,omitempty"`
	Data     string `json:"data,omitempty"`     // base64 encoded for images
	MIMEType string `json:"mimeType,omitempty"`
	URI      string `json:"uri,omitempty"`
}

// ResourceResponse is the result of resources/read.
type ResourceResponse struct {
	Contents []ResourceContent `json:"contents"`
}

// ResourceContent represents a resource returned by resources/read.
type ResourceContent struct {
	URI      string `json:"uri"`
	Text     string `json:"text,omitempty"`
	MIMEType string `json:"mimeType,omitempty"`
	Blob     string `json:"blob,omitempty"`
	JSON     json.RawMessage `json:"json,omitempty"`
}

// PromptResponse is the result of prompts/get.
// Aliased to prompt.PromptResponse for spec-compliant JSON serialization.

// PromptMsg is a single message in a prompt response.
// Deprecated: Use prompt.PromptMessage from core/prompt package.
type PromptMsg = prompt.PromptMessage

// PromptResponse aliases prompt.PromptResponse.
type PromptResponse = prompt.PromptResponse
