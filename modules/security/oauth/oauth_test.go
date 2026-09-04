package oauth

import (
	"context"
	"testing"

	"golang.org/x/oauth2"
)

type mockRequest struct {
	headers map[string]string
}

func (r *mockRequest) Header(name string) string {
	return r.headers[name]
}

func TestNewAuthenticator(t *testing.T) {
	a := NewAuthenticator("https://auth.example.com", "client123", []string{"read", "write"})
	if a == nil {
		t.Fatal("expected non-nil authenticator")
	}
}

func TestAuthenticateMissingHeader(t *testing.T) {
	a := NewAuthenticator("https://auth.example.com", "client123", []string{})
	_, err := a.Authenticate(context.Background(), &mockRequest{headers: map[string]string{}})
	if err == nil {
		t.Fatal("expected error for missing header")
	}
}

func TestTokenIsValid(t *testing.T) {
	if !TokenIsValid(&oauth2.Token{AccessToken: "test"}) {
		t.Fatal("token should be valid")
	}
	if TokenIsValid(nil) {
		t.Fatal("nil token should be invalid")
	}
	if TokenIsValid(&oauth2.Token{}) {
		t.Fatal("empty token should be invalid")
	}
}

func TestNewOAuthClient(t *testing.T) {
	config := &oauth2.Config{
		ClientID:    "test",
		ClientSecret: "secret",
		RedirectURL: "http://localhost/callback",
		Scopes:      []string{"read"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://example.com/auth",
			TokenURL: "https://example.com/token",
		},
	}
	client := NewOAuthClient(config)
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}
