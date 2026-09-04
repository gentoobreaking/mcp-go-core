// Package config provides hot-reloadable server configuration with atomic swap.
// Supports YAML config files with file-watching via fsnotify.
package config

import (
	"os"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// Config is the hot-reloadable server configuration.
type Config struct {
	mu          sync.RWMutex
	lastLoaded  time.Time
	lastLoadErr error
	path        string
	Server      ServerConfig           `yaml:"server"`
	RateLimits  map[string]LimitConfig `yaml:"rateLimits,omitempty"`
	Features    map[string]bool        `yaml:"features,omitempty"`
	Logging     LoggingConfig          `yaml:"logging"`
}

// ServerConfig holds server-level settings.
type ServerConfig struct {
	Name         string        `yaml:"name"`
	Version      string        `yaml:"version"`
	ShutdownWait time.Duration `yaml:"shutdownWait"`
	Transport    string        `yaml:"transport"`
}

// LimitConfig defines per-method rate limits.
type LimitConfig struct {
	Rate  float64 `yaml:"rate"`  // requests per second
	Burst int     `yaml:"burst"` // max burst
}

// LoggingConfig holds logging settings.
type LoggingConfig struct {
	Level  string `yaml:"level"`  // debug, info, warning, error
	Format string `yaml:"format"` // json or text
}

// NewConfig creates a Config from a YAML file path.
func NewConfig(path string) *Config {
	return &Config{
		path:       path,
		lastLoaded: time.Now(),
	}
}

// Load reads and parses the YAML config file.
// On success, atomically swaps the config. On failure, keeps previous config.
func (c *Config) Load() error {
	data, err := os.ReadFile(c.path)
	if err != nil {
		c.setLoadErr(err)
		return err
	}

	var parsed Config
	parsed.path = c.path
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		c.setLoadErr(err)
		return err
	}

	parsed.lastLoaded = time.Now()
	parsed.lastLoadErr = nil

	c.mu.Lock()
	c.Server = parsed.Server
	c.RateLimits = parsed.RateLimits
	c.Features = parsed.Features
	c.Logging = parsed.Logging
	c.lastLoaded = parsed.lastLoaded
	c.lastLoadErr = nil
	c.mu.Unlock()

	return nil
}

// setLoadErr records a load error but preserves the previously loaded config.
func (c *Config) setLoadErr(err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lastLoadErr = err
}

// LastLoaded returns the timestamp of the last successful config load.
func (c *Config) LastLoaded() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastLoaded
}

// LastLoadErr returns any error from the last load attempt.
func (c *Config) LastLoadErr() error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastLoadErr
}

// GetServer returns a copy of the server config (thread-safe).
func (c *Config) GetServer() ServerConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Server
}

// GetFeatures returns a copy of the feature flags (thread-safe).
func (c *Config) GetFeatures() map[string]bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make(map[string]bool, len(c.Features))
	for k, v := range c.Features {
		result[k] = v
	}
	return result
}

// GetRateLimits returns a copy of rate limit configs (thread-safe).
func (c *Config) GetRateLimits() map[string]LimitConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make(map[string]LimitConfig, len(c.RateLimits))
	for k, v := range c.RateLimits {
		result[k] = v
	}
	return result
}

// HealthInfo returns configuration for health endpoint reporting.
type HealthInfo struct {
	Path       string `json:"path"`
	LastLoaded string `json:"last_loaded"`
	LastError  string `json:"last_error,omitempty"`
}

// Health returns config health status (thread-safe).
func (c *Config) Health() HealthInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()
	info := HealthInfo{
		Path:       c.path,
		LastLoaded: c.lastLoaded.Format(time.RFC3339),
	}
	if c.lastLoadErr != nil {
		info.LastError = c.lastLoadErr.Error()
	}
	return info
}
