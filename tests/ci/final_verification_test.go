package ci

import (
	"context"
	"testing"

	"github.com/project/mcp-go-core/core/prompt"
	"github.com/project/mcp-go-core/core/protocol"
	"github.com/project/mcp-go-core/core/resource"
	"github.com/project/mcp-go-core/core/router"
	"github.com/project/mcp-go-core/core/server"
	"github.com/project/mcp-go-core/core/tool"
	"github.com/project/mcp-go-core/internal/featuregraph"
	"github.com/project/mcp-go-core/modules/storage/memory"
)

func TestFinalCoreTypes(t *testing.T) {
	_ = protocol.JSONRPCVersion
	_ = protocol.Request{JSONRPC: "2.0"}
	_ = protocol.Response{JSONRPC: "2.0"}
	_ = router.NewRouter()
	_ = server.NewServer()

	tk := tool.NewTool("test", "test", tool.Schema{}, func(_ context.Context, _ *protocol.Request) (*protocol.Response, error) {
		return &protocol.Response{Result: "ok"}, nil
	})
	if tk.Name() != "test" {
		t.Fatal("tool name mismatch")
	}

	r := resource.NewResource("mcp://test", "test", "desc", func(_ context.Context, _ *protocol.Request) (*protocol.Response, error) {
		return &protocol.Response{Result: "ok"}, nil
	})
	if r.URI() != "mcp://test" {
		t.Fatal("resource URI mismatch")
	}

	p := prompt.NewPrompt("test", "desc", func(_ context.Context, _ prompt.PromptRequest) (prompt.PromptResponse, error) {
		return prompt.PromptResponse{}, nil
	})
	if p.Name() != "test" {
		t.Fatal("prompt name mismatch")
	}
}

func TestFinalModules(t *testing.T) {
	store := memory.New()
	_ = store
}

func TestFinalIntegration(t *testing.T) {
	srv := server.NewServer(server.WithName("test"))
	srv.AddTool(tool.NewTool("greet", "say hello",
		tool.Schema{"type": "object"},
		func(_ context.Context, _ *protocol.Request) (*protocol.Response, error) {
			return &protocol.Response{Result: "hello"}, nil
		},
	))
	srv.AddResource(resource.NewResource("mcp://greeting", "greeting", "a greeting",
		func(_ context.Context, _ *protocol.Request) (*protocol.Response, error) {
			return &protocol.Response{Result: "hello"}, nil
		},
	))
	srv.AddPrompt(prompt.NewPrompt("explain", "explain something",
		func(_ context.Context, _ prompt.PromptRequest) (prompt.PromptResponse, error) {
			return prompt.PromptResponse{}, nil
		},
	))
	_ = srv
}

func TestFinalFeatureGraph(t *testing.T) {
	g := featuregraph.NewGraph()
	g.AddFeature(featuregraph.FeatureDescriptor{Name: "http"})
	g.AddFeature(featuregraph.FeatureDescriptor{Name: "jwt"})

	res, err := featuregraph.Resolve(featuregraph.Config{
		Graph:          g,
		ExplicitEnable: []string{"http"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Enabled) != 1 {
		t.Fatal("expected 1 enabled feature")
	}

	lock := featuregraph.GenerateLock(res, "1.0.0")
	if lock.GraphHash == "" {
		t.Fatal("expected non-empty graph hash")
	}
}
