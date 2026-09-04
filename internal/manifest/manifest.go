// Package manifest provides build manifest operations.
package manifest

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Manifest holds build metadata.
type Manifest struct {
	ApplicationName     string   `json:"application_name"`
	Version             string   `json:"version"`
	Profile             string   `json:"profile"`
	Features            []string `json:"features"`
	Modules             []string `json:"modules"`
	GoVersion           string   `json:"go_version"`
	FrameworkVersion    string   `json:"framework_version"`
	GitCommit           string   `json:"git_commit"`
	FeatureLockHash     string   `json:"feature_lock_hash"`
	BinarySize          int64    `json:"binary_size"`
	BuildTimestamp      string   `json:"build_timestamp"`
}

// WriteManifest writes the manifest to dist/build-manifest.json.
func WriteManifest(projectDir string, m *Manifest) error {
	distDir := filepath.Join(projectDir, "dist")
	if err := os.MkdirAll(distDir, 0755); err != nil {
		return err
	}

	m.BuildTimestamp = time.Now().UTC().Format(time.RFC3339)

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(distDir, "build-manifest.json"), data, 0644)
}

// CopyFeaturesLock copies features.lock to dist/.
func CopyFeaturesLock(projectDir string) error {
	src := filepath.Join(projectDir, ".mcp", "features.lock")
	distDir := filepath.Join(projectDir, "dist")
	if err := os.MkdirAll(distDir, 0755); err != nil {
		return err
	}
	dst := filepath.Join(distDir, "features.lock")

	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

// WriteChecksums writes sha256 checksums to dist/checksums.txt.
func WriteChecksums(projectDir string, binaryName string) error {
	distDir := filepath.Join(projectDir, "dist")
	binaryPath := filepath.Join(distDir, binaryName)

	data, err := os.ReadFile(binaryPath)
	if err != nil {
		return err
	}

	h := sha256.Sum256(data)
	checksum := hex.EncodeToString(h[:]) + "  " + binaryName + "\n"

	return os.WriteFile(filepath.Join(distDir, "checksums.txt"), []byte(checksum), 0644)
}

// ReadManifest reads a build manifest.
func ReadManifest(projectDir string) (*Manifest, error) {
	path := filepath.Join(projectDir, "dist", "build-manifest.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// VerifyChecksums verifies the binary against the checksum file.
func VerifyChecksums(projectDir string, binaryName string) error {
	checksumPath := filepath.Join(projectDir, "dist", "checksums.txt")
	data, err := os.ReadFile(checksumPath)
	if err != nil {
		return err
	}

	binaryPath := filepath.Join(projectDir, "dist", binaryName)
	binary, err := os.ReadFile(binaryPath)
	if err != nil {
		return err
	}

	h := sha256.Sum256(binary)
	expected := hex.EncodeToString(h[:])

	lines := string(data)
	for _, line := range splitLines(lines) {
		if len(line) > 0 && line[:64] == expected {
			return nil
		}
	}
	return fmt.Errorf("checksum mismatch")
}

func splitLines(s string) []string {
	var lines []string
	current := ""
	for _, c := range s {
		if c == '\n' {
			lines = append(lines, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}
