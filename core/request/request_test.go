package request

import "testing"

func TestToolRequest(t *testing.T) {
	req := ToolRequest{Name: "greet", Arguments: nil}
	if req.Name != "greet" {
		t.Fatal("name mismatch")
	}
}

func TestResourceRequest(t *testing.T) {
	req := ResourceRequest{URI: "mcp://test"}
	if req.URI != "mcp://test" {
		t.Fatal("URI mismatch")
	}
}

func TestPromptRequest(t *testing.T) {
	req := PromptRequest{Name: "explain", Arguments: nil}
	if req.Name != "explain" {
		t.Fatal("name mismatch")
	}
}
