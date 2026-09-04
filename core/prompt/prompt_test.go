package prompt

import (
	"context"
	"testing"
)

func TestNewPrompt(t *testing.T) {
	p := NewPrompt("explain", "explain something", func(ctx context.Context, req PromptRequest) (PromptResponse, error) {
		return PromptResponse{Messages: []PromptMessage{{Role: "user", Content: "hi"}}}, nil
	})

	if p.Name() != "explain" {
		t.Fatal("name mismatch")
	}
	if p.Description() != "explain something" {
		t.Fatal("description mismatch")
	}
}

func TestPromptGet(t *testing.T) {
	p := NewPrompt("explain", "", func(ctx context.Context, req PromptRequest) (PromptResponse, error) {
		return PromptResponse{Messages: []PromptMessage{{Role: "user", Content: req.Name}}}, nil
	})
	resp, err := p.Get(context.Background(), PromptRequest{Name: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Messages[0].Content != "hello" {
		t.Fatal("content mismatch")
	}
}

func TestPromptNameValidation(t *testing.T) {
	p := NewPrompt("", "desc", func(ctx context.Context, req PromptRequest) (PromptResponse, error) {
		return PromptResponse{}, nil
	})
	if p.Name() != "" {
		t.Fatal("expected empty name")
	}
}
