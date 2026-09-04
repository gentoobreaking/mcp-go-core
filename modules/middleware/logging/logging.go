// Package logging provides structured logging middleware for MCP servers.
// Supports both standard text format (RFC 3339 timestamps) and JSON output.
package logging

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/project/mcp-go-core/core/middleware"
)

// Level represents a log level.
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Format represents the log output format.
type Format int

const (
	FormatText Format = iota
	FormatJSON
)

// Logger is a structured logger implementing core/middleware.Logger.
type Logger struct {
	mu      sync.RWMutex
	w       io.Writer
	level   Level
	format  Format
	fields  map[string]any
}

// Option configures a Logger.
type Option func(*Logger)

// WithWriter sets the output writer.
func WithWriter(w io.Writer) Option {
	return func(l *Logger) { l.w = w }
}

// WithLevel sets the minimum log level.
func WithLevel(lvl Level) Option {
	return func(l *Logger) { l.level = lvl }
}

// WithFormat sets the output format.
func WithFormat(f Format) Option {
	return func(l *Logger) { l.format = f }
}

// WithField adds a structured field to all log entries.
func WithField(key string, value any) Option {
	return func(l *Logger) {
		l.fields[key] = value
	}
}

// NewLogger creates a new structured Logger.
func NewLogger(opts ...Option) *Logger {
	l := &Logger{
		w:      os.Stderr,
		level:  LevelInfo,
		format: FormatText,
		fields: make(map[string]any),
	}
	for _, opt := range opts {
		opt(l)
	}
	return l
}

// Debug logs a debug message.
func (l *Logger) Debug(msg string, args ...any) {
	l.log(LevelDebug, msg, args...)
}

// Info logs an info message.
func (l *Logger) Info(msg string, args ...any) {
	l.log(LevelInfo, msg, args...)
}

// Warn logs a warning message.
func (l *Logger) Warn(msg string, args ...any) {
	l.log(LevelWarn, msg, args...)
}

// Error logs an error message.
func (l *Logger) Error(msg string, args ...any) {
	l.log(LevelError, msg, args...)
}

// log writes a log entry with the given level and message.
func (l *Logger) log(lvl Level, msg string, args ...any) {
	l.mu.RLock()
	format := l.format
	threshold := l.level
	fields := make(map[string]any, len(l.fields))
	for k, v := range l.fields {
		fields[k] = v
	}
	l.mu.RUnlock()

	if lvl < threshold {
		return
	}

	// Parse key-value pairs from args
	for i := 0; i+1 < len(args); i += 2 {
		key, ok := args[i].(string)
		if !ok {
			continue
		}
		fields[key] = args[i+1]
	}

	now := time.Now().UTC().Format(time.RFC3339)

	switch format {
	case FormatJSON:
		l.writeJSON(lvl, now, msg, fields)
	default:
		l.writeText(lvl, now, msg, fields)
	}
}

// writeText writes a standard-format log line with RFC 3339 timestamps.
func (l *Logger) writeText(lvl Level, timestamp, msg string, fields map[string]any) {
	var sb strings.Builder
	sb.WriteString(timestamp)
	sb.WriteString(" [")
	sb.WriteString(lvl.String())
	sb.WriteString("] ")
	sb.WriteString(msg)

	for k, v := range fields {
		sb.WriteString(" ")
		sb.WriteString(k)
		sb.WriteString("=")
		sb.WriteString(fmt.Sprintf("%v", v))
	}
	sb.WriteString("\n")

	w := l.getWriter()
	w.Write([]byte(sb.String()))
}

// writeJSON writes a JSON-structured log entry.
func (l *Logger) writeJSON(lvl Level, timestamp, msg string, fields map[string]any) {
	entry := map[string]any{
		"timestamp": timestamp,
		"level":     lvl.String(),
		"message":   msg,
	}
	for k, v := range fields {
		entry[k] = v
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return
	}
	data = append(data, '\n')

	w := l.getLogger().getWriter()
	w.Write(data)
}

// getWriter returns the current writer (thread-safe).
func (l *Logger) getWriter() io.Writer {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.w
}

// getLogger returns the logger for internal calls.
func (l *Logger) getLogger() *Logger {
	return l
}

// Middleware returns a middleware that logs requests and responses.
func (l *Logger) Middleware() middleware.Middleware {
	return middleware.Logging(l)
}

// DefaultLogger returns a Logger configured for standard text output.
func DefaultLogger() *Logger {
	return NewLogger(
		WithWriter(os.Stderr),
		WithLevel(LevelInfo),
		WithFormat(FormatText),
	)
}

// JSONLogger returns a Logger configured for JSON structured output.
func JSONLogger() *Logger {
	return NewLogger(
		WithWriter(os.Stderr),
		WithLevel(LevelDebug),
		WithFormat(FormatJSON),
	)
}
