// Package tests contains integration tests for mcp-go-core.
package tests

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/project/mcp-go-core/testutil"
)

func TestEchoServer_RoundTrip(t *testing.T) {
	echoServer := testutil.NewEchoServer()
	handler := echoServer.Handler()

	ctx := context.Background()
	msg := json.RawMessage(`{"method":"tools/list","params":{}}`)

	resp, err := handler(ctx, msg)
	if err != nil {
		t.Fatal(err)
	}

	result, ok := resp.(map[string]any)
	if !ok {
		t.Fatal("expected map response")
	}

	if result["jsonrpc"] != "2.0" {
		t.Fatal("expected jsonrpc 2.0")
	}

	if echo := result["echo"]; echo == nil {
		t.Fatal("expected echo field")
	}

	if echoServer.ReceivedCount() != 1 {
		t.Fatalf("expected 1 message, got %d", echoServer.ReceivedCount())
	}
}

func TestMockTransport_Intercept(t *testing.T) {
	tr := testutil.NewMockTransport()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go tr.Serve(ctx, nil)

	// Intercept requests
	tr.Intercept(func(msg json.RawMessage) (any, error) {
		return map[string]any{
			"jsonrpc": "2.0",
			"echo":    string(msg),
		}, nil
	})

	msg := json.RawMessage(`{"method":"test","id":1}`)
	resp, err := tr.Send(ctx, msg)
	if err != nil {
		t.Fatal(err)
	}

	if resp == nil {
		t.Fatal("expected response")
	}

	if len(tr.Requests()) != 1 {
		t.Fatalf("expected 1 request, got %d", len(tr.Requests()))
	}

	if len(tr.Responses()) != 1 {
		t.Fatalf("expected 1 response, got %d", len(tr.Responses()))
	}

	if tr.IsClosed() {
		t.Fatal("transport should not be closed")
	}
}

func TestSession_Connect(t *testing.T) {
	t.Run("session registration and validation", func(t *testing.T) {
		sess := testutil.NewTestSession(t)

		id := sess.RegisterSession()
		if id == "" {
			t.Fatal("expected non-empty session ID")
		}

		sess.AssertConnected(id)

		sess.UnregisterSession(id)
		sess.AssertDisconnected(id)

		if sess.SessionCount() != 0 {
			t.Fatalf("expected 0 sessions, got %d", sess.SessionCount())
		}
	})

	t.Run("multiple sessions", func(t *testing.T) {
		sess := testutil.NewTestSession(t)

		for i := 0; i < 5; i++ {
			_ = sess.RegisterSession()
		}

		if sess.SessionCount() != 5 {
			t.Fatalf("expected 5 sessions, got %d", sess.SessionCount())
		}

		sess.CloseAll()
		if sess.SessionCount() != 0 {
			t.Fatal("expected 0 sessions after close all")
		}
	})

	t.Run("session with handler", func(t *testing.T) {
		echoServer := testutil.NewEchoServer()
		handler := echoServer.Handler()
		sess := testutil.NewTestSession(t)

		_ = sess.RegisterSession()

		msg := json.RawMessage(`{"method":"test"}`)
		resp, err := sess.Send(handler, msg)
		if err != nil {
			t.Fatal(err)
		}

		result, ok := resp.(map[string]any)
		if !ok {
			t.Fatal("expected map response")
		}

		testutil.AssertJSONContains(t, result, "echo")
	})
}

func TestEchoServer_Async(t *testing.T) {
	echoServer := testutil.NewEchoServer()
	handler := echoServer.Handler()

	ctx := context.Background()

	for i := 0; i < 3; i++ {
		msg := json.RawMessage(`{"method":"test","id":` + json.Number(string(rune('1'+i))) + `}`)
		_, err := handler(ctx, msg)
		if err != nil {
			t.Fatal(err)
		}
	}

	if echoServer.ReceivedCount() != 3 {
		t.Fatalf("expected 3 messages, got %d", echoServer.ReceivedCount())
	}
}

func TestEchoServer_SessionData(t *testing.T) {
	echoServer := testutil.NewEchoServer()
	handler := echoServer.Handler()

	_ = handler
	_ = time.Second // keep import used
}
