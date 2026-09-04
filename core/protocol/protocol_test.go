package protocol

import (
	"encoding/json"
	"testing"
)

func TestMessageRoundTrip(t *testing.T) {
	params := json.RawMessage(`{}`)
	req := Request{
		JSONRPC: "2.0",
		ID:      int64Ptr(1),
		Method:  "tools/list",
		Params:  params,
	}
	if req.JSONRPC != JSONRPCVersion {
		t.Fatalf("expected %s, got %s", JSONRPCVersion, req.JSONRPC)
	}
	if req.Method != "tools/list" {
		t.Fatal("method mismatch")
	}
}

func int64Ptr(v int64) *int64 { return &v }
