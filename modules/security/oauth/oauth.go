// Package oauth provides OAuth 2.1 authentication for MCP with PKCE support.
package oauth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
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
// In production, it validates JWT signatures locally and falls back to
// token introspection for opaque tokens.
type Authenticator struct {
	providerURL string
	clientID    string
	allowed     map[string]bool
	tokenURL    string
	issuer      string
	hmacSecret  string
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

// WithJWT enables JWT-based token validation using a shared HMAC secret.
func (a *Authenticator) WithJWT(issuer, hmacSecret string) *Authenticator {
	a.issuer = issuer
	a.hmacSecret = hmacSecret
	return a
}

// Authenticate validates the OAuth token from the Authorization header.
// It first attempts JWT validation (signature + claims), then falls back
// to token introspection for opaque tokens.
func (a *Authenticator) Authenticate(ctx context.Context, req Request) (Identity, error) {
	authHeader := req.Header("Authorization")
	if authHeader == "" {
		return Identity{}, errors.New("missing Authorization header")
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return Identity{}, errors.New("invalid Authorization header format")
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == "" {
		return Identity{}, errors.New("empty bearer token")
	}

	// Try JWT validation first if HMAC secret is configured
	if a.hmacSecret != "" {
		return a.validateJWT(token)
	}

	// Fall back to token introspection
	return IntrospectToken(ctx, &oauth2.Token{AccessToken: token, TokenType: "Bearer"}, a.tokenURL)
}

// jwtClaims is the JWT payload structure.
type jwtClaims struct {
	Subject   string `json:"sub"`
	Issuer    string `json:"iss"`
	Expiry    int64  `json:"exp"`
	IssuedAt  int64  `json:"iat"`
	Scope     string `json:"scope,omitempty"`
	Audience  string `json:"aud,omitempty"`
	NotBefore int64  `json:"nbf,omitempty"`
}

// validateJWT validates a JWT token's signature and claims.
func (a *Authenticator) validateJWT(tokenString string) (Identity, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return Identity{}, errors.New("invalid JWT format")
	}

	// Validate signature using HMAC-SHA256
	signingInput := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, []byte(a.hmacSecret))
	mac.Write([]byte(signingInput))
	expectedSig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(parts[2]), []byte(expectedSig)) {
		return Identity{}, errors.New("invalid JWT signature")
	}

	// Decode payload
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return Identity{}, fmt.Errorf("decode jwt payload: %w", err)
	}

	var claims jwtClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return Identity{}, fmt.Errorf("parse jwt claims: %w", err)
	}

	// Validate issuer
	if a.issuer != "" && claims.Issuer != a.issuer {
		return Identity{}, fmt.Errorf("jwt issuer mismatch: got %s, expected %s", claims.Issuer, a.issuer)
	}

	// Validate expiration
	if claims.Expiry < time.Now().Unix() {
		return Identity{}, errors.New("token expired")
	}

	// Validate issued at
	if claims.IssuedAt > time.Now().Unix()+300 {
		return Identity{}, errors.New("token issued in the future")
	}

	identity := Identity{
		Principal: claims.Subject,
		Scopes:    strings.Fields(claims.Scope),
	}

	// Validate scopes
	for _, scope := range identity.Scopes {
		if len(a.allowed) > 0 && !a.allowed[scope] {
			return Identity{}, fmt.Errorf("scope %q not allowed", scope)
		}
	}

	return identity, nil
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
	if tokenURL == "" {
		return Identity{}, errors.New("token introspection URL is required")
	}

	introspectURL := strings.TrimSuffix(tokenURL, "/token") + "/introspect"

	client := &http.Client{Timeout: 10 * time.Second}
	form := url.Values{}
	form.Set("token", token.AccessToken)
	form.Set("token_type_hint", "access_token")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, introspectURL, strings.NewReader(form.Encode()))
	if err != nil {
		return Identity{}, fmt.Errorf("create introspection request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	resp, err := client.Do(req)
	if err != nil {
		return Identity{}, fmt.Errorf("introspect token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Identity{}, fmt.Errorf("introspection endpoint returned status %d", resp.StatusCode)
	}

	var result struct {
		Active   bool   `json:"active"`
		Sub      string `json:"sub"`
		Scope    string `json:"scope"`
		Username string `json:"username"`
		ClientID string `json:"client_id"`
		Exp      int64  `json:"exp"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return Identity{}, fmt.Errorf("decode introspection response: %w", err)
	}

	if !result.Active {
		return Identity{}, errors.New("token is not active")
	}

	scopes := strings.Fields(result.Scope)
	principal := result.Sub
	if principal == "" {
		principal = result.Username
	}

	return Identity{
		Principal: principal,
		Scopes:    scopes,
		Token:     token,
	}, nil
}

// TokenIsValid checks if the token is still valid.
func TokenIsValid(token *oauth2.Token) bool {
	if token == nil {
		return false
	}
	if token.AccessToken == "" {
		return false
	}
	if token.Expiry.IsZero() {
		return true
	}
	return token.Expiry.After(time.Now())
}

// OAuthClient wraps oauth2 Config for MCP.
type OAuthClient struct {
	Config *oauth2.Config
}

// NewOAuthClient creates a new OAuthClient.
func NewOAuthClient(config *oauth2.Config) *OAuthClient {
	return &OAuthClient{Config: config}
}

// GetToken exchanges a code for a token.
func (c *OAuthClient) GetToken(ctx context.Context, code string) (*oauth2.Token, error) {
	return c.Config.Exchange(ctx, code)
}

// RevokeToken revokes a token.
func RevokeToken(ctx context.Context, token *oauth2.Token, revocationURL string) error {
	client := &http.Client{Timeout: 10 * time.Second}

	form := url.Values{}
	form.Set("token", token.AccessToken)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, revocationURL, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("create revocation request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("revoke token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("revocation endpoint returned status %d", resp.StatusCode)
	}

	return nil
}
