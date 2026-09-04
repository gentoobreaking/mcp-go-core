package http

import (
	"context"
	"encoding/json"
	"testing"
)

type mockHandler struct{}

func (m *mockHandler) HandleRequest(method string, params []byte) ([]byte, error) {
	return json.Marshal(map[string]any{"jsonrpc": "2.0", "result": "ok"})
}

func TestNewHTTPTransport(t *testing.T) {
	tr := New("localhost:18080")
	if tr == nil {
		t.Fatal("expected non-nil transport")
	}
}

func TestHTTPTransportServe(t *testing.T) {
	tr := New("localhost:18081")
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		_ = tr.Serve(ctx, func(ctx context.Context, msg json.RawMessage) (any, error) {
			return map[string]any{"status": "ok"}, nil
		})
	}()

	cancel() // immediate cancel
}
