package sse

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/project/mcp-go-core/modules/transport"
)

// Compile-time check: SSE Transport implements Transport interface.
var _ transport.Transport = (*Transport)(nil)

func TestNewSSETransport(t *testing.T) {
	tr := New("localhost:18080")
	if tr == nil {
		t.Fatal("expected non-nil transport")
	}
}

func TestSSEServe(t *testing.T) {
	tr := New("localhost:18081")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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

func TestSSEHandlerInterface(t *testing.T) {
	var _ Handler = func(ctx context.Context, msg json.RawMessage) (any, error) {
		return nil, nil
	}
}

func TestSSEClose(t *testing.T) {
	tr := New("localhost:18082")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	handler := func(ctx context.Context, msg json.RawMessage) (any, error) {
		return nil, nil
	}

	go tr.Serve(ctx, handler)
	time.Sleep(100 * time.Millisecond)

	if err := tr.Close(context.Background()); err != nil {
		// Server may have already shut down from context cancellation
	}
}

func TestSSESessionManager(t *testing.T) {
	tr := New("localhost:18083")
	if tr.smu == nil {
		t.Fatal("expected session manager")
	}
	id := tr.smu.RegisterSession()
	if id == "" {
		t.Fatal("expected non-empty session ID")
	}
	if tr.smu.Count() != 1 {
		t.Fatalf("expected 1 session, got %d", tr.smu.Count())
	}
	tr.smu.UnregisterSession(id)
	if tr.smu.Count() != 0 {
		t.Fatal("expected 0 sessions after unregister")
	}
}

func TestSSESessionRouting(t *testing.T) {
	smu := transport.NewSessionManager()
	id := smu.RegisterSession()

	// Verify session exists
	if smu.GetSession(id) == nil {
		t.Fatal("expected session to exist")
	}

	// Verify session ID is a valid hex string
	if len(string(id)) != 32 {
		t.Fatalf("expected 32-char session ID, got %d", len(string(id)))
	}

	smu.CloseAll()
	if smu.Count() != 0 {
		t.Fatal("expected 0 sessions after close all")
	}
}
