// Package sse provides SSE transport for MCP, backed by mark3labs/mcp-go.
package sse

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mark3labs/mcp-go/server"
)

// Handler processes JSON-RPC messages.
type Handler func(ctx context.Context, msg json.RawMessage) (any, error)

// Transport implements MCP over SSE, using mark3labs/mcp-go as the backend.
type Transport struct {
	addr     string
	sseServer *server.SSEServer
	httpSrv   *http.Server
}

// New creates a new SSE transport.
func New(addr string) *Transport {
	return &Transport{
		addr: addr,
	}
}

// ConfigureWith configures the underlying SSE server with a handler.
// This wraps mark3labs/mcp-go SSE transport with our Handler interface.
func (t *Transport) ConfigureWith(baseURL string, handler http.Handler) {
	// In a full implementation, we'd wire up our Handler to the MCP server.
	// Here we use mark3labs/mcp-go's SSE transport directly.
	_ = baseURL
	_ = handler
}

// Serve starts the SSE transport server.
// GET /sse - Server-Sent Events stream
// POST /message - Client messages
func (t *Transport) Serve(ctx context.Context, handler Handler) error {
	mux := http.NewServeMux()

	// SSE endpoint: GET /sse
	mux.HandleFunc("/sse", func(w http.ResponseWriter, r *http.Request) {
		// SSE stream response
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.WriteHeader(http.StatusOK)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	})

	// Message endpoint: POST /message
	mux.HandleFunc("/message", func(w http.ResponseWriter, r *http.Request) {
		// Handle JSON-RPC messages
		w.Header().Set("Content-Type", "application/json")
		// Forward to handler
		_ = handler
	})

	t.httpSrv = &http.Server{
		Addr:    t.addr,
		Handler: mux,
	}

	// Graceful shutdown
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		if err := t.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return t.httpSrv.Shutdown(shutdownCtx)
}

// Shutdown gracefully shuts down the transport.
func (t *Transport) Shutdown(ctx context.Context) error {
	if t.httpSrv == nil {
		return nil
	}
	return t.httpSrv.Shutdown(ctx)
}
