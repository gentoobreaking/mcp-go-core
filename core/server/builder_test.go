package server

import (
	"context"
	"testing"
	"time"

	"github.com/project/mcp-go-core/core/feature"
	"github.com/project/mcp-go-core/core/middleware"
	"github.com/project/mcp-go-core/core/prompt"
	"github.com/project/mcp-go-core/core/protocol"
	"github.com/project/mcp-go-core/core/tool"
	"github.com/project/mcp-go-core/modules/transport/stdio"
)

// mockTool implements tool.Tool
type mockTool struct {
	name string
}

func (m *mockTool) Name() string        { return m.name }
func (m *mockTool) Description() string { return "mock tool" }
func (m *mockTool) InputSchema() tool.Schema {
	return tool.Schema{"type": "object"}
}
func (m *mockTool) Handler() tool.ToolHandler {
	return func(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
		return &protocol.Response{}, nil
	}
}

// mockResource implements resource.Resource
type mockResource struct {
	uri  string
	name string
}

func (m *mockResource) URI() string            { return m.uri }
func (m *mockResource) Name() string           { return m.name }
func (m *mockResource) Description() string    { return "mock resource" }
func (m *mockResource) Read(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	return &protocol.Response{}, nil
}

// mockPrompt implements prompt.Prompt
type mockPrompt struct {
	name string
}

func (m *mockPrompt) Name() string        { return m.name }
func (m *mockPrompt) Description() string { return "mock prompt" }
func (m *mockPrompt) Get(ctx context.Context, req prompt.PromptRequest) (prompt.PromptResponse, error) {
	return prompt.PromptResponse{}, nil
}

func TestNewBuilder(t *testing.T) {
	b := NewBuilder()
	if b == nil {
		t.Fatal("expected non-nil builder")
	}
}

func TestBuilderWithName(t *testing.T) {
	b := NewBuilder().WithName("my-server")
	if b.name != "my-server" {
		t.Fatalf("expected name 'my-server', got '%s'", b.name)
	}
}

func TestBuilderWithTool(t *testing.T) {
	b := NewBuilder()
	b.WithTool(&mockTool{name: "greet"})
	if len(b.tools) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(b.tools))
	}
}

func TestBuilderWithResource(t *testing.T) {
	b := NewBuilder()
	b.WithResource(&mockResource{uri: "test://", name: "res1"})
	if len(b.resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(b.resources))
	}
}

func TestBuilderWithPrompt(t *testing.T) {
	b := NewBuilder()
	b.WithPrompt(&mockPrompt{name: "prompt1"})
	if len(b.prompts) != 1 {
		t.Fatalf("expected 1 prompt, got %d", len(b.prompts))
	}
}

func TestBuilderWithTimeout(t *testing.T) {
	b := NewBuilder().WithTimeout(5 * time.Second)
	if b.timeout != 5*time.Second {
		t.Fatal("timeout mismatch")
	}
}

func TestBuilderBuildNoTransport(t *testing.T) {
	b := NewBuilder()
	_, err := b.Build()
	if err == nil {
		t.Fatal("expected error when no transport set")
	}
}

func TestBuilderBuildSuccess(t *testing.T) {
	tr := stdio.New(nil, nil)
	b := NewBuilder().
		WithName("test-server").
		WithTool(&mockTool{name: "greet"}).
		WithResource(&mockResource{uri: "test://", name: "res1"}).
		WithPrompt(&mockPrompt{name: "p1"}).
		WithTransport(tr).
		WithMiddleware(middleware.Recovery())

	srv, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	if srv == nil {
		t.Fatal("expected non-nil server")
	}
	if srv.name != "test-server" {
		t.Fatal("name mismatch")
	}
}

func TestBuilderMustBuild(t *testing.T) {
	tr := stdio.New(nil, nil)
	b := NewBuilder().WithTransport(tr)

	srv := b.MustBuild()
	if srv == nil {
		t.Fatal("expected non-nil server")
	}
}

func TestBuilderMustBuildPanics(t *testing.T) {
	b := NewBuilder()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on MustBuild without transport")
		}
	}()
	b.MustBuild()
}

func TestBuilderFluentAPI(t *testing.T) {
	tr := stdio.New(nil, nil)
	b := NewBuilder().
		WithName("fluent-server").
		WithTool(&mockTool{name: "t1"}).
		WithTool(&mockTool{name: "t2"}).
		WithResource(&mockResource{uri: "test://r1", name: "r1"}).
		WithPrompt(&mockPrompt{name: "p1"}).
		WithTimeout(15 * time.Second).
		WithTransport(tr).
		WithMiddleware(middleware.Recovery(), middleware.Logging(middleware.LoggerFunc(
			func(level, msg string, args ...any) {},
		)))

	if len(b.tools) != 2 {
		t.Fatalf("expected 2 tools, got %d", len(b.tools))
	}
	if len(b.resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(b.resources))
	}
	if len(b.prompts) != 1 {
		t.Fatalf("expected 1 prompt, got %d", len(b.prompts))
	}
	if len(b.mw) != 2 {
		t.Fatalf("expected 2 middlewares, got %d", len(b.mw))
	}

	srv, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	_ = srv
	_ = srv
}

func TestServerWithFlags(t *testing.T) {
	tr := stdio.New(nil, nil)
	flags := feature.NewFlags(map[string]bool{"advanced": true})
	s := NewBuilder().
		WithTool(&mockTool{name: "t1"}).
		WithFlags(flags).
		WithTransport(tr).
		MustBuild()

	if s.flags == nil {
		t.Fatal("expected flags to be set")
	}
	if !s.flags.Get("advanced").Enabled {
		t.Fatal("expected advanced flag to be enabled")
	}
}
