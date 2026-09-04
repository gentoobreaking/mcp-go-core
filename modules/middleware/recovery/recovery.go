// Package recovery provides panic recovery middleware for MCP servers.
// Catches panics in request handlers and converts them to structured errors.
package recovery

import (
	"fmt"

	"github.com/project/mcp-go-core/core/middleware"
	"github.com/project/mcp-go-core/core/mcperror"
)

// RecoveryError represents a panic that was caught by the recovery middleware.
type RecoveryError struct {
	Value any
}

// Error implements the error interface.
func (e *RecoveryError) Error() string {
	return fmt.Sprintf("recovered from panic: %v", e.Value)
}

// Code returns the MCP error code for recovered panics.
func (e *RecoveryError) Code() int {
	return mcperror.ErrCodeInternalError
}

// Middleware returns a middleware that catches panics and returns a structured error.
func Middleware() middleware.Middleware {
	return func(next middleware.Handler) middleware.Handler {
		return middleware.HandlerFunc(func(method string, params []byte) (result []byte, err error) {
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

// Wrap wraps a handler with panic recovery.
func Wrap(h middleware.Handler) middleware.Handler {
	return Middleware()(h)
}
