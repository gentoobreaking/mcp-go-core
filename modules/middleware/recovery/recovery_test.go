package recovery

import (
	"strings"
	"testing"

	"github.com/project/mcp-go-core/core/middleware"
	"github.com/project/mcp-go-core/core/mcperror"
)

func TestRecoveryMiddleware(t *testing.T) {
	mw := Middleware()
	inner := middleware.HandlerFunc(func(method string, params []byte) ([]byte, error) {
		panic("test panic")
	})

	wrapped := mw(inner)

	result, err := wrapped.HandleRequest("tools/call", nil)
	if err == nil {
		t.Fatal("expected error after panic")
	}
	if result != nil {
		t.Fatalf("expected nil result, got: %v", result)
	}

	recErr, ok := err.(*RecoveryError)
	if !ok {
		t.Fatalf("expected RecoveryError, got: %T", err)
	}
	if !strings.Contains(recErr.Error(), "test panic") {
		t.Fatalf("expected panic message in error, got: %s", recErr.Error())
	}
}

func TestRecoveryMiddlewareNoPanic(t *testing.T) {
	mw := Middleware()
	inner := middleware.HandlerFunc(func(method string, params []byte) ([]byte, error) {
		return []byte(`{"ok":true}`), nil
	})

	wrapped := mw(inner)

	result, err := wrapped.HandleRequest("tools/call", nil)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if string(result) != `{"ok":true}` {
		t.Fatalf("unexpected result: %s", result)
	}
}

func TestRecoveryErrorCode(t *testing.T) {
	recErr := &RecoveryError{Value: "test"}
	if recErr.Code() != mcperror.ErrCodeInternalError {
		t.Fatalf("expected CodeInternalError (%d), got: %d", mcperror.ErrCodeInternalError, recErr.Code())
	}
}

func TestRecoveryErrorString(t *testing.T) {
	recErr := &RecoveryError{Value: "something broke"}
	expected := "recovered from panic: something broke"
	if recErr.Error() != expected {
		t.Fatalf("expected '%s', got '%s'", expected, recErr.Error())
	}
}

func TestWrap(t *testing.T) {
	panicHandler := middleware.HandlerFunc(func(method string, params []byte) ([]byte, error) {
		panic("wrapped panic")
	})

	wrapped := Wrap(panicHandler)

	_, err := wrapped.HandleRequest("test", nil)
	if err == nil {
		t.Fatal("expected error")
	}

	recErr, ok := err.(*RecoveryError)
	if !ok {
		t.Fatalf("expected RecoveryError, got: %T", err)
	}
	if !strings.Contains(recErr.Error(), "wrapped panic") {
		t.Fatalf("expected panic message, got: %s", recErr.Error())
	}
}

func TestRecoveryMiddlewareWithArgs(t *testing.T) {
	mw := Middleware()
	inner := middleware.HandlerFunc(func(method string, params []byte) ([]byte, error) {
		panic(method)
	})

	wrapped := mw(inner)

	_, err := wrapped.HandleRequest("panic-method", nil)
	if err == nil {
		t.Fatal("expected error")
	}

	recErr, ok := err.(*RecoveryError)
	if !ok {
		t.Fatalf("expected RecoveryError, got: %T", err)
	}
	if !strings.Contains(recErr.Error(), "panic-method") {
		t.Fatalf("expected method name in error, got: %s", recErr.Error())
	}
}

func TestRecoveryMiddlewareNilParams(t *testing.T) {
	mw := Middleware()
	inner := middleware.HandlerFunc(func(method string, params []byte) ([]byte, error) {
		if params != nil {
			t.Fatal("expected nil params")
		}
		panic("nil test")
	})

	wrapped := mw(inner)
	_, err := wrapped.HandleRequest("test", nil)
	if err == nil {
		t.Fatal("expected error from panic")
	}
}
