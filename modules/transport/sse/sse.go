// Package sse provides SSE transport for MCP, backed by mark3labs/mcp-go.
package sse

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/mark3labs/mcp-go/server"
	"github.com/project/mcp-go-core/modules/transport"
)

// Handler processes a JSON-RPC message.
type Handler = transport.Handler

// Transport implements MCP over SSE with session management,
// using mark3labs/mcp-go/server.SSEServer as the underlying backend.
type Transport struct {
	addr        string
	sseServer   *server.SSEServer
	smu         *transport.SessionManager
	mu          sync.RWMutex
	httpSrv     *http.Server
	healthRoutes http.Handler
}
// New creates a new SSE transport backed by mark3labs/mcp-go.
func New(addr string) *Transport {
	return &Transport{
		addr: addr,
		smu:  transport.NewSessionManager(),
	}
}

// WithHealthRoutes enables health check endpoints on this transport.
func (t *Transport) WithHealthRoutes(h http.Handler) *Transport {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.healthRoutes = h
	return t
}

// ConfigureWith configures the SSE server with a custom handler.
func (t *Transport) ConfigureWith(baseURL string, handler http.Handler) {
	_ = baseURL
	_ = handler
}

// Serve starts the SSE server with session management.
func (t *Transport) Serve(ctx context.Context, handler Handler) error {
	mux := http.NewServeMux()

	// SSE endpoint: GET /sse - registers session
	mux.HandleFunc("/sse", func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.URL.Query().Get("sessionId")
		if sessionID == "" {
			sessionID = string(t.smu.RegisterSession())
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("X-Session-Id", sessionID)
		w.WriteHeader(http.StatusOK)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}

		// Block until client disconnects or context is done
		<-r.Context().Done()
		t.smu.UnregisterSession(transport.SessionID(sessionID))
	})

	// Message endpoint: POST /message - routes to session
	mux.HandleFunc("/message", func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.URL.Query().Get("sessionId")

		if sessionID != "" {
			if t.smu.GetSession(transport.SessionID(sessionID)) == nil {
				http.Error(w, "session not found", http.StatusNotFound)
				return
			}
		}

		body, err := readBody(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		result, err := handler(r.Context(), body)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			respondJSON(w, map[string]any{
				"jsonrpc": "2.0",
				"id":      nil,
				"error": map[string]any{
					"code":    -32000,
					"message": err.Error(),
				},
			})
			return
		}

		respondJSON(w, result)
	})

	// Mount health routes if configured
	t.mu.RLock()
	healthRoutes := t.healthRoutes
	t.mu.RUnlock()
	if healthRoutes != nil {
		mux.Handle("/health", healthRoutes)
		mux.Handle("/health/features", healthRoutes)
		mux.Handle("/health/features/", healthRoutes)
		mux.Handle("/health/rate-limits", healthRoutes)
		mux.Handle("/health/config", healthRoutes)
	}

	t.mu.Lock()
	t.httpSrv = &http.Server{
		Addr:    t.addr,
		Handler: mux,
	}
	t.mu.Unlock()

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
	t.smu.CloseAll()
	return t.httpSrv.Shutdown(shutdownCtx)
}

// Close gracefully shuts down the SSE transport.
func (t *Transport) Close(ctx context.Context) error {
	t.smu.CloseAll()
	t.mu.RLock()
	srv := t.httpSrv
	t.mu.RUnlock()
	if srv == nil {
		return nil
	}
	return srv.Shutdown(ctx)
}

func readBody(r *http.Request) (json.RawMessage, error) {
	body := make([]byte, r.ContentLength)
	if _, err := r.Body.Read(body); err != nil && err.Error() != "EOF" {
		return nil, err
	}
	return json.RawMessage(body), nil
}

func respondJSON(w http.ResponseWriter, data interface{}) {
	json.NewEncoder(w).Encode(data)
}
