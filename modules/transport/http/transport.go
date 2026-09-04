// Package http provides Streamable HTTP transport for MCP.
package http

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/project/mcp-go-core/modules/transport"
)

// Handler processes a JSON-RPC message.
type Handler = transport.Handler

// Transport implements MCP over Streamable HTTP.
type Transport struct {
	server       *http.Server
	healthRoutes http.Handler // optional; set via WithHealthRoutes
	mu           sync.Mutex
}
// WithHealthRoutes enables health check endpoints on this transport.
func (t *Transport) WithHealthRoutes(h http.Handler) *Transport {
	t.healthRoutes = h
	return t
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
	// Mount health routes if configured — health mux takes priority, falls through to main mux
	var rootHandler http.Handler = mux
	if t.healthRoutes != nil {
		rootHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/health") {
				t.healthRoutes.ServeHTTP(w, r)
				return
			}
			mux.ServeHTTP(w, r)
		})
	}
	t.server.Handler = rootHandler
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

// Close shuts down the HTTP transport.
func (t *Transport) Close(ctx context.Context) error {
	return t.server.Shutdown(ctx)
}
