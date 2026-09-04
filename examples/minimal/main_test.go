package main

import (
	"context"
	"testing"
)

func TestServerAddTool(t *testing.T) {
	srv := NewServer()
	srv.AddTool(Tool{
		Name:        "greet",
		Description: "Returns a greeting message",
	}, func(ctx context.Context, req *Request) (any, *Error) {
		return map[string]any{"message": "Hello from mcp-go-core!"}, nil
	})

	if len(srv.tools) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(srv.tools))
	}
	if srv.tools["greet"].Name != "greet" {
		t.Fatal("tool name mismatch")
	}
}

func TestServerErrorForUnknownMethod(t *testing.T) {
	srv := NewServer()
	resp := srv.handle(context.Background(), &Request{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "unknown",
	})

	if resp.Error == nil {
		t.Fatal("expected error for unknown method")
	}
	if resp.Error.Code != -32601 {
		t.Fatalf("expected error code -32601, got %d", resp.Error.Code)
	}
}
