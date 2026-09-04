package lifecycle

import (
	"context"
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	m := NewManager()
	if m.CurrentState() != Created {
		t.Fatalf("expected Created, got %s", m.CurrentState())
	}
}

func TestValidTransitions(t *testing.T) {
	m := NewManager()
	if err := m.Transition(Configured); err != nil {
		t.Fatal(err)
	}
	if m.CurrentState() != Configured {
		t.Fatal("not configured")
	}
	if err := m.Transition(Initialized); err != nil {
		t.Fatal(err)
	}
}

func TestInvalidTransition(t *testing.T) {
	m := NewManager()
	if err := m.Transition(Shutdown); err == nil {
		t.Fatal("expected error for invalid transition")
	}
}

func TestShutdown(t *testing.T) {
	m := NewManager()
	m.Transition(Configured)
	m.Transition(Initialized)
	m.Transition(Started)
	m.Transition(Running)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := m.Shutdown(ctx); err != nil {
		t.Fatal(err)
	}
	if m.CurrentState() != Shutdown {
		t.Fatal("not shutdown")
	}
}
