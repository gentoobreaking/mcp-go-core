// Package mcperror provides MCP error types with standardized error codes.
package mcperror

import (
	"fmt"
)

// MCP error category codes (string-based).
const (
	CodeProtocol      = "protocol"
	CodeValidation    = "validation"
	CodeAuth          = "auth"
	CodeAuthorization = "authorization"
	CodeTransport     = "transport"
	CodeTool          = "tool"
	CodeInternal      = "internal"
	CodeTimeout       = "timeout"
	CodeCancellation  = "cancellation"
	CodeRateLimit     = "rate_limit"
)

// JSON-RPC 2.0 standard error codes.
const (
	ErrCodeParseError     = -32700
	ErrCodeInvalidRequest = -32600
	ErrCodeMethodNotFound = -32601
	ErrCodeInvalidParams  = -32602
	ErrCodeInternalError  = -32603
)

// Standard error messages.
const (
	ErrMsgParseError       = "Parse error"
	ErrMsgInvalidRequest   = "Invalid Request"
	ErrMsgMethodNotFound   = "Method not found"
	ErrMsgInvalidParams    = "Invalid params"
	ErrMsgInternalError    = "Internal error"
)

// JSONRPCError is the JSON representation of an error.
type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Error represents an MCP error with structured code, message, and cause.
type Error struct {
	Code    string // MCP category code (e.g., "protocol", "auth")
	MsgCode int    // JSON-RPC 2.0 numeric code
	Message string
	Cause   error
	Data    interface{}
}

// Error implements the error interface.
func (e *Error) Error() string {
	msg := e.Message
	if e.Cause != nil {
		msg = fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	if e.Code != "" {
		return fmt.Sprintf("%s: %s", e.Code, msg)
	}
	return msg
}

// JSONRPCError returns the JSON-RPC 2.0 error representation.
func (e *Error) JSONRPCError() JSONRPCError {
	return JSONRPCError{
		Code:    e.MsgCode,
		Message: e.Message,
		Data:    e.Data,
	}
}

// Unwrap allows errors.Is and errors.As to work.
func (e *Error) Unwrap() error {
	return e.Cause
}

// Is implements errors.Is for comparing with standard error codes.
func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	if !ok {
		return false
	}
	return e.MsgCode == t.MsgCode || e.Code == t.Code
}

// As implements errors.As for extracting Error from wrapped errors.
func (e *Error) As(target interface{}) bool {
	p, ok := target.(**Error)
	if !ok {
		return false
	}
	*p = e
	return true
}

// NewError creates a new Error with the given code and message.
func NewError(code, message string, cause ...error) *Error {
	var c error
	if len(cause) > 0 {
		c = cause[0]
	}
	return &Error{Code: code, Message: message, Cause: c}
}

// NewJSONRPCError creates an Error with a JSON-RPC 2.0 numeric code.
func NewJSONRPCError(code int, message string, cause ...error) *Error {
	var c error
	if len(cause) > 0 {
		c = cause[0]
	}
	return &Error{MsgCode: code, Message: message, Cause: c}
}

// Standard error constructors following JSON-RPC 2.0 spec.
func NewParseError(cause ...error) *Error {
	return NewJSONRPCError(ErrCodeParseError, ErrMsgParseError, cause...)
}

func NewInvalidRequestError(cause ...error) *Error {
	return NewJSONRPCError(ErrCodeInvalidRequest, ErrMsgInvalidRequest, cause...)
}

func NewMethodNotFoundError(method string, cause ...error) *Error {
	msg := ErrMsgMethodNotFound
	if method != "" {
		msg = fmt.Sprintf("%s: %s", ErrMsgMethodNotFound, method)
	}
	return NewJSONRPCError(ErrCodeMethodNotFound, msg, cause...)
}

func NewInvalidParamsError(message string, cause ...error) *Error {
	return NewJSONRPCError(ErrCodeInvalidParams, message, cause...)
}

func NewInternalError(cause ...error) *Error {
	return NewJSONRPCError(ErrCodeInternalError, ErrMsgInternalError, cause...)
}

// Sentinel errors for errors.Is comparisons.
var (
	ErrParseError       = &Error{MsgCode: ErrCodeParseError}
	ErrInvalidRequest   = &Error{MsgCode: ErrCodeInvalidRequest}
	ErrMethodNotFound   = &Error{MsgCode: ErrCodeMethodNotFound}
	ErrInvalidParams    = &Error{MsgCode: ErrCodeInvalidParams}
	ErrInternalError    = &Error{MsgCode: ErrCodeInternalError}
)
