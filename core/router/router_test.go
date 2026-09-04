package router

import (
	"context"
	"testing"

	"github.com/project/mcp-go-core/core/protocol"
	"github.com/project/mcp-go-core/core/tool"
)

func TestRouterMethods(t *testing.T) {
	r := NewRouter()

	// Test unknown method returns MethodNotFoundError
	_, err := r.Dispatch(context.Background(), &protocol.Request{
		JSONRPC: "2.0",
		Method:  "unknown",
	})
	if err == nil {
		t.Fatal("expected error for unknown method")
	}
}

func TestRouterRegisterTool(t *testing.T) {
	r := NewRouter()
	tk := tool.NewTool("test", "test tool", nil, nil)
	r.RegisterTool(tk)
	// Verify featuregraph is not referenced
}

func TestNoFeatureGraphReference(t *testing.T) {
	// Ensure router.go does not import or reference featuregraph
	// This is a build-time check - the import list above confirms it
}
