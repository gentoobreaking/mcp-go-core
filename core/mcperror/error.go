// Package mcperror provides MCP error types with standardized error codes.
package mcperror

import "fmt"

// Error codes as defined by MCP specification.
const (
	CodeProtocol       = "protocol"
	CodeValidation     = "validation"
	CodeAuth           = "auth"
	CodeAuthorization  = "authorization"
	CodeTransport      = "transport"
	CodeTool           = "tool"
	CodeInternal       = "internal"
	CodeTimeout        = "timeout"
	CodeCancellation   = "cancellation"
)

// Error represents an MCP error with structured code, message, and cause.
type Error struct {
	Code    string
	Message string
	Cause   error
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap allows errors.Is and errors.As to work.
func (e *Error) Unwrap() error {
	return e.Cause
}

// NewError creates a new Error with the given code and message.
func NewError(code, message string, cause ...error) *Error {
	var c error
	if len(cause) > 0 {
		c = cause[0]
	}
	return &Error{Code: code, Message: message, Cause: c}
}
