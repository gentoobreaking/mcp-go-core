package middleware

import (
	"context"
	"fmt"
	"testing"
)

type mockHandler struct {
	called bool
}

func (h *mockHandler) HandleRequest(method string, params []byte) ([]byte, error) {
	h.called = true
	return []byte("handled"), nil
}

func TestChain(t *testing.T) {
	h := &mockHandler{}
	mw := func(next Handler) Handler {
		return HandlerFunc(func(method string, params []byte) ([]byte, error) {
			result, err := next.HandleRequest(method, params)
			return append(result, []byte(":wrapped")...), err
		})
	}

	chained := Chain(h, mw)
	resp, err := chained.HandleRequest("test", nil)
	if err != nil {
		t.Fatal(err)
	}
	if string(resp) != "handled:wrapped" {
		t.Fatalf("expected 'handled:wrapped', got %s", string(resp))
	}
	if !h.called {
		t.Fatal("handler not called")
	}
}

func TestLogging(t *testing.T) {
	h := &mockHandler{}
	logger := LoggerFunc(func(level, msg string, args ...any) {
		fmt.Printf("%s: %s\n", level, msg)
	})
	mw := Logging(logger)
	chained := Chain(h, mw)
	_, _ = chained.HandleRequest("test", nil)
}

func TestRecovery(t *testing.T) {
	panicHandler := HandlerFunc(func(method string, params []byte) ([]byte, error) {
		panic("test panic")
	})
	mw := Recovery()
	chained := Chain(panicHandler, mw)

	// Should not panic
	_, err := chained.HandleRequest("test", nil)
	if err == nil {
		t.Fatal("expected error from panic recovery")
	}
}

func TestLoggerFunc(t *testing.T) {
	var captured string
	f := LoggerFunc(func(level, msg string, args ...any) {
		captured = level + ":" + msg
	})
	f.Debug("msg", "key", "val")
	f.Info("info", "key", "val")
	f.Warn("warn")
	f.Error("err")
	if captured != "error:err" {
		t.Fatalf("last call should be error:err, got %s", captured)
	}
}

func TestContextLogger(t *testing.T) {
	_ = context.Background()
}
