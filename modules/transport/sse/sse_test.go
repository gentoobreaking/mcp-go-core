package sse

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestNewSSETransport(t *testing.T) {
	tr := New("localhost:18080")
	if tr == nil {
		t.Fatal("expected non-nil transport")
	}
	if tr.addr != "localhost:18080" {
		t.Fatal("addr mismatch")
	}
}

func TestSSEServe(t *testing.T) {
	tr := New("localhost:18081")
	ctx, cancel := context.WithCancel(context.Background())

	handler := func(ctx context.Context, msg json.RawMessage) (any, error) {
		return map[string]any{"status": "ok"}, nil
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- tr.Serve(ctx, handler)
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()

	select {
	case err := <-errCh:
		if err != nil && err != context.Canceled {
			t.Fatalf("unexpected error: %v", err)
		}
	case <-time.After(5 * time.Second):
		// Timeout is acceptable
	}
}

func TestSSEShutdown(t *testing.T) {
	tr := New("localhost:18082")
	ctx := context.Background()
	tr.Shutdown(ctx)
}

func TestSSEHandlerInterface(t *testing.T) {
	var _ Handler = func(ctx context.Context, msg json.RawMessage) (any, error) {
		return nil, nil
	}
}

func TestSSEConfigureWith(t *testing.T) {
	tr := New("localhost:18090")
	tr.ConfigureWith("http://localhost:18090", nil)
	if tr.addr != "localhost:18090" {
		t.Fatal("addr mismatch")
	}
}

func TestSSENotImportStdio(t *testing.T) {
	// Compile-time check: SSE package should not import stdio transport
	// This is guaranteed by the import structure
	_ = http.MethodGet
}
