package protocol

import (
	"testing"
)

func TestErrorCodes(t *testing.T) {
	tests := []struct {
		name     string
		code     int
		expected string
	}{
		{"parse_error", ParsedErrorCode, ParseErrorMessage},
		{"invalid_request", InvalidRequestCode, InvalidRequestMessage},
		{"method_not_found", MethodNotFoundCode, MethodNotFoundMessage},
		{"invalid_params", InvalidParamsCode, InvalidParamsMessage},
		{"internal_error", InternalErrorCode, InternalErrorMessage},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if ErrorMessages[tt.code] != tt.expected {
				t.Fatalf("expected %s, got %s", tt.expected, ErrorMessages[tt.code])
			}
		})
	}
}

func TestNewParseError(t *testing.T) {
	e := NewParseError()
	if e.Code != ParsedErrorCode {
		t.Fatalf("expected code %d, got %d", ParsedErrorCode, e.Code)
	}
	if e.Message != ParseErrorMessage {
		t.Fatalf("expected message %s, got %s", ParseErrorMessage, e.Message)
	}
}

func TestNewInvalidRequestError(t *testing.T) {
	e := NewInvalidRequestError()
	if e.Code != InvalidRequestCode {
		t.Fatal("code mismatch")
	}
}

func TestNewMethodNotFoundError(t *testing.T) {
	e := NewMethodNotFoundError()
	if e.Code != MethodNotFoundCode {
		t.Fatal("code mismatch")
	}
}

func TestNewInvalidParamsError(t *testing.T) {
	e := NewInvalidParamsError()
	if e.Code != InvalidParamsCode {
		t.Fatal("code mismatch")
	}
}

func TestNewInternalError(t *testing.T) {
	e := NewInternalError()
	if e.Code != InternalErrorCode {
		t.Fatal("code mismatch")
	}
}

func TestPromptListParams(t *testing.T) {
	p := PromptListParams{Hint: "test"}
	if p.Hint != "test" {
		t.Fatal("hint mismatch")
	}
}

func TestResourceListParams(t *testing.T) {
	p := ResourceListParams{Hint: "resource"}
	if p.Hint != "resource" {
		t.Fatal("hint mismatch")
	}
}

func TestToolListParams(t *testing.T) {
	p := ToolListParams{Hint: "tool"}
	if p.Hint != "tool" {
		t.Fatal("hint mismatch")
	}
}

func TestPromptListResult(t *testing.T) {
	r := PromptListResult{Prompts: []Prompt{{Name: "p1"}}}
	if len(r.Prompts) != 1 {
		t.Fatal("expected 1 prompt")
	}
}

func TestResourceListResult(t *testing.T) {
	r := ResourceListResult{Resources: []Resource{{URI: "test://"}}}
	if len(r.Resources) != 1 {
		t.Fatal("expected 1 resource")
	}
}

func TestToolListResult(t *testing.T) {
	r := ToolListResult{Tools: []Tool{{Name: "t1"}}}
	if len(r.Tools) != 1 {
		t.Fatal("expected 1 tool")
	}
}

func TestResourceTemplate(t *testing.T) {
	rt := ResourceTemplate{
		URITemplate: "file:///{path}",
		Name:        "files",
	}
	if rt.URITemplate != "file:///{path}" {
		t.Fatal("URI template mismatch")
	}
}

func TestImplementation(t *testing.T) {
	impl := Implementation{Name: "mcp-server", Version: "1.0.0"}
	if impl.Name != "mcp-server" || impl.Version != "1.0.0" {
		t.Fatal("impl mismatch")
	}
}

func TestServerCapabilities(t *testing.T) {
	caps := ServerCapabilities{
		Tools: &ToolsCapability{ListAvailable: true, Call: true},
		Prompts: &PromptsCapability{ListAvailable: true, Get: true},
	}
	if !caps.Tools.ListAvailable {
		t.Fatal("tools list expected")
	}
	if !caps.Prompts.Get {
		t.Fatal("prompts get expected")
	}
}

func TestClientCapabilities(t *testing.T) {
	caps := ClientCapabilities{
		Roots:    &RootsCapability{ListAvailable: true},
		Sampling: &SamplingCapability{CreateMessage: true},
	}
	if !caps.Roots.ListAvailable {
		t.Fatal("roots list expected")
	}
	if !caps.Sampling.CreateMessage {
		t.Fatal("sampling create expected")
	}
}

func TestInitializeRequest(t *testing.T) {
	req := InitializeRequest{
		JSONRPC: "2.0",
		Method:  "initialize",
		Params: InitializeParams{
			ProtocolVersion: "2024-11-05",
			ClientInfo:      Implementation{Name: "test", Version: "1.0"},
		},
	}
	if req.Params.ProtocolVersion != "2024-11-05" {
		t.Fatal("protocol version mismatch")
	}
}

func TestInitializeResponse(t *testing.T) {
	resp := InitializeResponse{
		JSONRPC: "2.0",
		Result: InitializeResult{
			ProtocolVersion: "2024-11-05",
			ServerInfo:      Implementation{Name: "mcp", Version: "1.0"},
		},
	}
	if resp.Result.ServerInfo.Name != "mcp" {
		t.Fatal("server info mismatch")
	}
}

func TestJSONRPCMessage_MarshalRequest(t *testing.T) {
	msg := JSONRPCMessage{
		Request: &Request{
			JSONRPC: "2.0",
			ID:      ptrInt64(1),
			Method:  "tools/list",
		},
	}
	data, err := msg.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"jsonrpc":"2.0","id":1,"method":"tools/list","params":null}` {
		t.Fatalf("unexpected JSON: %s", string(data))
	}
}

func TestJSONRPCMessage_MarshalResponse(t *testing.T) {
	msg := JSONRPCMessage{
		Response: &Response{
			JSONRPC: "2.0",
			ID:      ptrInt64(1),
			Result:  "ok",
		},
	}
	data, err := msg.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty JSON")
	}
}

func TestJSONRPCMessage_UnmarshalRequest(t *testing.T) {
	jsonStr := `{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}`
	var msg JSONRPCMessage
	if err := msg.UnmarshalJSON([]byte(jsonStr)); err != nil {
		t.Fatal(err)
	}
	if msg.Request == nil {
		t.Fatal("expected Request")
	}
	if msg.Request.Method != "tools/list" {
		t.Fatal("method mismatch")
	}
}

func TestJSONRPCMessage_UnmarshalResponse(t *testing.T) {
	jsonStr := `{"jsonrpc":"2.0","id":1,"result":"ok"}`
	var msg JSONRPCMessage
	if err := msg.UnmarshalJSON([]byte(jsonStr)); err != nil {
		t.Fatal(err)
	}
	if msg.Response == nil {
		t.Fatal("expected Response")
	}
	if msg.Response.Result != "ok" {
		t.Fatal("result mismatch")
	}
}

func TestNotification(t *testing.T) {
	n := Notification{
		JSONRPC: "2.0",
		Method:  "initialized",
	}
	if n.Method != "initialized" {
		t.Fatal("method mismatch")
	}
}

func ptrInt64(v int64) *int64 { return &v }
