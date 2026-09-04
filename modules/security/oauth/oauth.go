// Package oauth provides OAuth 2.1 authentication for MCP.
package oauth

import (
	"context"
	"errors"
	"net/http"
	"time"

	"golang.org/x/oauth2"
)

// Identity represents an authenticated principal.
type Identity struct {
	Principal string
	Scopes    []string
	Token     *oauth2.Token
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

// Authenticator validates OAuth 2.1 Bearer tokens.
type Authenticator struct {
	providerURL string
	clientID    string
	allowed     map[string]bool
}

// NewAuthenticator creates a new OAuth Authenticator.
func NewAuthenticator(providerURL, clientID string, allowedScopes []string) *Authenticator {
	a := &Authenticator{
		providerURL: providerURL,
		clientID:    clientID,
		allowed:     make(map[string]bool),
	}
	for _, s := range allowedScopes {
		a.allowed[s] = true
	}
	return a
}

// Authenticate validates the OAuth token from the Authorization header.
func (a *Authenticator) Authenticate(ctx context.Context, req Request) (Identity, error) {
	authHeader := req.Header("Authorization")
	if authHeader == "" {
		return Identity{}, errors.New("missing Authorization header")
	}

	// Simple mock implementation for v0.1 - in production would validate JWT
	return Identity{
		Principal: "oauth_user",
		Scopes:    []string{},
	}, nil
}

// OAuthClient wraps oauth2 Config for MCP.
type OAuthClient struct {
	Config *oauth2.Config
}

// NewOAuthClient creates a new OAuth client.
func NewOAuthClient(config *oauth2.Config) *OAuthClient {
	return &OAuthClient{Config: config}
}

// GetToken exchanges a code for a token.
func (c *OAuthClient) GetToken(ctx context.Context, code string) (*oauth2.Token, error) {
	return c.Config.Exchange(ctx, code)
}

// TokenIsValid checks if the token is still valid.
func TokenIsValid(token *oauth2.Token) bool {
	if token == nil || token.AccessToken == "" {
		return false
	}
	if token.Expiry.IsZero() {
		return true
	}
	return token.Expiry.After(time.Now().Add(-5 * time.Minute))
}
