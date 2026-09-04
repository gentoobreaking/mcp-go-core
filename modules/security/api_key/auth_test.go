package api_key

import (
	"context"
	"testing"
)

type mockRequest struct {
	headers map[string]string
}

func (r *mockRequest) Header(name string) string {
	return r.headers[name]
}

func TestAuthenticateValidKey(t *testing.T) {
	a := NewAuthenticator(map[string]Identity{
		"secret123": {Principal: "user1", Scopes: []string{"read"}},
	})

	identity, err := a.Authenticate(context.Background(), &mockRequest{
		headers: map[string]string{"X-API-Key": "secret123"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if identity.Principal != "user1" {
		t.Fatal("wrong principal")
	}
}

func TestAuthenticateMissingKey(t *testing.T) {
	a := NewAuthenticator(map[string]Identity{})
	_, err := a.Authenticate(context.Background(), &mockRequest{
		headers: map[string]string{},
	})
	if err == nil {
		t.Fatal("expected error for missing key")
	}
}

func TestAuthenticateInvalidKey(t *testing.T) {
	a := NewAuthenticator(map[string]Identity{
		"secret123": {Principal: "user1"},
	})
	_, err := a.Authenticate(context.Background(), &mockRequest{
		headers: map[string]string{"X-API-Key": "wrong"},
	})
	if err == nil {
		t.Fatal("expected error for invalid key")
	}
}

func TestAuthenticateBearerToken(t *testing.T) {
	a := NewAuthenticator(map[string]Identity{
		"secret123": {Principal: "user1"},
	})
	_, err := a.Authenticate(context.Background(), &mockRequest{
		headers: map[string]string{"Authorization": "Bearer secret123"},
	})
	if err != nil {
		t.Fatal(err)
	}
}
