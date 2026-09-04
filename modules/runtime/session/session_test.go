package session

import (
	"context"
	"testing"
	"time"

	"github.com/project/mcp-go-core/core/lifecycle"
)

func TestNewManager(t *testing.T) {
	m := NewManager()
	if m == nil {
		t.Fatal("expected non-nil manager")
	}
	if m.Count() != 0 {
		t.Fatalf("expected 0 sessions, got %d", m.Count())
	}
}

func TestCreateSession(t *testing.T) {
	m := NewManager()

	s, err := m.Create("test-session", map[string]any{"client": "claude"})
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}

	if s.ID == "" {
		t.Fatal("expected non-empty ID")
	}
	if s.Name != "test-session" {
		t.Fatalf("expected name test-session, got %s", s.Name)
	}
	if s.State != lifecycle.Running {
		t.Fatalf("expected Running, got %s", s.State)
	}
	if m.Count() != 1 {
		t.Fatalf("expected 1 session, got %d", m.Count())
	}
}

func TestGetSession(t *testing.T) {
	m := NewManager()

	s, _ := m.Create("test-session", nil)

	got, err := m.Get(s.ID)
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if got.ID != s.ID {
		t.Fatalf("expected ID %s, got %s", s.ID, got.ID)
	}
}

func TestGetSessionNotFound(t *testing.T) {
	m := NewManager()

	_, err := m.Get("nonexistent")
	if err != ErrSessionNotFound {
		t.Fatalf("expected ErrSessionNotFound, got: %v", err)
	}
}

func TestDestroySession(t *testing.T) {
	m := NewManager()

	s, _ := m.Create("test-session", nil)

	if m.Count() != 1 {
		t.Fatalf("expected 1 session, got %d", m.Count())
	}

	if err := m.Destroy(s.ID); err != nil {
		t.Fatalf("Destroy error: %v", err)
	}

	if m.Count() != 0 {
		t.Fatalf("expected 0 sessions, got %d", m.Count())
	}
}

func TestDestroySessionNotFound(t *testing.T) {
	m := NewManager()

	err := m.Destroy("nonexistent")
	if err != ErrSessionNotFound {
		t.Fatalf("expected ErrSessionNotFound, got: %v", err)
	}
}

func TestDestroyAll(t *testing.T) {
	m := NewManager()

	_, _ = m.Create("session1", nil)
	_, _ = m.Create("session2", nil)
	_, _ = m.Create("session3", nil)

	if m.Count() != 3 {
		t.Fatalf("expected 3 sessions, got %d", m.Count())
	}

	m.DestroyAll()

	if m.Count() != 0 {
		t.Fatalf("expected 0 sessions, got %d", m.Count())
	}
}

func TestActiveSessions(t *testing.T) {
	m := NewManager()

	_, _ = m.Create("session1", nil)
	_, _ = m.Create("session2", nil)

	active := m.ActiveSessions()
	if len(active) != 2 {
		t.Fatalf("expected 2 active sessions, got %d", len(active))
	}
}

func TestCloseWithContext(t *testing.T) {
	m := NewManager()

	s, _ := m.Create("test-session", nil)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := m.Close(ctx, s.ID)
	if err != nil {
		t.Fatalf("Close error: %v", err)
	}

	if m.Count() != 0 {
		t.Fatalf("expected 0 sessions after close, got %d", m.Count())
	}
}

func TestCloseNotFound(t *testing.T) {
	m := NewManager()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := m.Close(ctx, "nonexistent")
	if err != ErrSessionNotFound {
		t.Fatalf("expected ErrSessionNotFound, got: %v", err)
	}
}

func TestSessionMetadata(t *testing.T) {
	m := NewManager()

	info := map[string]any{
		"client_id": "claude-desktop",
		"version":   "1.0.0",
	}

	s, _ := m.Create("test-session", info)

	if s.Info["client_id"] != "claude-desktop" {
		t.Fatal("expected client_id in info")
	}
	if s.Info["version"] != "1.0.0" {
		t.Fatal("expected version in info")
	}
}

func TestCloseAll(t *testing.T) {
	m := NewManager()

	_, _ = m.Create("session1", nil)
	_, _ = m.Create("session2", nil)

	m.CloseAll()

	if m.Count() != 0 {
		t.Fatalf("expected 0 sessions after CloseAll, got %d", m.Count())
	}
}
