// Package filesystem provides filesystem-based storage for MCP servers.
// Stores resources as files on disk with URI-to-path mapping.
package filesystem

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	ErrNotFound = errors.New("key not found")
	ErrInvalidPath = errors.New("invalid path: potential path traversal detected")
)

// Store implements filesystem-based storage.
type Store struct {
	mu   sync.RWMutex
	root string
	mode fs.FileMode
}

// New creates a new filesystem store rooted at the given directory.
// If root is empty, the current working directory is used.
func New(root string) (*Store, error) {
	if root == "" {
		root = "."
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve root path: %w", err)
	}

	if err := os.MkdirAll(absRoot, 0755); err != nil {
		return nil, fmt.Errorf("create root directory: %w", err)
	}

	return &Store{
		root: absRoot,
		mode: 0644,
	}, nil
}

// keyToPath converts a storage key to a safe filesystem path.
func (s *Store) keyToPath(key string) (string, error) {
	// Sanitize: remove leading slashes
	cleanKey := strings.TrimPrefix(key, "/")

	// Reject path traversal attempts
	if strings.Contains(cleanKey, "..") {
		return "", ErrInvalidPath
	}

	path := filepath.Join(s.root, cleanKey)

	// Verify the resolved path is within root
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	rel, err := filepath.Rel(s.root, absPath)
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(rel, "..") {
		return "", ErrInvalidPath
	}

	return absPath, nil
}

// Get retrieves a value by key.
func (s *Store) Get(ctx context.Context, key string) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	path, err := s.keyToPath(key)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) || errors.Is(err, fs.ErrNotExist) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("read file: %w", err)
	}

	return data, nil
}

// Set stores a value by key.
func (s *Store) Set(ctx context.Context, key string, value []byte) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	path, err := s.keyToPath(key)
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	s.mu.RLock()
	mode := s.mode
	s.mu.RUnlock()

	if err := os.WriteFile(path, value, mode); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}

// Delete removes a key.
func (s *Store) Delete(ctx context.Context, key string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	path, err := s.keyToPath(key)
	if err != nil {
		return err
	}

	err = os.Remove(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) || errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("remove file: %w", err)
	}

	return nil
}

// Keys returns all stored keys (relative to root).
func (s *Store) Keys(ctx context.Context) ([]string, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	s.mu.RLock()
	root := s.root
	s.mu.RUnlock()

	var keys []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		// Convert to forward slashes for consistency
		key := filepath.ToSlash(rel)
		keys = append(keys, key)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("walk directory: %w", err)
	}

	return keys, nil
}

// Root returns the absolute root path.
func (s *Store) Root() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.root
}

// Close is a no-op for filesystem storage.
func (s *Store) Close() error {
	return nil
}
