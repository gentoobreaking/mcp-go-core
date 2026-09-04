// Package jwt provides JWT authentication for MCP.
// This package does NOT import OAuth, OTel, or Kubernetes packages.
package jwt

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

// Identity represents an authenticated principal.
type Identity struct {
	Principal string
	Claims    map[string]any
}

// Request is the interface for extracting auth info from a request.
type Request interface {
	Header(name string) string
}

// Authenticator validates JWT tokens.
type Authenticator struct {
	secret string
	issuer string
}

// NewAuthenticator creates a new JWT Authenticator.
func NewAuthenticator(secret, issuer string) *Authenticator {
	return &Authenticator{secret: secret, issuer: issuer}
}

// Authenticate validates the JWT from the Authorization header.
func (a *Authenticator) Authenticate(ctx context.Context, req Request) (Identity, error) {
	authHeader := req.Header("Authorization")
	if authHeader == "" {
		return Identity{}, errors.New("missing Authorization header")
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return Identity{}, errors.New("invalid Authorization header format")
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	return a.validateToken(token)
}

// validateToken validates the JWT signature and expiration.
func (a *Authenticator) validateToken(tokenString string) (Identity, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return Identity{}, errors.New("invalid JWT format")
	}

	_, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return Identity{}, errors.New("invalid JWT header")
	}

	_, err = base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return Identity{}, errors.New("invalid JWT payload")
	}

	var claims map[string]any
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return Identity{}, errors.New("invalid JWT payload base64")
	}
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return Identity{}, errors.New("invalid JWT payload JSON")
	}

	// Check expiration
	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			return Identity{}, errors.New("token expired")
		}
	}

	// Check issuer
	if a.issuer != "" {
		if issuer, ok := claims["iss"].(string); ok && issuer != a.issuer {
			return Identity{}, errors.New("invalid issuer")
		}
	}

	// Verify signature
	payload := parts[0] + "." + parts[1]
	expectedSig := sign(payload, a.secret)
	if !hmac.Equal([]byte(parts[2]), []byte(expectedSig)) {
		return Identity{}, errors.New("invalid signature")
	}

	principal := ""
	if sub, ok := claims["sub"].(string); ok {
		principal = sub
	}

	return Identity{
		Principal: principal,
		Claims:    claims,
	}, nil
}

// sign creates a HMAC-SHA256 signature.
func sign(data, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}
