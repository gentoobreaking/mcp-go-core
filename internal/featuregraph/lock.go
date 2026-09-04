package featuregraph

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// LockFile represents the .mcp/features.lock file.
type LockFile struct {
	FrameworkVersion string     `json:"framework_version"`
	Profile          string     `json:"profile"`
	Features         []string   `json:"features"`
	Modules          []string   `json:"modules"`
	DependencyGraph  [][]string `json:"dependency_graph"`
	GraphHash        string     `json:"graph_hash"`
	GeneratedAt      string     `json:"generated_at"`
}

// GenerateLock generates a deterministic lock file from a resolution.
func GenerateLock(resolution *Resolution, frameworkVersion string) *LockFile {
	g := resolution.Graph

	edges := make([][]string, 0)
	for _, name := range g.ListFeatures() {
		f := g.GetFeature(name)
		if f == nil {
			continue
		}
		for _, dep := range f.Dependencies {
			edges = append(edges, []string{name, dep.Name})
		}
	}
	sort.Slice(edges, func(i, j int) bool {
		if edges[i][0] != edges[j][0] {
			return edges[i][0] < edges[j][0]
		}
		return edges[i][1] < edges[j][1]
	})

	features := make([]string, 0, len(resolution.Enabled)+len(resolution.Inferred))
	features = append(features, resolution.Enabled...)
	features = append(features, resolution.Inferred...)

	allFeatures := make([]string, 0, len(features))
	seen := make(map[string]bool)
	for _, f := range features {
		if !seen[f] {
			seen[f] = true
			allFeatures = append(allFeatures, f)
		}
	}
	sort.Strings(allFeatures)

	lock := &LockFile{
		FrameworkVersion: frameworkVersion,
		Profile:          resolution.Profile,
		Features:         allFeatures,
		Modules:          g.ListModules(),
		DependencyGraph:  edges,
		GeneratedAt:      time.Now().UTC().Format(time.RFC3339),
	}
	lock.GraphHash = computeGraphHash(lock)
	return lock
}

// computeGraphHash computes a deterministic SHA256 hash.
func computeGraphHash(lock *LockFile) string {
	h := sha256.New()

	sortedFeatures := make([]string, len(lock.Features))
	copy(sortedFeatures, lock.Features)
	sort.Strings(sortedFeatures)
	for _, f := range sortedFeatures {
		h.Write([]byte(f))
		h.Write([]byte{0})
	}

	for _, edge := range lock.DependencyGraph {
		h.Write([]byte(edge[0]))
		h.Write([]byte{0})
		h.Write([]byte(edge[1]))
		h.Write([]byte{0})
	}

	h.Write([]byte(lock.Profile))
	h.Write([]byte{0})
	h.Write([]byte(lock.FrameworkVersion))
	h.Write([]byte{0})

	return hex.EncodeToString(h.Sum(nil))
}

// WriteLock writes the lock file to .mcp/features.lock.
func WriteLock(baseDir string, lock *LockFile) error {
	lockDir := filepath.Join(baseDir, ".mcp")
	if err := os.MkdirAll(lockDir, 0755); err != nil {
		return fmt.Errorf("failed to create .mcp directory: %w", err)
	}

	data, err := json.MarshalIndent(lock, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal lock file: %w", err)
	}

	path := filepath.Join(lockDir, "features.lock")
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write lock file: %w", err)
	}
	return nil
}

// ComputeLockHash computes the hash for verification.
func ComputeLockHash(features []string, edges [][]string, profile, version string) string {
	h := sha256.New()

	sortedFeatures := make([]string, len(features))
	copy(sortedFeatures, features)
	sort.Strings(sortedFeatures)
	for _, f := range sortedFeatures {
		h.Write([]byte(f))
		h.Write([]byte{0})
	}
	for _, edge := range edges {
		h.Write([]byte(edge[0]))
		h.Write([]byte{0})
		h.Write([]byte(edge[1]))
		h.Write([]byte{0})
	}
	h.Write([]byte(profile))
	h.Write([]byte{0})
	h.Write([]byte(version))
	h.Write([]byte{0})
	return hex.EncodeToString(h.Sum(nil))
}
