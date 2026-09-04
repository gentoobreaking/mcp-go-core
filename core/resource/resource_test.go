package resource

import (
	"context"
	"testing"

	"github.com/project/mcp-go-core/core/protocol"
)

func TestNewResource(t *testing.T) {
	r := NewResource("mcp://test", "test", "a test resource", func(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
		return &protocol.Response{JSONRPC: "2.0", Result: "content"}, nil
	})

	if r.URI() != "mcp://test" {
		t.Fatal("URI mismatch")
	}
	if r.Name() != "test" {
		t.Fatal("name mismatch")
	}
}

func TestResourceRead(t *testing.T) {
	r := NewResource("mcp://test", "test", "test", func(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
		return &protocol.Response{JSONRPC: "2.0", Result: "content"}, nil
	})
	resp, err := r.Read(context.Background(), &protocol.Request{JSONRPC: "2.0", Method: "resources/read"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Result != "content" {
		t.Fatal("unexpected result")
	}
}
