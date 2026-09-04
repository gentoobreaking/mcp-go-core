package transport

import (
	"testing"
)

func TestNewSessionID(t *testing.T) {
	id1 := NewSessionID()
	id2 := NewSessionID()
	if id1 == "" {
		t.Fatal("expected non-empty session ID")
	}
	if id1 == id2 {
		t.Fatal("expected unique session IDs")
	}
	// Session ID should be hex-encoded
	if len(string(id1)) != 32 {
		t.Fatalf("expected 32-char hex, got %d", len(string(id1)))
	}
}

func TestSessionManagerRegister(t *testing.T) {
	sm := NewSessionManager()
	id := sm.RegisterSession()
	if id == "" {
		t.Fatal("expected non-empty session ID")
	}
	if sm.Count() != 1 {
		t.Fatalf("expected 1 session, got %d", sm.Count())
	}
}

func TestSessionManagerGet(t *testing.T) {
	sm := NewSessionManager()
	id := sm.RegisterSession()
	done := sm.GetSession(id)
	if done == nil {
		t.Fatal("expected non-nil done channel")
	}
}

func TestSessionManagerGetNotFound(t *testing.T) {
	sm := NewSessionManager()
	done := sm.GetSession("nonexistent")
	if done != nil {
		t.Fatal("expected nil for nonexistent session")
	}
}

func TestSessionManagerUnregister(t *testing.T) {
	sm := NewSessionManager()
	id := sm.RegisterSession()
	if sm.Count() != 1 {
		t.Fatal("expected 1 session")
	}
	sm.UnregisterSession(id)
	if sm.Count() != 0 {
		t.Fatal("expected 0 sessions after unregister")
	}
}

func TestSessionManagerCloseAll(t *testing.T) {
	sm := NewSessionManager()
	sm.RegisterSession()
	sm.RegisterSession()
	if sm.Count() != 2 {
		t.Fatal("expected 2 sessions")
	}
	sm.CloseAll()
	if sm.Count() != 0 {
		t.Fatal("expected 0 sessions after close all")
	}
}
