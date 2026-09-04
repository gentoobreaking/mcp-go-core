package builder

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"fmt"
	"time"

	"github.com/project/mcp-go-core/internal/featuregraph"
)

func setupTestGraph() *featuregraph.Graph {
	g := featuregraph.NewGraph()
	g.AddFeature(featuregraph.FeatureDescriptor{
		Name: "http_transport",
		Dependencies: []featuregraph.Dependency{
			{Name: "transport", Type: featuregraph.DependencyHard},
		},
	})
	g.AddFeature(featuregraph.FeatureDescriptor{})
	g.AddFeature(featuregraph.FeatureDescriptor{Name: "transport"})
	g.AddFeature(featuregraph.FeatureDescriptor{Name: "oauth"})
	return g
}

func TestConfigStage(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "mcp.yaml"), []byte("profile: development\n"), 0644)

	stage := &ConfigStage{}
	bctx := &BuildContext{ProjectDir: dir}
	err := stage.Run(context.Background(), bctx)
	if err != nil {
		t.Fatal(err)
	}
	if bctx.Config.Profile != "development" {
		t.Fatal("profile not loaded")
	}
}

func TestResolveStage(t *testing.T) {
	g := setupTestGraph()

	stage := &ResolveStage{Graph: g}
	bctx := &BuildContext{
		Config: &BuildConfig{
			Profile:         "development",
			ExplicitEnable:  []string{"http_transport"},
			ExplicitDisable: []string{},
		},
	}
	err := stage.Run(context.Background(), bctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(bctx.Resolution.Enabled) != 1 {
		t.Fatalf("expected 1 enabled, got %d", len(bctx.Resolution.Enabled))
	}
}

func TestLockStage(t *testing.T) {
	g := setupTestGraph()
	res, _ := featuregraph.Resolve(featuregraph.Config{
		Graph:          g,
		ExplicitEnable: []string{"http_transport"},
	})

	dir := t.TempDir()
	stage := &LockStage{FrameworkVersion: "1.0.0"}
	bctx := &BuildContext{
		ProjectDir:  dir,
		Config:      &BuildConfig{Profile: "development"},
		Resolution:  res,
	}
	err := stage.Run(context.Background(), bctx)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dir, ".mcp", "features.lock")); err != nil {
		t.Fatal("features.lock not written")
	}
}

func TestPipelineRun(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "mcp.yaml"), []byte("profile: development\n"), 0644)

	g := setupTestGraph()
	pipeline := NewPipeline()
	pipeline.Stages = []Stage{
		&ConfigStage{},
		&ResolveStage{Graph: g},
		&LockStage{FrameworkVersion: "1.0.0"},
		&ManifestStage{ApplicationName: "test", Version: "1.0.0"},
	}

	cfg := &BuildConfig{
		Profile:          "development",
		FrameworkVersion: "1.0.0",
		ExplicitEnable:   []string{"http_transport"},
		OutputBinary:     "dist/server",
	}

	// Override ProjectDir in pipeline run
	result, err := runPipelineWithDir(context.Background(), pipeline, cfg, dir)
	if err != nil {
		t.Fatal(err)
	}
	if result == nil {
		t.Fatal("expected result")
	}
	if result.Manifest == nil {
		t.Fatal("expected manifest")
	}
}

func runPipelineWithDir(ctx context.Context, p *Pipeline, cfg *BuildConfig, dir string) (*BuildResult, error) {
	bctx := &BuildContext{
		Config:       cfg,
		ProjectDir:   dir,
		GeneratedDir: filepath.Join(dir, ".mcp", "generated"),
		StartTime:    time.Now(),
	}

	for _, stage := range p.Stages {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		if err := stage.Run(ctx, bctx); err != nil {
			return nil, fmt.Errorf("stage %s failed: %w", stage.Name(), err)
		}
	}

	return &BuildResult{
		OutputPath:   bctx.OutputPath,
		Features:     bctx.Resolution.Enabled,
		Modules:      bctx.Manifest.Modules,
		BinarySize:   bctx.Manifest.BinarySize,
		Manifest:     bctx.Manifest,
		Verification: bctx.Verification,
	}, nil
}

func TestStageNames(t *testing.T) {
	stages := []Stage{
		&ConfigStage{},
		&AnalyzeStage{},
		&ResolveStage{},
		&LockStage{},
		&GenerateStage{},
		&CompileStage{},
		&VerifyStage{},
		&BenchmarkStage{},
		&ManifestStage{},
	}

	expected := []string{
		"config", "analyze", "resolve", "lock", "generate", "compile", "verify", "benchmark", "manifest",
	}
	for i, s := range stages {
		if s.Name() != expected[i] {
			t.Fatalf("stage %d: expected %s, got %s", i, expected[i], s.Name())
		}
	}
}
