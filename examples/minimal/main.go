package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// Request represents a JSON-RPC request.
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Response represents a JSON-RPC response.
type Response struct {
	JSONRPC string  `json:"jsonrpc"`
	ID      any     `json:"id"`
	Result  any     `json:"result,omitempty"`
	Error   *Error  `json:"error,omitempty"`
}

// Error represents a JSON-RPC error.
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Tool is a minimal tool definition.
type Tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Annotations map[string]any  `json:"annotations,omitempty"`
}

// Handler processes a tool call.
type Handler func(ctx context.Context, req *Request) (any, *Error)

// Server is a minimal MCP server.
type Server struct {
	mu       sync.Mutex
	tools    map[string]Tool
	handlers map[string]Handler
}

// NewServer creates a new Server.
func NewServer() *Server {
	return &Server{
		tools:    make(map[string]Tool),
		handlers: make(map[string]Handler),
	}
}

// AddTool registers a tool with a handler.
func (s *Server) AddTool(tool Tool, h Handler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tools[tool.Name] = tool
	s.handlers[tool.Name] = h
}

// Run starts the server, reading from stdin and writing to stdout.
func (s *Server) Run(ctx context.Context) error {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var req Request
		if err := json.Unmarshal(line, &req); err != nil {
			fmt.Fprintf(os.Stderr, "parse error: %v\n", err)
			continue
		}
		resp := s.handle(ctx, &req)
		data, _ := json.Marshal(resp)
		fmt.Println(string(data))
	}
	return scanner.Err()
}

func (s *Server) handle(ctx context.Context, req *Request) *Response {
	s.mu.Lock()
	tool, ok := s.tools[req.Method]
	h, hOk := s.handlers[req.Method]
	s.mu.Unlock()

	if !ok || !hOk {
		return &Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &Error{Code: -32601, Message: "Method not found"},
		}
	}
	_ = tool
	result, err := h(ctx, req)
	resp := &Response{JSONRPC: "2.0", ID: req.ID, Result: result}
	if err != nil {
		resp.Error = err
	}
	return resp
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	srv := NewServer()
	srv.AddTool(Tool{
		Name:        "greet",
		Description: "Returns a greeting message",
	}, func(ctx context.Context, req *Request) (any, *Error) {
		return map[string]any{"message": "Hello from mcp-go-core!"}, nil
	})

	if err := srv.Run(ctx); err != nil && err != context.Canceled {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}
