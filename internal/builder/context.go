// Package builder provides the build pipeline for MCP Go Core.
// Pipeline sequence: Config → Analyze → Resolve → Lock → Generate → Compile → Verify.
package builder

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/project/mcp-go-core/internal/featuregraph"
)

// BuildContext holds build-time state shared across pipeline stages.
type BuildContext struct {
	Config       *BuildConfig
	Resolution   *featuregraph.Resolution
	Manifest     *BuildManifest
	GeneratedDir string
	OutputPath   string
	ProjectDir   string
	StartTime    time.Time
	Verification *VerificationResult
}

// BuildConfig holds build configuration from mcp.yaml.
type BuildConfig struct {
	Profile          string   `json:"profile"`
	FrameworkVersion string   `json:"framework_version"`
	ExplicitEnable   []string `json:"features"`
	ExplicitDisable  []string `json:"disable_features"`
	OutputBinary     string   `json:"output_binary"`
	ModulePath       string   `json:"module"`
}

// BuildManifest holds build metadata.
type BuildManifest struct {
	ApplicationName  string   `json:"application_name"`
	Version          string   `json:"version"`
	Profile          string   `json:"profile"`
	Features         []string `json:"features"`
	Modules          []string `json:"modules"`
	GoVersion        string   `json:"go_version"`
	FrameworkVersion string   `json:"framework_version"`
	GitCommit        string   `json:"git_commit"`
	FeatureLockHash  string   `json:"feature_lock_hash"`
	BinarySize       int64    `json:"binary_size"`
	BuildTimestamp   string   `json:"build_timestamp"`
}

// BuildResult holds the result of a build.
type BuildResult struct {
	OutputPath   string
	Features     []string
	Modules      []string
	BinarySize   int64
	Duration     time.Duration
	Verification *VerificationResult
	Manifest     *BuildManifest
}

// VerificationResult holds build verification results.
type VerificationResult struct {
	Passed        bool
	Errors        []string
	WarnedModules []string
}

// Stage interface for a build pipeline stage.
type Stage interface {
	Name() string
	Run(ctx context.Context, bctx *BuildContext) error
}

// Pipeline orchestrates the build stages.
type Pipeline struct {
	Stages []Stage
}

// NewPipeline creates a new build pipeline.
func NewPipeline() *Pipeline {
	return &Pipeline{}
}

// Run executes all stages in order.
func (p *Pipeline) Run(ctx context.Context, cfg *BuildConfig) (*BuildResult, error) {
	bctx := &BuildContext{
		Config:       cfg,
		ProjectDir:   ".",
		GeneratedDir: filepath.Join(".mcp", "generated"),
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

	result := &BuildResult{
		OutputPath:   bctx.OutputPath,
		Features:     bctx.Resolution.Enabled,
		Modules:      bctx.Manifest.Modules,
		BinarySize:   bctx.Manifest.BinarySize,
		Duration:     time.Since(bctx.StartTime),
		Manifest:     bctx.Manifest,
		Verification: &VerificationResult{Passed: true},
	}

	return result, nil
}

// LoadConfig loads build configuration from mcp.yaml.
func LoadConfig(projectDir string) (*BuildConfig, error) {
	path := filepath.Join(projectDir, "mcp.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read mcp.yaml: %w", err)
	}

	cfg := &BuildConfig{
		Profile:      "development",
		ModulePath:   "github.com/project/mcp-go-core",
		OutputBinary: "dist/server",
	}

	lines := string(data)
	for _, line := range strings.Split(lines, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "profile:") {
			cfg.Profile = strings.TrimSpace(strings.TrimPrefix(line, "profile:"))
		}
		if strings.HasPrefix(line, "module:") {
			cfg.ModulePath = strings.TrimSpace(strings.TrimPrefix(line, "module:"))
		}
	}

	return cfg, nil
}
