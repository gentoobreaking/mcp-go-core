package config

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func writeTestConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}
	return path
}

const validConfig = `
server:
  name: test-server
  version: "1.0.0"
  transport: stdio
  shutdownWait: 10s
features:
  advanced: true
  beta: false
logging:
  level: info
  format: json
rateLimits:
  tools/call:
    rate: 30
    burst: 30
  tools/list:
    rate: 10
    burst: 10
`

func TestConfigLoad(t *testing.T) {
	path := writeTestConfig(t, validConfig)
	c := NewConfig(path)

	if err := c.Load(); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	srv := c.GetServer()
	if srv.Name != "test-server" {
		t.Fatalf("expected server name 'test-server', got %q", srv.Name)
	}
	if srv.Version != "1.0.0" {
		t.Fatalf("expected version 1.0.0, got %q", srv.Version)
	}

	feats := c.GetFeatures()
	if feats["advanced"] != true {
		t.Fatal("expected advanced feature to be true")
	}

	limits := c.GetRateLimits()
	if limits["tools/call"].Rate != 30 {
		t.Fatal("expected tools/call rate 30")
	}
}

func TestConfigLoadInvalidKeepsPrevious(t *testing.T) {
	path := writeTestConfig(t, validConfig)
	c := NewConfig(path)

	if err := c.Load(); err != nil {
		t.Fatalf("initial load failed: %v", err)
	}

	// Overwrite with invalid YAML
	os.WriteFile(path, []byte("server: [invalid"), 0644)

	// Load should fail but keep previous config
	err := c.Load()
	if err == nil {
		t.Fatal("expected error for invalid config")
	}

	if c.LastLoadErr() == nil {
		t.Fatal("expected LastLoadErr to be recorded")
	}

	// Previous config should still be accessible
	srv := c.GetServer()
	if srv.Name != "test-server" {
		t.Fatalf("expected previous name retained, got %q", srv.Name)
	}
}

func TestConfigAtomicSwap(t *testing.T) {
	path := writeTestConfig(t, validConfig)
	c := NewConfig(path)

	if err := c.Load(); err != nil {
		t.Fatalf("initial load failed: %v", err)
	}

	var wg sync.WaitGroup
	stop := make(chan struct{})

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
					_ = c.GetServer()
					_ = c.GetFeatures()
					_ = c.GetRateLimits()
				}
			}
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			content := strings.Replace(validConfig, "test-server", "test-server-"+string(rune('A'+i%26)), 1)
			os.WriteFile(path, []byte(content), 0644)
			c.Load()
		}
		close(stop)
	}()

	wg.Wait()
}

func TestConfigHealth(t *testing.T) {
	path := writeTestConfig(t, validConfig)
	c := NewConfig(path)

	if err := c.Load(); err != nil {
		t.Fatalf("load failed: %v", err)
	}

	h := c.Health()
	if h.Path != path {
		t.Fatalf("expected path %s, got %s", path, h.Path)
	}
	if h.LastError != "" {
		t.Fatalf("expected no error, got %s", h.LastError)
	}
}

func TestConfigHealthAfterError(t *testing.T) {
	path := writeTestConfig(t, validConfig)
	c := NewConfig(path)
	c.Load()

	os.WriteFile(path, []byte("broken: ["), 0644)
	_ = c.Load()

	h := c.Health()
	if h.LastError == "" {
		t.Fatal("expected LastError populated after failed load")
	}
}

func TestWatcherAutoReload(t *testing.T) {
	path := writeTestConfig(t, validConfig)
	c := NewConfig(path)
	if err := c.Load(); err != nil {
		t.Fatalf("initial load: %v", err)
	}

	w, err := NewWatcher(c, nil)
	if err != nil {
		t.Fatalf("NewWatcher: %v", err)
	}
	defer w.Stop()

	if err := w.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}

	updated := strings.Replace(validConfig, "test-server", "updated-server", 1)
	os.WriteFile(path, []byte(updated), 0644)
	time.Sleep(500 * time.Millisecond)

	srv := c.GetServer()
	if srv.Name != "updated-server" {
		t.Fatalf("expected auto-reloaded name, got %q", srv.Name)
	}
}

func TestWatcherReload(t *testing.T) {
	path := writeTestConfig(t, validConfig)
	c := NewConfig(path)
	if err := c.Load(); err != nil {
		t.Fatalf("initial load: %v", err)
	}

	w, err := NewWatcher(c, nil)
	if err != nil {
		t.Fatalf("NewWatcher: %v", err)
	}

	updated := strings.Replace(validConfig, "test-server", "forced-reload", 1)
	os.WriteFile(path, []byte(updated), 0644)

	if err := w.Reload(); err != nil {
		t.Fatalf("forced reload: %v", err)
	}

	srv := c.GetServer()
	if srv.Name != "forced-reload" {
		t.Fatalf("expected forced-reload, got %q", srv.Name)
	}
}

var _ = yaml.Unmarshal