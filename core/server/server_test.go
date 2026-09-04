package server

import (
	"context"
	"testing"
	"time"

	"github.com/project/mcp-go-core/core/tool"
)

func TestNewServer(t *testing.T) {
	s := NewServer()
	if s == nil {
		t.Fatal("expected non-nil server")
	}
}

func TestServerWithOpts(t *testing.T) {
	s := NewServer(WithName("test"), WithShutdownTimeout(5 * time.Second))
	if s.name != "test" {
		t.Fatal("name not set")
	}
	if s.timeout != 5*time.Second {
		t.Fatal("timeout not set")
	}
}

func TestServerAddTool(t *testing.T) {
	s := NewServer()
	tk := tool.NewTool("test", "test", nil, nil)
	s.AddTool(tk)
	// Tool registered successfully
}

func TestServerRunCancel(t *testing.T) {
	s := NewServer()
	ctx, cancel := context.WithCancel(context.Background())
	go s.Run(ctx)
	cancel()
	time.Sleep(100 * time.Millisecond) // allow shutdown to process
}
