package oauth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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
	if !strings.Contains(err.Error(), "missing") {
		t.Fatalf("expected missing header error, got: %v", err)
	}
}

func TestAuthenticateInvalidFormat(t *testing.T) {
	a := NewAuthenticator("https://auth.example.com", "client123", []string{})
	_, err := a.Authenticate(context.Background(), &mockRequest{
		headers: map[string]string{"Authorization": "Basic sometoken"},
	})
	if err == nil {
		t.Fatal("expected error for invalid format")
	}
}

func TestAuthenticateEmptyBearer(t *testing.T) {
	a := NewAuthenticator("https://auth.example.com", "client123", []string{})
	_, err := a.Authenticate(context.Background(), &mockRequest{
		headers: map[string]string{"Authorization": "Bearer "},
	})
	if err == nil {
		t.Fatal("expected error for empty bearer token")
	}
}

func TestAuthenticateJWTValid(t *testing.T) {
	secret := "test-secret"
	now := time.Now().Unix()
	claims := jwtClaims{
		Subject:  "user123",
		Issuer:   "test-iss",
		Expiry:   now + 3600,
		IssuedAt: now,
		Scope:    "read write",
	}

	token := makeTestJWT(claims, secret)

	a := NewAuthenticator("https://auth.example.com", "client123", []string{"read", "write"})
	a = a.WithJWT("test-iss", secret)

	identity, err := a.Authenticate(context.Background(), &mockRequest{
		headers: map[string]string{"Authorization": "Bearer " + token},
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if identity.Principal != "user123" {
		t.Fatalf("expected principal user123, got: %s", identity.Principal)
	}
	if len(identity.Scopes) != 2 {
		t.Fatalf("expected 2 scopes, got: %d", len(identity.Scopes))
	}
}

func TestAuthenticateJWTExpired(t *testing.T) {
	secret := "test-secret"
	now := time.Now().Unix()
	claims := jwtClaims{
		Subject:  "user123",
		Issuer:   "test-iss",
		Expiry:   now - 3600, // expired
		IssuedAt: now - 7200,
		Scope:    "read",
	}

	token := makeTestJWT(claims, secret)

	a := NewAuthenticator("https://auth.example.com", "client123", []string{"read"})
	a = a.WithJWT("test-iss", secret)

	_, err := a.Authenticate(context.Background(), &mockRequest{
		headers: map[string]string{"Authorization": "Bearer " + token},
	})
	if err == nil {
		t.Fatal("expected error for expired token")
	}
	if !strings.Contains(err.Error(), "expired") {
		t.Fatalf("expected expired error, got: %v", err)
	}
}

func TestAuthenticateJWTInvalidSignature(t *testing.T) {
	secret := "test-secret"
	now := time.Now().Unix()
	claims := jwtClaims{
		Subject:  "user123",
		Issuer:   "test-iss",
		Expiry:   now + 3600,
		IssuedAt: now,
		Scope:    "read",
	}

	token := makeTestJWT(claims, secret)
	// Tamper with signature
	token = token[:len(token)-5] + "XXXXX"

	a := NewAuthenticator("https://auth.example.com", "client123", []string{"read"})
	a = a.WithJWT("test-iss", secret)

	_, err := a.Authenticate(context.Background(), &mockRequest{
		headers: map[string]string{"Authorization": "Bearer " + token},
	})
	if err == nil {
		t.Fatal("expected error for invalid signature")
	}
}

func TestAuthenticateJWTIssuerMismatch(t *testing.T) {
	secret := "test-secret"
	now := time.Now().Unix()
	claims := jwtClaims{
		Subject:  "user123",
		Issuer:   "wrong-iss",
		Expiry:   now + 3600,
		IssuedAt: now,
		Scope:    "read",
	}

	token := makeTestJWT(claims, secret)

	a := NewAuthenticator("https://auth.example.com", "client123", []string{"read"})
	a = a.WithJWT("test-iss", secret)

	_, err := a.Authenticate(context.Background(), &mockRequest{
		headers: map[string]string{"Authorization": "Bearer " + token},
	})
	if err == nil {
		t.Fatal("expected error for issuer mismatch")
	}
	if !strings.Contains(err.Error(), "issuer") {
		t.Fatalf("expected issuer mismatch error, got: %v", err)
	}
}

func TestAuthenticateJWTScopeNotAllowed(t *testing.T) {
	secret := "test-secret"
	now := time.Now().Unix()
	claims := jwtClaims{
		Subject:  "user123",
		Issuer:   "test-iss",
		Expiry:   now + 3600,
		IssuedAt: now,
		Scope:    "admin",
	}

	token := makeTestJWT(claims, secret)

	a := NewAuthenticator("https://auth.example.com", "client123", []string{"read"})
	a = a.WithJWT("test-iss", secret)

	_, err := a.Authenticate(context.Background(), &mockRequest{
		headers: map[string]string{"Authorization": "Bearer " + token},
	})
	if err == nil {
		t.Fatal("expected error for disallowed scope")
	}
}

func TestAuthenticateIntrospectionFallback(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"active":   true,
			"sub":      "introspected_user",
			"scope":    "read write",
			"username": "intro_user",
		})
	}))
	defer server.Close()

	a := NewAuthenticator("https://auth.example.com", "client123", []string{})
	a.tokenURL = server.URL + "/token"

	identity, err := a.Authenticate(context.Background(), &mockRequest{
		headers: map[string]string{"Authorization": "Bearer opaque-token"},
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if identity.Principal != "introspected_user" {
		t.Fatalf("expected introspected_user, got: %s", identity.Principal)
	}
}

func TestAuthenticateIntrospectionInactiveToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"active": false,
		})
	}))
	defer server.Close()

	a := NewAuthenticator("https://auth.example.com", "client123", []string{})
	a.tokenURL = server.URL + "/token"

	_, err := a.Authenticate(context.Background(), &mockRequest{
		headers: map[string]string{"Authorization": "Bearer inactive-token"},
	})
	if err == nil {
		t.Fatal("expected error for inactive token")
	}
}

func TestAuthenticateIntrospectionError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	a := NewAuthenticator("https://auth.example.com", "client123", []string{})
	a.tokenURL = server.URL + "/token"

	_, err := a.Authenticate(context.Background(), &mockRequest{
		headers: map[string]string{"Authorization": "Bearer token"},
	})
	if err == nil {
		t.Fatal("expected error for introspection failure")
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
	expired := &oauth2.Token{
		AccessToken: "test",
		Expiry:      time.Now().Add(-1 * time.Hour),
	}
	if TokenIsValid(expired) {
		t.Fatal("expired token should be invalid")
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

	u, err := pkce.AuthorizationURL("https://auth.example.com/authorize", "client123", "http://localhost/cb", []string{"read", "write"})
	if err != nil {
		t.Fatal(err)
	}

	if u == "" {
		t.Fatal("expected non-empty URL")
	}

	if !strings.Contains(u, pkce.Challenge) {
		t.Fatal("URL should contain challenge")
	}

	if !strings.Contains(u, "S256") {
		t.Fatal("URL should contain method S256")
	}
}

func TestIntrospectToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"active": true,
			"sub":    "test_user",
			"scope":  "read write",
		})
	}))
	defer server.Close()

	identity, err := IntrospectToken(context.Background(), &oauth2.Token{
		AccessToken: "test-token",
		Expiry:      time.Now().Add(1 * time.Hour),
	}, server.URL+"/token")
	if err != nil {
		t.Fatal(err)
	}
	if identity.Principal == "" {
		t.Fatal("expected non-empty principal")
	}
	if identity.Principal != "test_user" {
		t.Fatalf("expected test_user, got: %s", identity.Principal)
	}
}

func TestIntrospectToken_Invalid(t *testing.T) {
	_, err := IntrospectToken(context.Background(), nil, "")
	if err == nil {
		t.Fatal("expected error for nil token")
	}
	if !strings.Contains(err.Error(), "introspection URL is required") {
		t.Fatalf("expected URL required error, got: %v", err)
	}

	_, err = IntrospectToken(context.Background(), &oauth2.Token{}, "https://auth.example.com/token")
	if err == nil {
		t.Fatal("expected error for empty token")
	}
}

func TestRevokeToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("expected Bearer token in Authorization header")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	err := RevokeToken(context.Background(), &oauth2.Token{
		AccessToken: "test-token",
	}, server.URL+"/revoke")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestRevokeTokenError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	err := RevokeToken(context.Background(), &oauth2.Token{
		AccessToken: "test-token",
	}, server.URL+"/revoke")
	if err == nil {
		t.Fatal("expected error for server failure")
	}
}

// makeTestJWT creates a JWT token signed with HMAC-SHA256 for testing.
func makeTestJWT(claims jwtClaims, secret string) string {
	header := `{"alg":"HS256","typ":"JWT"}`
	headerB64 := base64.RawURLEncoding.EncodeToString([]byte(header))

	payload, _ := json.Marshal(claims)
	payloadB64 := base64.RawURLEncoding.EncodeToString(payload)

	signingInput := headerB64 + "." + payloadB64

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signingInput))
	signature := mac.Sum(nil)
	sigB64 := base64.RawURLEncoding.EncodeToString(signature)

	return signingInput + "." + sigB64
}
