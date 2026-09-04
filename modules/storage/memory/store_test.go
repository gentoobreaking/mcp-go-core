package memory

import (
	"context"
	"testing"
)

func TestStoreSetGet(t *testing.T) {
	store := New()
	ctx := context.Background()

	if err := store.Set(ctx, "test", []byte("value")); err != nil {
		t.Fatal(err)
	}

	val, err := store.Get(ctx, "test")
	if err != nil {
		t.Fatal(err)
	}
	if string(val) != "value" {
		t.Fatalf("expected 'value', got '%s'", string(val))
	}
}

func TestStoreGetNotFound(t *testing.T) {
	store := New()
	_, err := store.Get(context.Background(), "nonexistent")
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestStoreDelete(t *testing.T) {
	store := New()
	ctx := context.Background()

	store.Set(ctx, "test", []byte("value"))
	store.Delete(ctx, "test")

	_, err := store.Get(ctx, "test")
	if err != ErrNotFound {
		t.Fatal("expected ErrNotFound after delete")
	}
}

func TestStoreKeys(t *testing.T) {
	store := New()
	ctx := context.Background()
	store.Set(ctx, "a", []byte("1"))
	store.Set(ctx, "b", []byte("2"))

	keys, err := store.Keys(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keys))
	}
}
