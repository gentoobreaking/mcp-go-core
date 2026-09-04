package jwt

import (
	"context"
	"encoding/base64"
	"strings"
	"testing"
	"time"
)

type mockRequest struct {
	headers map[string]string
}

func (r *mockRequest) Header(name string) string {
	return r.headers[name]
}

func makeToken(secret, issuer string, sub string, exp int64) string {
	header := `{"alg":"HS256","typ":"JWT"}`
	payload := `{"sub":"` + sub + `","iss":"` + issuer + `","exp":` + i64toa(exp) + `}`
	hB64 := base64.RawURLEncoding.EncodeToString([]byte(header))
	pB64 := base64.RawURLEncoding.EncodeToString([]byte(payload))
	sig := sign(hB64 + "." + pB64, secret)
	return hB64 + "." + pB64 + "." + sig
}

func i64toa(n int64) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	pos := len(buf)
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}

func TestAuthenticateValid(t *testing.T) {
	a := NewAuthenticator("secret", "test-iss")
	exp := time.Now().Add(1 * time.Hour).Unix()
	token := makeToken("secret", "test-iss", "user1", exp)
	identity, err := a.Authenticate(context.Background(), &mockRequest{
		headers: map[string]string{"Authorization": "Bearer " + token},
	})
	if err != nil {
		t.Fatal(err)
	}
	if identity.Principal != "user1" {
		t.Fatalf("expected user1, got %s", identity.Principal)
	}
}

func TestAuthenticateMissingHeader(t *testing.T) {
	a := NewAuthenticator("secret", "test-iss")
	_, err := a.Authenticate(context.Background(), &mockRequest{
		headers: map[string]string{},
	})
	if err == nil {
		t.Fatal("expected error for missing header")
	}
}

func TestAuthenticateInvalidFormat(t *testing.T) {
	a := NewAuthenticator("secret", "test-iss")
	_, err := a.Authenticate(context.Background(), &mockRequest{
		headers: map[string]string{"Authorization": "NotBearer token"},
	})
	if err == nil {
		t.Fatal("expected error for invalid format")
	}
}

func TestAuthenticateExpiredToken(t *testing.T) {
	a := NewAuthenticator("secret", "test-iss")
	exp := time.Now().Add(-1 * time.Hour).Unix()
	token := makeToken("secret", "test-iss", "user1", exp)
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

func TestAuthenticateInvalidSignature(t *testing.T) {
	a := NewAuthenticator("secret", "test-iss")
	exp := time.Now().Add(1 * time.Hour).Unix()
	token := makeToken("wrongsecret", "test-iss", "user1", exp)
	_, err := a.Authenticate(context.Background(), &mockRequest{
		headers: map[string]string{"Authorization": "Bearer " + token},
	})
	if err == nil {
		t.Fatal("expected error for invalid signature")
	}
}

func TestAuthenticateInvalidIssuer(t *testing.T) {
	a := NewAuthenticator("secret", "expected-iss")
	exp := time.Now().Add(1 * time.Hour).Unix()
	token := makeToken("secret", "wrong-iss", "user1", exp)
	_, err := a.Authenticate(context.Background(), &mockRequest{
		headers: map[string]string{"Authorization": "Bearer " + token},
	})
	if err == nil {
		t.Fatal("expected error for invalid issuer")
	}
}
