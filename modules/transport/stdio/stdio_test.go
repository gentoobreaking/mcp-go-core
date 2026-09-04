package stdio

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
)

func TestTransportServe(t *testing.T) {
	input := bytes.NewBufferString(`{"jsonrpc":"2.0","id":1,"method":"ping","params":{}}` + "\n")
	var output bytes.Buffer

	transport := New(input, &output)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	handler := func(ctx context.Context, msg json.RawMessage) (any, error) {
		return map[string]any{"status": "ok"}, nil
	}

	err := transport.Serve(ctx, handler)
	if err != nil && err != context.Canceled {
		t.Fatalf("unexpected error: %v", err)
	}

	if output.Len() == 0 {
		t.Fatal("expected output")
	}

	var resp map[string]any
	if err := json.Unmarshal(output.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp["jsonrpc"] != "2.0" {
		t.Fatal("expected jsonrpc 2.0")
	}
}

func TestTransportSend(t *testing.T) {
	var buf bytes.Buffer
	transport := New(nil, &buf)
	err := transport.Send(map[string]any{"jsonrpc": "2.0", "id": 1, "result": "ok"})
	if err != nil {
		t.Fatal(err)
	}
	if buf.Len() == 0 {
		t.Fatal("expected output")
	}
}

func TestTransportParseError(t *testing.T) {
	input := bytes.NewBufferString(`not json` + "\n")
	var output bytes.Buffer
	transport := New(input, &output)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	handler := func(ctx context.Context, msg json.RawMessage) (any, error) {
		return nil, nil
	}

	_ = transport.Serve(ctx, handler)
	if output.Len() == 0 {
		t.Fatal("expected error response")
	}
}
