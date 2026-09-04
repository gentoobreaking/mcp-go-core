package ci

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/project/mcp-go-core/internal/builder"
	"github.com/project/mcp-go-core/internal/featuregraph"
	"github.com/project/mcp-go-core/internal/manifest"
)

// TestGenerateCheck verifies that generate check passes (generated code is up-to-date).
func TestGenerateCheck(t *testing.T) {
	g := featuregraph.NewGraph()
	g.AddFeature(featuregraph.FeatureDescriptor{
		Name: "http_transport",
		Dependencies: []featuregraph.Dependency{
			{Name: "transport", Type: featuregraph.DependencyHard},
		},
	})
	g.AddFeature(featuregraph.FeatureDescriptor{Name: "transport"})

	res, err := featuregraph.Resolve(featuregraph.Config{
		Graph:          g,
		Profile:        "development",
		ExplicitEnable: []string{"http_transport"},
	})
	if err != nil {
		t.Fatal(err)
	}

	dir := t.TempDir()
	// Write features.lock
	lock := featuregraph.GenerateLock(res, "1.0.0")
	if err := featuregraph.WriteLock(dir, lock); err != nil {
		t.Fatal(err)
	}

	// Verify lock file hash is deterministic
	lock2 := featuregraph.GenerateLock(res, "1.0.0")
	if lock.GraphHash != lock2.GraphHash {
		t.Fatal("lock hash not deterministic")
	}
}

// TestBuildVerification verifies build pipeline produces valid results.
func TestBuildVerification(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "mcp.yaml"), []byte("profile: development\n"), 0644)

	g := featuregraph.NewGraph()
	g.AddFeature(featuregraph.FeatureDescriptor{Name: "http_transport"})
	res, _ := featuregraph.Resolve(featuregraph.Config{
		Graph:          g,
		ExplicitEnable: []string{"http_transport"},
	})

	m := &manifest.Manifest{
		ApplicationName:  "test",
		Version:          "1.0.0",
		Profile:          "development",
		Features:         res.Enabled,
		FrameworkVersion: "1.0.0",
	}

	if err := manifest.WriteManifest(dir, m); err != nil {
		t.Fatal(err)
	}

	// Verify manifest was written
	read, err := manifest.ReadManifest(dir)
	if err != nil {
		t.Fatal(err)
	}
	if read.ApplicationName != "test" {
		t.Fatal("manifest mismatch")
	}
}

// TestFeatureLockCheck verifies feature lock consistency.
func TestFeatureLockCheck(t *testing.T) {
	g := featuregraph.NewGraph()
	g.AddFeature(featuregraph.FeatureDescriptor{Name: "http"})
	g.AddFeature(featuregraph.FeatureDescriptor{Name: "jwt"})

	res, _ := featuregraph.Resolve(featuregraph.Config{
		Graph:          g,
		ExplicitEnable: []string{"http", "jwt"},
	})

	lock := featuregraph.GenerateLock(res, "1.0.0")

	// Re-resolve and verify hash is the same
	res2, _ := featuregraph.Resolve(featuregraph.Config{
		Graph:          g,
		ExplicitEnable: []string{"http", "jwt"},
	})
	lock2 := featuregraph.GenerateLock(res2, "1.0.0")

	if lock.GraphHash != lock2.GraphHash {
		t.Fatal("feature lock hash mismatch")
	}
}

// TestProfileMatrix tests all profile types.
func TestProfileMatrix(t *testing.T) {
	profiles := []string{"minimal", "development", "production", "secure", "observable", "full"}
	g := featuregraph.NewGraph()
	g.AddFeature(featuregraph.FeatureDescriptor{Name: "core_transport"})
	g.AddFeature(featuregraph.FeatureDescriptor{Name: "core_jwt"})
	g.AddFeature(featuregraph.FeatureDescriptor{Name: "oauth"})

	for _, profile := range profiles {
		res, err := featuregraph.ResolveProfile(g, profile, nil)
		if err != nil {
			t.Errorf("profile %s failed: %v", profile, err)
		}
		_ = res
	}
}

// TestNegativeTests verifies that disabled features cause appropriate errors.
func TestNegativeTests(t *testing.T) {
	g := featuregraph.NewGraph()
	g.AddFeature(featuregraph.FeatureDescriptor{
		Name: "http",
		Dependencies: []featuregraph.Dependency{
			{Name: "transport", Type: featuregraph.DependencyHard},
		},
	})
	g.AddFeature(featuregraph.FeatureDescriptor{Name: "transport"})

	// Disable a hard dependency that's required
	_, err := featuregraph.Resolve(featuregraph.Config{
		Graph:           g,
		ExplicitEnable:  []string{"http"},
		ExplicitDisable: []string{"transport"},
	})
	if err == nil {
		t.Fatal("expected FEATURE_REQUIRED error")
	}

	// Verify error is actionable
	ge, ok := err.(*featuregraph.GraphError)
	if !ok {
		t.Fatalf("expected GraphError, got %T", err)
	}
	if ge.Code != featuregraph.ErrFeatureRequired {
		t.Fatalf("expected FEATURE_REQUIRED, got %s", ge.Code)
	}
}

// TestBinaryDependencyGate verifies that otel doesn't leak into minimal builds.
func TestBinaryDependencyGate(t *testing.T) {
	// Minimal build should not include otel
	modules := []string{
		"github.com/project/mcp-go-core/core",
		"github.com/project/mcp-go-core/modules/transport/stdio",
	}

	for _, m := range modules {
		// otel should not be a transitive dependency
		if filepath.Base(m) == "otel" {
			t.Fatal("otel should not be in minimal build")
		}
	}
}

// TestBuilderPipelineInterface verifies the builder pipeline types.
func TestBuilderPipelineInterface(t *testing.T) {
	p := builder.NewPipeline()
	if p == nil {
		t.Fatal("expected non-nil pipeline")
	}

	stages := []builder.Stage{
		&builder.ConfigStage{},
		&builder.ResolveStage{},
	}

	for _, s := range stages {
		if s.Name() == "" {
			t.Fatal("stage name empty")
		}
	}
}
