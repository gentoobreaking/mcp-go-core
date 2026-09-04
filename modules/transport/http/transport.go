// Package http provides Streamable HTTP transport for MCP.
package http

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sync"
)

// Handler processes a JSON-RPC message.
type Handler func(ctx context.Context, msg json.RawMessage) (any, error)

// Transport implements MCP over Streamable HTTP.
type Transport struct {
	server *http.Server
	mu     sync.Mutex
}

// New creates a new HTTP transport.
func New(addr string) *Transport {
	return &Transport{
		server: &http.Server{
			Addr: addr,
		},
	}
}

// Serve starts the HTTP server.
func (t *Transport) Serve(ctx context.Context, handler Handler) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			t.handlePost(w, r, handler, ctx)
		} else if r.Method == http.MethodGet {
			t.handleStream(w, r, handler, ctx)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	t.server.Handler = mux

	errCh := make(chan error, 1)
	go func() {
		if err := t.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	<-ctx.Done()
	return t.server.Shutdown(context.Background())
}

func (t *Transport) handlePost(w http.ResponseWriter, r *http.Request, handler Handler, ctx context.Context) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var req map[string]any
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	id, _ := req["id"]
	result, err := handler(ctx, body)
	if err != nil {
		resp := map[string]any{"jsonrpc": "2.0", "id": id, "error": map[string]any{"code": -32000, "message": err.Error()}}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp := map[string]any{"jsonrpc": "2.0", "id": id, "result": result}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (t *Transport) handleStream(w http.ResponseWriter, r *http.Request, handler Handler, ctx context.Context) {
	w.Header().Set("Content-Type", "application/jsonlines")
	w.WriteHeader(http.StatusOK)
	flusher, _ := w.(http.Flusher)
	if flusher != nil {
		flusher.Flush()
	}
}
