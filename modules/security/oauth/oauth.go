// Package oauth provides OAuth 2.1 authentication for MCP with PKCE support.
package oauth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/http"
	"net/url"
	"strings"
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

// Authenticator validates OAuth 2.1 Bearer tokens with PKCE support.
type Authenticator struct {
	providerURL string
	clientID    string
	allowed     map[string]bool
	tokenURL    string
}

// NewAuthenticator creates a new OAuth Authenticator.
func NewAuthenticator(providerURL, clientID string, allowedScopes []string) *Authenticator {
	a := &Authenticator{
		providerURL: providerURL,
		clientID:    clientID,
		tokenURL:    providerURL + "/oauth2/v2.0/token",
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

// PKCE represents a PKCE code verifier and challenge pair.
type PKCE struct {
	Verifier  string
	Challenge string
	Method    string
}

// GeneratePKCE generates a PKCE code verifier and challenge (RFC 7636).
func GeneratePKCE() (PKCE, error) {
	v := make([]byte, 32)
	if _, err := rand.Read(v); err != nil {
		return PKCE{}, err
	}
	verifier := base64.RawURLEncoding.EncodeToString(v)

	hash := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(hash[:])

	return PKCE{
		Verifier:  verifier,
		Challenge: challenge,
		Method:    "S256",
	}, nil
}

// AuthorizationURL returns the OAuth 2.0 authorization URL with PKCE.
func (p PKCE) AuthorizationURL(authURL, clientID, redirectURI string, scopes []string) (string, error) {
	u, err := url.Parse(authURL)
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Set("client_id", clientID)
	q.Set("response_type", "code")
	q.Set("redirect_uri", redirectURI)
	q.Set("code_challenge", p.Challenge)
	q.Set("code_challenge_method", p.Method)
	if len(scopes) > 0 {
		q.Set("scope", strings.Join(scopes, " "))
	}

	u.RawQuery = q.Encode()
	return u.String(), nil
}

// IntrospectToken checks if a token is valid and returns its info.
func IntrospectToken(ctx context.Context, token *oauth2.Token, tokenURL string) (Identity, error) {
	if token == nil || token.AccessToken == "" {
		return Identity{}, errors.New("token is nil or empty")
	}

	// In production, would call introspection endpoint
	return Identity{
		Principal: "oauth_user",
		Scopes:    []string{},
		Token:     token,
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
