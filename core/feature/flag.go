// Package feature provides runtime feature flag management for MCP servers.
// Flags are backed by configuration and can be evaluated at runtime,
// with middleware gating dispatch to disabled tools/resources/prompts.
package feature

import (
	"sync"
)

// Flag represents the state of a single feature flag.
type Flag struct {
	Enabled bool `json:"enabled"`
}

// Flags is a thread-safe store of feature flags.
type Flags struct {
	mu    sync.RWMutex
	flags map[string]Flag
}

// NewFlags creates a new Flags store with the given initial flags.
func NewFlags(initial map[string]bool) *Flags {
	f := &Flags{flags: make(map[string]Flag, len(initial))}
	for name, enabled := range initial {
		f.flags[name] = Flag{Enabled: enabled}
	}
	return f
}

// Get returns the Flag for the given name. Zero-value Flag if not found.
func (f *Flags) Get(name string) Flag {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.flags[name]
}

// Set updates a flag at runtime. Returns true if the flag existed.
func (f *Flags) Set(name string, enabled bool) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	_, exists := f.flags[name]
	f.flags[name] = Flag{Enabled: enabled}
	return exists
}

// IsDisabled returns true if the flag exists and is disabled.
// If the flag doesn't exist, it's considered disabled (fail-closed).
func (f *Flags) IsDisabled(name string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	flag, ok := f.flags[name]
	if !ok {
		return true // unknown flag → disabled
	}
	return !flag.Enabled
}

// EnabledList returns names of all enabled flags.
func (f *Flags) EnabledList() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	list := make([]string, 0, len(f.flags))
	for name, flag := range f.flags {
		if flag.Enabled {
			list = append(list, name)
		}
	}
	return list
}

// Snapshot returns a copy of all flags for health/status reporting.
func (f *Flags) Snapshot() map[string]Flag {
	f.mu.RLock()
	defer f.mu.RUnlock()
	snap := make(map[string]Flag, len(f.flags))
	for k, v := range f.flags {
		snap[k] = v
	}
	return snap
}
