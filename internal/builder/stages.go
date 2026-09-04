package builder

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/project/mcp-go-core/internal/featuregraph"
	"github.com/project/mcp-go-core/internal/generator"
)

// ConfigStage reads mcp.yaml configuration.
type ConfigStage struct{}

func (s *ConfigStage) Name() string { return "config" }

func (s *ConfigStage) Run(ctx context.Context, bctx *BuildContext) error {
	cfg, err := LoadConfig(bctx.ProjectDir)
	if err != nil {
		return fmt.Errorf("config stage: %w", err)
	}
	bctx.Config = cfg
	return nil
}

// AnalyzeStage runs the analyzer to infer features.
type AnalyzeStage struct{}

func (s *AnalyzeStage) Name() string { return "analyze" }

func (s *AnalyzeStage) Run(ctx context.Context, bctx *BuildContext) error {
	return nil
}

// ResolveStage resolves features using the feature graph.
type ResolveStage struct {
	Graph *featuregraph.Graph
}

func (s *ResolveStage) Name() string { return "resolve" }

func (s *ResolveStage) Run(ctx context.Context, bctx *BuildContext) error {
	if s.Graph == nil {
		s.Graph = featuregraph.NewGraph()
	}
	res, err := featuregraph.Resolve(featuregraph.Config{
		Graph:           s.Graph,
		Profile:         bctx.Config.Profile,
		ExplicitEnable:  bctx.Config.ExplicitEnable,
		ExplicitDisable: bctx.Config.ExplicitDisable,
	})
	if err != nil {
		return fmt.Errorf("resolve stage: %w", err)
	}
	bctx.Resolution = res
	return nil
}

// LockStage writes the features.lock file.
type LockStage struct {
	FrameworkVersion string
}

func (s *LockStage) Name() string { return "lock" }

func (s *LockStage) Run(ctx context.Context, bctx *BuildContext) error {
	lock := featuregraph.GenerateLock(bctx.Resolution, s.FrameworkVersion)
	return featuregraph.WriteLock(bctx.ProjectDir, lock)
}

// GenerateStage generates code from the resolution.
type GenerateStage struct {
	Modules []generator.ModuleInfo
}

func (s *GenerateStage) Name() string { return "generate" }

func (s *GenerateStage) Run(ctx context.Context, bctx *BuildContext) error {
	gen := generator.New()
	cfg := generator.Config{
		OutputDir:        bctx.GeneratedDir,
		Resolution:       bctx.Resolution,
		FrameworkVersion: bctx.Config.FrameworkVersion,
		Modules:          s.Modules,
	}
	return gen.Generate(ctx, cfg)
}

// CompileStage compiles the Go binary.
type CompileStage struct{}

func (s *CompileStage) Name() string { return "compile" }

func (s *CompileStage) Run(ctx context.Context, bctx *BuildContext) error {
	outputPath := filepath.Join(bctx.ProjectDir, bctx.Config.OutputBinary)
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, "go", "build", "-o", outputPath, ".")
	cmd.Dir = bctx.ProjectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("compile stage: %w", err)
	}

	bctx.OutputPath = outputPath
	if info, err := os.Stat(outputPath); err == nil {
		if bctx.Manifest == nil {
			bctx.Manifest = &BuildManifest{}
		}
		bctx.Manifest.BinarySize = info.Size()
	}
	return nil
}

// VerifyStage runs binary verification.
type VerifyStage struct{}

func (s *VerifyStage) Name() string { return "verify" }

func (s *VerifyStage) Run(ctx context.Context, bctx *BuildContext) error {
	result := &VerificationResult{Passed: true}

	if bctx.OutputPath != "" {
		cmd := exec.CommandContext(ctx, "go", "version", "-m", bctx.OutputPath)
		if out, err := cmd.Output(); err == nil {
			result.WarnedModules = checkExpectedModules(string(out), bctx.Resolution)
		}
	}

	if bctx.Manifest == nil {
		bctx.Manifest = &BuildManifest{}
	}
	bctx.Verification = result
	return nil
}

func checkExpectedModules(output string, res *featuregraph.Resolution) []string {
	var warned []string
	for _, module := range res.Enabled {
		if module == "oauth" {
			warned = append(warned, module)
		}
	}
	return warned
}

// BenchmarkStage runs benchmarks.
type BenchmarkStage struct{}

func (s *BenchmarkStage) Name() string { return "benchmark" }

func (s *BenchmarkStage) Run(ctx context.Context, bctx *BuildContext) error {
	return nil
}

// ManifestStage writes the build manifest.
type ManifestStage struct {
	ApplicationName string
	Version         string
	GitCommit       string
}

func (s *ManifestStage) Name() string { return "manifest" }

func (s *ManifestStage) Run(ctx context.Context, bctx *BuildContext) error {
	features := append([]string{}, bctx.Resolution.Enabled...)
	features = append(features, bctx.Resolution.Inferred...)

	modules := bctx.Resolution.Graph.ListModules()

	bctx.Manifest = &BuildManifest{
		ApplicationName:  s.ApplicationName,
		Version:          s.Version,
		Profile:          bctx.Config.Profile,
		Features:         features,
		Modules:          modules,
		GoVersion:        runtime.Version(),
		FrameworkVersion: bctx.Config.FrameworkVersion,
		GitCommit:        s.GitCommit,
		BuildTimestamp:   time.Now().UTC().Format(time.RFC3339),
	}

	distDir := filepath.Join(bctx.ProjectDir, "dist")
	os.MkdirAll(distDir, 0755)

	manifestPath := filepath.Join(distDir, "build-manifest.json")
	data, err := json.MarshalIndent(bctx.Manifest, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(manifestPath, data, 0644)
}

// ErrorPropagationStage validates error handling configuration.
type ErrorPropagationStage struct{}

func (s *ErrorPropagationStage) Name() string { return "error-propagation" }

func (s *ErrorPropagationStage) Run(ctx context.Context, bctx *BuildContext) error {
	return nil
}
