// Package middleware provides the middleware abstraction.
// A Middleware wraps a Handler to add cross-cutting behavior.
package middleware

import "fmt"

// Handler is the interface for processing requests.
type Handler interface {
	HandleRequest(method string, params []byte) ([]byte, error)
}

// HandlerFunc is a convenience type for Handler functions.
type HandlerFunc func(method string, params []byte) ([]byte, error)

// HandleRequest calls the underlying function.
func (f HandlerFunc) HandleRequest(method string, params []byte) ([]byte, error) {
	return f(method, params)
}

// Middleware wraps a Handler.
type Middleware func(Handler) Handler

// Logger interface for logging.
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

// LoggerFunc implements Logger using a function.
type LoggerFunc func(level, msg string, args ...any)

func (f LoggerFunc) Debug(msg string, args ...any) { f("debug", msg, args...) }
func (f LoggerFunc) Info(msg string, args ...any)  { f("info", msg, args...) }
func (f LoggerFunc) Warn(msg string, args ...any)  { f("warn", msg, args...) }
func (f LoggerFunc) Error(msg string, args ...any) { f("error", msg, args...) }

// Chain applies middlewares in order.
func Chain(h Handler, middlewares ...Middleware) Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

// Logging middleware logs requests and responses.
func Logging(l Logger) Middleware {
	return func(next Handler) Handler {
		return HandlerFunc(func(method string, params []byte) ([]byte, error) {
			l.Debug("request", "method", method)
			result, err := next.HandleRequest(method, params)
			if err != nil {
				l.Error("response", "error", err)
			} else {
				l.Debug("response", "method", method)
			}
			return result, err
		})
	}
}

// Recovery middleware catches panics and returns an error.
func Recovery() Middleware {
	return func(next Handler) Handler {
		return HandlerFunc(func(method string, params []byte) (result []byte, err error) {
			defer func() {
				if r := recover(); r != nil {
					result = nil
					err = &RecoveryError{Value: r}
				}
			}()
			return next.HandleRequest(method, params)
		})
	}
}

// RecoveryError represents a panic that was caught.
type RecoveryError struct {
	Value any
}

func (e *RecoveryError) Error() string {
	return fmt.Sprintf("recovered from panic: %v", e.Value)
}
