package mcperror

import "testing"

func TestError(t *testing.T) {
	e := NewError(CodeProtocol, "bad request")
	if e.Code != CodeProtocol {
		t.Fatal("code mismatch")
	}
	if e.Error() != "protocol: bad request" {
		t.Fatalf("unexpected error: %s", e.Error())
	}
}

func TestErrorWithCause(t *testing.T) {
	inner := NewError(CodeTool, "inner", nil)
	e := NewError(CodeInternal, "wrapper", inner)
	if e.Error() != "internal: wrapper: tool: inner" {
		t.Fatalf("unexpected error: %s", e.Error())
	}
	if e.Unwrap() != inner {
		t.Fatal("Unwrap did not return inner error")
	}
}
