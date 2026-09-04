// Package prompt provides the Prompt interface for MCP prompts.
package prompt

import "context"

// PromptRequest represents request params for prompts/get.
type PromptRequest struct {
	Name      string
	Arguments map[string]any
}

// PromptResponse represents a prompt response.
type PromptResponse struct {
	Description string
	Messages    []PromptMessage
}

// PromptMessage is a single message in a prompt response.
type PromptMessage struct {
	Role    string // "user" or "assistant"
	Content string
}

// Prompt is the interface for an MCP prompt.
type Prompt interface {
	Name() string
	Description() string
	Get(ctx context.Context, req PromptRequest) (PromptResponse, error)
}

// BasePrompt provides a partial implementation of Prompt.
type BasePrompt struct {
	name        string
	description string
	getFunc     func(ctx context.Context, req PromptRequest) (PromptResponse, error)
}

// NewPrompt creates a new Prompt.
func NewPrompt(name, description string, getFn func(ctx context.Context, req PromptRequest) (PromptResponse, error)) Prompt {
	return &BasePrompt{
		name:        name,
		description: description,
		getFunc:     getFn,
	}
}

func (p *BasePrompt) Name() string         { return p.name }
func (p *BasePrompt) Description() string  { return p.description }
func (p *BasePrompt) Get(ctx context.Context, req PromptRequest) (PromptResponse, error) {
	return p.getFunc(ctx, req)
}
