package tool

import (
	"context"
	"testing"

	"github.com/project/mcp-go-core/core/protocol"
)

func TestNewTool(t *testing.T) {
	tool := NewTool("greet", "say hello", Schema{"type": "object"}, func(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
		return &protocol.Response{JSONRPC: "2.0", Result: "hello"}, nil
	})

	if tool.Name() != "greet" {
		t.Fatal("name mismatch")
	}
	if tool.Description() != "say hello" {
		t.Fatal("description mismatch")
	}
	if tool.InputSchema()["type"] != "object" {
		t.Fatal("schema mismatch")
	}
}

func TestNewToolEmptyName(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for empty name")
		}
	}()
	NewTool("", "desc", nil, nil)
}
