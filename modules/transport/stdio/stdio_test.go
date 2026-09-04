package stdio

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/project/mcp-go-core/modules/transport"
)

// Compile-time check: stdio Transport implements Transport interface.
var _ transport.Transport = (*Transport)(nil)

func TestNewStdioTransport(t *testing.T) {
	tr := New(bytes.NewReader([]byte{}), &bytes.Buffer{})
	if tr == nil {
		t.Fatal("expected non-nil transport")
	}
}

func TestStdioServe(t *testing.T) {
	in := bytes.NewReader([]byte{})
	out := &bytes.Buffer{}
	tr := New(in, out)

	ctx, cancel := context.WithCancel(context.Background())
	handler := func(ctx context.Context, msg json.RawMessage) (any, error) {
		return map[string]any{"status": "ok"}, nil
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- tr.Serve(ctx, handler)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-errCh:
		if err != nil && err != context.Canceled {
			t.Fatalf("unexpected error: %v", err)
		}
	case <-time.After(2 * time.Second):
	}
}

func TestStdioClose(t *testing.T) {
	tr := New(&bytes.Buffer{}, &bytes.Buffer{})
	if err := tr.Close(context.Background()); err != nil {
		t.Fatal(err)
	}
}
