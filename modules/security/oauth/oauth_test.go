package oauth

import (
	"context"
	"testing"
	"time"

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
	if len(a.allowed) != 2 {
		t.Fatalf("expected 2 allowed scopes, got %d", len(a.allowed))
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
		ClientID:     "test",
		ClientSecret: "secret",
		RedirectURL:  "http://localhost/callback",
		Scopes:       []string{"read"},
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

func TestGeneratePKCE(t *testing.T) {
	pkce, err := GeneratePKCE()
	if err != nil {
		t.Fatal(err)
	}
	if pkce.Verifier == "" {
		t.Fatal("expected non-empty verifier")
	}
	if pkce.Challenge == "" {
		t.Fatal("expected non-empty challenge")
	}
	if pkce.Method != "S256" {
		t.Fatalf("expected S256, got %s", pkce.Method)
	}
	if len(pkce.Verifier) < 32 {
		t.Fatal("verifier too short")
	}
}

func TestPKCE_UniqueVerification(t *testing.T) {
	p1, err := GeneratePKCE()
	if err != nil {
		t.Fatal(err)
	}
	p2, err := GeneratePKCE()
	if err != nil {
		t.Fatal(err)
	}
	if p1.Verifier == p2.Verifier {
		t.Fatal("expected unique verifier")
	}
	if p1.Challenge == p2.Challenge {
		t.Fatal("expected unique challenge")
	}
}

func TestPKCEAuthorizationURL(t *testing.T) {
	pkce, err := GeneratePKCE()
	if err != nil {
		t.Fatal(err)
	}

	url, err := pkce.AuthorizationURL("https://auth.example.com/authorize", "client123", "http://localhost/cb", []string{"read", "write"})
	if err != nil {
		t.Fatal(err)
	}

	if url == "" {
		t.Fatal("expected non-empty URL")
	}

	if !contains(url, pkce.Challenge) {
		t.Fatal("URL should contain challenge")
	}

	if !contains(url, "S256") {
		t.Fatal("URL should contain method S256")
	}
}

func TestIntrospectToken(t *testing.T) {
	identity, err := IntrospectToken(context.Background(), &oauth2.Token{
		AccessToken: "test-token",
		Expiry:      time.Now().Add(1 * time.Hour),
	}, "https://auth.example.com/token")
	if err != nil {
		t.Fatal(err)
	}
	if identity.Principal == "" {
		t.Fatal("expected non-empty principal")
	}
}

func TestIntrospectToken_Invalid(t *testing.T) {
	_, err := IntrospectToken(context.Background(), nil, "https://auth.example.com/token")
	if err == nil {
		t.Fatal("expected error for nil token")
	}

	_, err = IntrospectToken(context.Background(), &oauth2.Token{}, "https://auth.example.com/token")
	if err == nil {
		t.Fatal("expected error for empty token")
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
