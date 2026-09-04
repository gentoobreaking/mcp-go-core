// Package protocol provides MCP protocol base types.
package protocol

import "encoding/json"

// JSONRPCVersion is the JSON-RPC version used by MCP.
const JSONRPCVersion = "2.0"

// Message is a JSON-RPC message (request or response).
type Message struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      *int64          `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  any             `json:"result,omitempty"`
	Error   *Error          `json:"error,omitempty"`
}

// Error represents a JSON-RPC error.
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Request represents an incoming JSON-RPC request.
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      *int64          `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

// Response represents a JSON-RPC response.
type Response struct {
	JSONRPC string  `json:"jsonrpc"`
	ID      *int64  `json:"id"`
	Result  any     `json:"result,omitempty"`
	Error   *Error  `json:"error,omitempty"`
}
