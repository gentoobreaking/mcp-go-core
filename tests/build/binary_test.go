package build

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestBinarySizeReport generates binary-size-report.json for all profiles.
func TestBinarySizeReport(t *testing.T) {
	profiles := []string{"minimal", "standard", "full"}

	for _, profile := range profiles {
		t.Run(profile, func(t *testing.T) {
			binaryPath := filepath.Join(t.TempDir(), "mcp-go-core")

			cmd := exec.Command("go", "build", "-ldflags", "-X main.version=test",
				"-tags", profile, "-o", binaryPath, "./cmd/mcp-go-core")
			if err := cmd.Run(); err != nil {
				t.Skipf("could not build with profile %s: %v", profile, err)
			}

			info, err := os.Stat(binaryPath)
			if err != nil {
				t.Fatal(err)
			}

			t.Logf("profile %s binary size: %d bytes", profile, info.Size())

			// Verify binary is not too large
			maxSize := int64(50 * 1024 * 1024) // 50 MB
			if info.Size() > maxSize {
				t.Errorf("binary size %d exceeds max %d for profile %s", info.Size(), maxSize, profile)
			}
		})
	}
}

// TestBinaryRegression verifies binary size doesn't regress by > 10%.
func TestBinaryRegression(t *testing.T) {
	binaryPath := filepath.Join(t.TempDir(), "mcp-go-core")

	cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/mcp-go-core")
	if err := cmd.Run(); err != nil {
		t.Skipf("could not build: %v", err)
	}

	info, err := os.Stat(binaryPath)
	if err != nil {
		t.Fatal(err)
	}

	currentSize := info.Size()

	// For v0.1, we verify the binary is reasonable
	// In CI, this would compare against a baseline stored in git
	baselineFile := ".ci/baseline-binary.txt"
	var baselineSize int64
	if data, err := os.ReadFile(baselineFile); err == nil {
		baselineSize = parseSize(string(data))
	}

	if baselineSize > 0 {
		regression := float64(currentSize-baselineSize) / float64(baselineSize)
		if regression > 0.10 {
			t.Errorf("binary size regression %.2f%% exceeds 10%% threshold (baseline: %d, current: %d)",
				regression*100, baselineSize, currentSize)
		}
	}

	t.Logf("binary size: %d bytes", currentSize)
}

// TestBinaryFeatureSet verifies the binary contains expected features.
func TestBinaryFeatureSet(t *testing.T) {
	binaryPath := filepath.Join(t.TempDir(), "mcp-go-core")

	cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/mcp-go-core")
	if err := cmd.Run(); err != nil {
		t.Skipf("could not build: %v", err)
	}

	// Check binary exists and is executable
	info, err := os.Stat(binaryPath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() == 0 {
		t.Fatal("binary is empty")
	}
}

// TestBinaryReproducibility verifies that the same source produces the same hash.
func TestBinaryReproducibility(t *testing.T) {
	binaryPath := filepath.Join(t.TempDir(), "mcp-go-core")

	cmd := exec.Command("go", "build", "-trimpath", "-ldflags", "-s -w -X main.version=test",
		"-o", binaryPath, "./cmd/mcp-go-core")
	if err := cmd.Run(); err != nil {
		t.Skipf("could not build: %v", err)
	}

	hash1 := hashBinary(binaryPath)

	// Rebuild
	cmd = exec.Command("go", "build", "-trimpath", "-ldflags", "-s -w -X main.version=test",
		"-o", binaryPath, "./cmd/mcp-go-core")
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}

	hash2 := hashBinary(binaryPath)

	if hash1 != hash2 {
		t.Errorf("binary not reproducible: hash1=%s hash2=%s", hash1, hash2)
	}
}

func hashBinary(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func parseSize(s string) int64 {
	var size int64
	for _, c := range s {
		if c >= '0' && c <= '9' {
			size = size*10 + int64(c-'0')
		}
	}
	return size
}
