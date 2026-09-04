package manifest

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteAndReadManifest(t *testing.T) {
	dir := t.TempDir()
	m := &Manifest{
		ApplicationName:  "test-app",
		Version:          "1.0.0",
		Profile:          "development",
		Features:         []string{"http_transport"},
		Modules:          []string{"transport/http"},
		FrameworkVersion: "1.0.0",
	}

	if err := WriteManifest(dir, m); err != nil {
		t.Fatal(err)
	}

	// Verify file exists
	if _, err := os.Stat(filepath.Join(dir, "dist", "build-manifest.json")); err != nil {
		t.Fatal("manifest not written")
	}

	// Read back
	read, err := ReadManifest(dir)
	if err != nil {
		t.Fatal(err)
	}
	if read.ApplicationName != "test-app" {
		t.Fatal("application name mismatch")
	}
}

func TestCopyFeaturesLock(t *testing.T) {
	dir := t.TempDir()

	// Create source lock file
	os.MkdirAll(filepath.Join(dir, ".mcp"), 0755)
	os.WriteFile(filepath.Join(dir, ".mcp", "features.lock"), []byte("{}"), 0644)

	if err := CopyFeaturesLock(dir); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dir, "dist", "features.lock")); err != nil {
		t.Fatal("features.lock not copied")
	}
}

func TestWriteChecksums(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "dist"), 0755)
	os.WriteFile(filepath.Join(dir, "dist", "server"), []byte("test binary"), 0644)

	if err := WriteChecksums(dir, "server"); err != nil {
		t.Fatal(err)
	}

	// Verify checksum file exists
	if _, err := os.Stat(filepath.Join(dir, "dist", "checksums.txt")); err != nil {
		t.Fatal("checksums.txt not written")
	}
}

func TestVerifyChecksums(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "dist"), 0755)
	os.WriteFile(filepath.Join(dir, "dist", "server"), []byte("test binary"), 0644)

	if err := WriteChecksums(dir, "server"); err != nil {
		t.Fatal(err)
	}

	// Verify should pass
	if err := VerifyChecksums(dir, "server"); err != nil {
		t.Fatal(err)
	}

	// Modify binary and verify should fail
	os.WriteFile(filepath.Join(dir, "dist", "server"), []byte("modified"), 0644)
	err := VerifyChecksums(dir, "server")
	if err == nil {
		t.Fatal("expected checksum verification to fail")
	}
}
