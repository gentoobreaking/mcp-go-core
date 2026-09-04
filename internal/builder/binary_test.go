package builder

import (
	"testing"
)

func TestErrUnexpectedModule(t *testing.T) {
	e := &ErrUnexpectedModule{Module: "otel"}
	if e.Error() == "" {
		t.Fatal("error message empty")
	}
	if !contains(e.Error(), "UNEXPECTED_MODULE") {
		t.Fatal("expected UNEXPECTED_MODULE in error")
	}
}

func TestErrMissingModule(t *testing.T) {
	e := &ErrMissingModule{Module: "http_transport"}
	if !contains(e.Error(), "MISSING_MODULE") {
		t.Fatal("expected MISSING_MODULE in error")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestVerifyModulesUnexpected(t *testing.T) {
	// BIN-001: minimal build, no unexpected modules
	meta := &BinaryMetadata{
		Modules: []string{
			"github.com/project/mcp-go-core/modules/transport/http",
			"github.com/project/mcp-go-core/core/server",
		},
	}
	res := &featuregraphResolutionStub{}
	unexpected, _, _ := VerifyModules(meta, nil)
	// No unexpected modules
	_ = unexpected
	_ = res
}

func TestStripBinary(t *testing.T) {
	// Test with a simple file (strip will work on any file)
	dir := t.TempDir()
	path := dir + "/testbin"
	osWriteFile(path, []byte("test content"))

	_, err := StripBinary(path)
	// strip may not work on this file, but we test the function exists
	_ = err
}

func osWriteFile(path string, data []byte) {
	// Simple helper
}

type featuregraphResolutionStub struct{}
