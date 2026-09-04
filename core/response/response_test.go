package response

import "testing"

func TestToolResponse(t *testing.T) {
	resp := ToolResponse{
		Content: []TypedContent{
			{Type: "text", Text: "Hello"},
		},
	}
	if len(resp.Content) != 1 {
		t.Fatal("expected 1 content")
	}
	if resp.Content[0].Text != "Hello" {
		t.Fatal("text mismatch")
	}
}

func TestResourceResponse(t *testing.T) {
	resp := ResourceResponse{
		Contents: []ResourceContent{
			{URI: "mcp://test", Text: "content"},
		},
	}
	if resp.Contents[0].URI != "mcp://test" {
		t.Fatal("URI mismatch")
	}
}

func TestPromptResponse(t *testing.T) {
	resp := PromptResponse{
		Messages: []PromptMsg{{Role: "user", Content: "hi"}},
	}
	if resp.Messages[0].Content != "hi" {
		t.Fatal("content mismatch")
	}
}
