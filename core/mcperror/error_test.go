package mcperror

import (
	"errors"
	"testing"
)

func TestErrorCodes(t *testing.T) {
	if ErrCodeParseError != -32700 {
		t.Fatal("parse error code mismatch")
	}
	if ErrCodeInvalidRequest != -32600 {
		t.Fatal("invalid request code mismatch")
	}
	if ErrCodeMethodNotFound != -32601 {
		t.Fatal("method not found code mismatch")
	}
	if ErrCodeInvalidParams != -32602 {
		t.Fatal("invalid params code mismatch")
	}
	if ErrCodeInternalError != -32603 {
		t.Fatal("internal error code mismatch")
	}
}

func TestNewParseError(t *testing.T) {
	e := NewParseError()
	if e.MsgCode != ErrCodeParseError {
		t.Fatalf("expected code %d, got %d", ErrCodeParseError, e.MsgCode)
	}
	if e.Message != ErrMsgParseError {
		t.Fatal("message mismatch")
	}
}

func TestNewInvalidRequestError(t *testing.T) {
	e := NewInvalidRequestError()
	if e.MsgCode != ErrCodeInvalidRequest {
		t.Fatal("code mismatch")
	}
}

func TestNewMethodNotFoundError(t *testing.T) {
	e := NewMethodNotFoundError("tools/list")
	if e.MsgCode != ErrCodeMethodNotFound {
		t.Fatal("code mismatch")
	}
	if e.Error() == "" {
		t.Fatal("expected error message")
	}
}

func TestNewInvalidParamsError(t *testing.T) {
	e := NewInvalidParamsError("missing required field")
	if e.MsgCode != ErrCodeInvalidParams {
		t.Fatal("code mismatch")
	}
}

func TestNewInternalError(t *testing.T) {
	e := NewInternalError()
	if e.MsgCode != ErrCodeInternalError {
		t.Fatal("code mismatch")
	}
}

func TestErrorError(t *testing.T) {
	e := NewError("auth", "unauthorized")
	if e.Error() != "auth: unauthorized" {
		t.Fatalf("unexpected: %s", e.Error())
	}
}

func TestErrorErrorWithCause(t *testing.T) {
	inner := errors.New("inner error")
	e := NewError("auth", "unauthorized", inner)
	if e.Error() != "auth: unauthorized: inner error" {
		t.Fatalf("unexpected: %s", e.Error())
	}
}

func TestJSONRPCError(t *testing.T) {
	e := NewInternalError()
	jrErr := e.JSONRPCError()
	if jrErr.Code != ErrCodeInternalError {
		t.Fatal("jsonrpc code mismatch")
	}
	if jrErr.Message != ErrMsgInternalError {
		t.Fatal("jsonrpc message mismatch")
	}
}

func TestErrorIs(t *testing.T) {
	e := NewParseError()
	if !errors.Is(e, ErrParseError) {
		t.Fatal("errors.Is should match sentinel")
	}
}

func TestErrorAs(t *testing.T) {
	e := NewParseError()
	var target *Error
	if !errors.As(e, &target) {
		t.Fatal("errors.As should work")
	}
	if target.MsgCode != ErrCodeParseError {
		t.Fatal("code mismatch after As")
	}
}

func TestNewJSONRPCError(t *testing.T) {
	e := NewJSONRPCError(ErrCodeMethodNotFound, "method not found")
	if e.MsgCode != ErrCodeMethodNotFound {
		t.Fatal("code mismatch")
	}
	if e.Code != "" {
		t.Fatal("category code should be empty")
	}
}
