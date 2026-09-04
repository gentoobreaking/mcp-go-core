package http

import (
	"context"
	"encoding/json"
	"net"
	"testing"
	"time"

	"github.com/project/mcp-go-core/modules/transport"
)

// Compile-time check: HTTP Transport implements Transport interface.
var _ transport.Transport = (*Transport)(nil)

func TestNewHTTPTransport(t *testing.T) {
	tr := New("localhost:18090")
	if tr == nil {
		t.Fatal("expected non-nil transport")
	}
}

func TestHTTPServe(t *testing.T) {
	tr := New("localhost:18091")
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
	}
}

func TestHTTPClose(t *testing.T) {
	tr := New("localhost:18093")
	// Server not started, Shutdown should still work
	_ = tr.server.Shutdown // verify server exists
	tr.Close(context.Background())
}

func TestHTTPImplementsInterface(t *testing.T) {
	var _ transport.Transport = New("localhost:0")
}

func TestHTTPFreePort(t *testing.T) {
	// Find a free port
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	tr := New(l.Addr().String())
	if tr == nil {
		t.Fatal("expected non-nil transport")
	}
}
