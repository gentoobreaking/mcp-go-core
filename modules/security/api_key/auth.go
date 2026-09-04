// Package api_key provides API Key authentication for MCP.
package api_key

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
)

// Identity represents an authenticated principal.
type Identity struct {
	Principal string
	Scopes    []string
}

// Request is the interface for extracting auth info from a request.
type Request interface {
	Header(name string) string
}

// HTTPRequest wraps *http.Request to satisfy Request interface.
type HTTPRequest struct {
	*http.Request
}

// Header returns the header value.
func (r *HTTPRequest) Header(name string) string {
	return r.Request.Header.Get(name)
}

// Authenticator validates API keys.
type Authenticator struct {
	keys map[string]Identity
}

// NewAuthenticator creates a new Authenticator with the given valid keys.
func NewAuthenticator(keys map[string]Identity) *Authenticator {
	return &Authenticator{keys: keys}
}

// Authenticate validates the API key from the request.
func (a *Authenticator) Authenticate(ctx context.Context, req Request) (Identity, error) {
	key := req.Header("X-API-Key")
	if key == "" {
		key = req.Header("Authorization")
		// Strip "Bearer " prefix if present
		for i := 0; i < len(key); i++ {
			if key[i] == ' ' {
				key = key[i+1:]
				break
			}
		}
	}

	if key == "" {
		return Identity{}, errors.New("missing API key")
	}

	identity, ok := a.keys[key]
	if !ok {
		return Identity{}, errors.New("invalid API key")
	}

	return identity, nil
}

// JSONRPCRequest wraps a JSON-RPC request for auth.
type JSONRPCRequest struct {
	ID      any             `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}
