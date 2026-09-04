package filesystem

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestNewStore(t *testing.T) {
	dir := t.TempDir()
	s, err := New(dir)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil store")
	}
	if s.Root() != dir {
		absDir, _ := filepath.Abs(dir)
		if s.Root() != absDir {
			t.Fatalf("expected root %s, got %s", absDir, s.Root())
		}
	}
}

func TestNewStoreEmptyRoot(t *testing.T) {
	wd, _ := os.Getwd()
	s, err := New("")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if s.Root() != wd {
		t.Fatalf("expected cwd as root")
	}
}

func TestStoreSetAndGet(t *testing.T) {
	dir := t.TempDir()
	s, _ := New(dir)
	ctx := context.Background()

	key := "test/key"
	value := []byte("hello world")

	if err := s.Set(ctx, key, value); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	got, err := s.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if string(got) != string(value) {
		t.Fatalf("expected %q, got %q", value, got)
	}
}

func TestStoreGetNotFound(t *testing.T) {
	dir := t.TempDir()
	s, _ := New(dir)
	ctx := context.Background()

	_, err := s.Get(ctx, "nonexistent")
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got: %v", err)
	}
}

func TestStoreDelete(t *testing.T) {
	dir := t.TempDir()
	s, _ := New(dir)
	ctx := context.Background()

	key := "test/key"
	s.Set(ctx, key, []byte("value"))

	if err := s.Delete(ctx, key); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err := s.Get(ctx, key)
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound after delete, got: %v", err)
	}
}

func TestStoreDeleteNonExistent(t *testing.T) {
	dir := t.TempDir()
	s, _ := New(dir)
	ctx := context.Background()

	// Should not error for non-existent key
	if err := s.Delete(ctx, "nonexistent"); err != nil {
		t.Fatalf("Delete of non-existent key should not error, got: %v", err)
	}
}

func TestStoreKeys(t *testing.T) {
	dir := t.TempDir()
	s, _ := New(dir)
	ctx := context.Background()

	s.Set(ctx, "file1.txt", []byte("content1"))
	s.Set(ctx, "file2.txt", []byte("content2"))
	s.Set(ctx, "subdir/file3.txt", []byte("content3"))

	keys, err := s.Keys(ctx)
	if err != nil {
		t.Fatalf("Keys failed: %v", err)
	}

	if len(keys) != 3 {
		t.Fatalf("expected 3 keys, got %d: %v", len(keys), keys)
	}

	keySet := make(map[string]bool)
	for _, k := range keys {
		keySet[k] = true
	}

	if !keySet["file1.txt"] {
		t.Fatal("expected file1.txt in keys")
	}
	if !keySet["file2.txt"] {
		t.Fatal("expected file2.txt in keys")
	}
	if !keySet["subdir/file3.txt"] {
		t.Fatal("expected subdir/file3.txt in keys")
	}
}

func TestStoreContextCancel(t *testing.T) {
	dir := t.TempDir()
	s, _ := New(dir)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := s.Set(ctx, "key", []byte("value"))
	if err != context.Canceled {
		t.Fatalf("expected context.Canceled, got: %v", err)
	}
}

func TestStorePathTraversal(t *testing.T) {
	dir := t.TempDir()
	s, _ := New(dir)
	ctx := context.Background()

	// Try path traversal
	_, err := s.Get(ctx, "../../../etc/passwd")
	if err == nil {
		t.Fatal("expected error for path traversal attempt")
	}
	if err != ErrInvalidPath {
		t.Fatalf("expected ErrInvalidPath, got: %v", err)
	}
}

func TestStoreNestedDirectories(t *testing.T) {
	dir := t.TempDir()
	s, _ := New(dir)
	ctx := context.Background()

	key := "a/b/c/file.txt"
	value := []byte("nested content")

	if err := s.Set(ctx, key, value); err != nil {
		t.Fatalf("Set with nested path failed: %v", err)
	}

	got, err := s.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get with nested path failed: %v", err)
	}
	if string(got) != string(value) {
		t.Fatalf("expected %q, got %q", value, got)
	}
}

func TestStoreOverwrite(t *testing.T) {
	dir := t.TempDir()
	s, _ := New(dir)
	ctx := context.Background()

	key := "test"
	s.Set(ctx, key, []byte("original"))

	err := s.Set(ctx, key, []byte("updated"))
	if err != nil {
		t.Fatalf("overwrite failed: %v", err)
	}

	got, _ := s.Get(ctx, key)
	if string(got) != "updated" {
		t.Fatalf("expected 'updated', got %q", got)
	}
}

func TestStoreRoot(t *testing.T) {
	dir := t.TempDir()
	s, _ := New(dir)

	absDir, _ := filepath.Abs(dir)
	if s.Root() != absDir {
		t.Fatalf("expected %s, got %s", absDir, s.Root())
	}
}

func TestStoreClose(t *testing.T) {
	dir := t.TempDir()
	s, _ := New(dir)

	if err := s.Close(); err != nil {
		t.Fatalf("Close should not error, got: %v", err)
	}
}
