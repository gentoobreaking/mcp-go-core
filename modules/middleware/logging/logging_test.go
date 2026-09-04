package logging

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/project/mcp-go-core/core/middleware"
)

func TestNewLogger(t *testing.T) {
	l := NewLogger()
	if l == nil {
		t.Fatal("expected non-nil logger")
	}
}

func TestLoggerInterface(t *testing.T) {
	var _ middleware.Logger = (*Logger)(nil)
}

func TestLoggerDebug(t *testing.T) {
	var buf bytes.Buffer
	l := NewLogger(
		WithWriter(&buf),
		WithLevel(LevelDebug),
		WithFormat(FormatText),
	)
	l.Debug("test message", "key", "value")
	output := buf.String()
	if !strings.Contains(output, "DEBUG") {
		t.Fatal("expected DEBUG level in output")
	}
	if !strings.Contains(output, "test message") {
		t.Fatal("expected message in output")
	}
	if !strings.Contains(output, "key=value") {
		t.Fatal("expected field in output")
	}
}

func TestLoggerInfo(t *testing.T) {
	var buf bytes.Buffer
	l := NewLogger(
		WithWriter(&buf),
		WithLevel(LevelInfo),
		WithFormat(FormatText),
	)
	l.Info("info message", "request", "tools/call")
	output := buf.String()
	if !strings.Contains(output, "INFO") {
		t.Fatal("expected INFO level in output")
	}
	if !strings.Contains(output, "info message") {
		t.Fatal("expected message in output")
	}
}

func TestLoggerJSONFormat(t *testing.T) {
	var buf bytes.Buffer
	l := NewLogger(
		WithWriter(&buf),
		WithLevel(LevelDebug),
		WithFormat(FormatJSON),
	)
	l.Info("json test", "method", "tools/list")
	output := buf.String()

	var entry map[string]any
	if err := json.Unmarshal(bytes.NewBufferString(output).Bytes(), &entry); err != nil {
		t.Fatalf("expected valid JSON, got: %v", err)
	}

	if entry["level"] != "INFO" {
		t.Fatalf("expected INFO level, got: %v", entry["level"])
	}
	if entry["message"] != "json test" {
		t.Fatalf("expected message, got: %v", entry["message"])
	}
	if entry["method"] != "tools/list" {
		t.Fatalf("expected method field, got: %v", entry["method"])
	}
}

func TestLoggerTextFormat(t *testing.T) {
	var buf bytes.Buffer
	l := NewLogger(
		WithWriter(&buf),
		WithLevel(LevelDebug),
		WithFormat(FormatText),
	)
	l.Info("text test", "k", "v")
	output := buf.String()

	// Should contain timestamp and level
	if len(output) < 20 {
		t.Fatal("output too short for timestamp")
	}
	if !strings.Contains(output, "INFO") {
		t.Fatal("expected INFO in text output")
	}
}

func TestLoggerWarn(t *testing.T) {
	var buf bytes.Buffer
	l := NewLogger(
		WithWriter(&buf),
		WithLevel(LevelDebug),
		WithFormat(FormatJSON),
	)
	l.Warn("warning message", "detail", "something")
	var entry map[string]any
	json.Unmarshal(buf.Bytes(), &entry)
	if entry["level"] != "WARN" {
		t.Fatalf("expected WARN, got: %v", entry["level"])
	}
}

func TestLoggerError(t *testing.T) {
	var buf bytes.Buffer
	l := NewLogger(
		WithWriter(&buf),
		WithLevel(LevelDebug),
		WithFormat(FormatJSON),
	)
	l.Error("error message", "error", "test error")
	var entry map[string]any
	json.Unmarshal(buf.Bytes(), &entry)
	if entry["level"] != "ERROR" {
		t.Fatalf("expected ERROR, got: %v", entry["level"])
	}
}

func TestLoggerLevelFilter(t *testing.T) {
	var buf bytes.Buffer
	l := NewLogger(
		WithWriter(&buf),
		WithLevel(LevelWarn),
		WithFormat(FormatJSON),
	)
	l.Debug("should not appear")
	l.Info("should not appear")
	output := buf.String()
	if output != "" {
		t.Fatalf("expected empty output, got: %s", output)
	}

	l.Warn("should appear")
	if buf.Len() == 0 {
		t.Fatal("expected WARN to pass through")
	}
}

func TestLoggerWithField(t *testing.T) {
	var buf bytes.Buffer
	l := NewLogger(
		WithWriter(&buf),
		WithLevel(LevelDebug),
		WithFormat(FormatJSON),
		WithField("service", "mcp-go-core"),
	)
	l.Info("test", "method", "tools/call")
	var entry map[string]any
	json.Unmarshal(buf.Bytes(), &entry)
	if entry["service"] != "mcp-go-core" {
		t.Fatalf("expected service field, got: %v", entry["service"])
	}
}

func TestMiddleware(t *testing.T) {
	var buf bytes.Buffer
	l := NewLogger(
		WithWriter(&buf),
		WithLevel(LevelDebug),
		WithFormat(FormatJSON),
	)

	// Create a handler chain with logging middleware
	inner := middleware.HandlerFunc(func(method string, params []byte) ([]byte, error) {
		return []byte(`{"result":"ok"}`), nil
	})

	chained := l.Middleware()(inner)

	result, err := chained.HandleRequest("tools/call", []byte(`{"name":"test"}`))
	if err != nil {
		t.Fatal(err)
	}
	if string(result) != `{"result":"ok"}` {
		t.Fatalf("unexpected result: %s", result)
	}

	// Verify log output contains method info
	output := buf.String()
	if !strings.Contains(output, "request") {
		t.Fatal("expected request log entry")
	}
	if !strings.Contains(output, "response") {
		t.Fatal("expected response log entry")
	}
}

func TestMiddlewareWithError(t *testing.T) {
	var buf bytes.Buffer
	l := NewLogger(
		WithWriter(&buf),
		WithLevel(LevelDebug),
		WithFormat(FormatJSON),
	)

	inner := middleware.HandlerFunc(func(method string, params []byte) ([]byte, error) {
		return nil, &middleware.RecoveryError{Value: "test panic"}
	})

	chained := l.Middleware()(inner)

	_, err := chained.HandleRequest("tools/call", nil)
	if err == nil {
		t.Fatal("expected error")
	}

	// Verify error was logged
	output := buf.String()
	if !strings.Contains(strings.ToLower(output), "error") {
		t.Fatal("expected error log entry")
	}
}

func TestDefaultLogger(t *testing.T) {
	l := DefaultLogger()
	if l == nil {
		t.Fatal("expected non-nil default logger")
	}
}

func TestJSONLogger(t *testing.T) {
	l := JSONLogger()
	if l == nil {
		t.Fatal("expected non-nil JSON logger")
	}
}

func TestLoggerImplementsInterface(t *testing.T) {
	l := NewLogger()
	var _ middleware.Logger = l
	l.Debug("test", "k", "v")
	l.Info("test", "k", "v")
	l.Warn("test", "k", "v")
	l.Error("test", "k", "v")
}
