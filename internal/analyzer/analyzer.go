// Package analyzer infers feature sets from multiple sources.
// Inference priority: Explicit Config > Generated Metadata > Known API > Go AST
package analyzer

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/project/mcp-go-core/internal/featuregraph"
)

// KnownAPIPatterns maps known Configure() call patterns to features.
var KnownAPIPatterns = map[string][]string{
	"http.Configure(":      {"http_transport", "transport"},
	"jwt.Configure(":        {"jwt_auth", "security"},
	"stdio.Configure(":      {"stdio_transport", "transport"},
	"sessions.Configure(":    {"sessions", "storage"},
	"logging.Configure(":     {"logging_middleware", "middleware"},
	"metrics.Configure(":     {"metrics_middleware", "middleware"},
	"tracing.Configure(":     {"tracing_middleware", "middleware"},
}

// KnownImports maps known Go import paths to features.
var KnownImports = map[string][]string{
	"github.com/project/mcp-go-core/modules/transport/http":    {"http_transport", "transport"},
	"github.com/project/mcp-go-core/modules/transport/stdio":   {"stdio_transport", "transport"},
	"github.com/project/mcp-go-core/modules/security/jwt":      {"jwt_auth", "security"},
	"github.com/project/mcp-go-core/modules/security/api_key":  {"api_key_auth", "security"},
	"github.com/project/mcp-go-core/modules/storage/memory":    {"sessions", "storage"},
	"github.com/project/mcp-go-core/modules/middleware/logging": {"logging_middleware", "middleware"},
	"github.com/project/mcp-go-core/modules/middleware/metrics":  {"metrics_middleware", "middleware"},
	"github.com/project/mcp-go-core/modules/middleware/tracing":  {"tracing_middleware", "middleware"},
}

// InferredFeature represents an inferred feature with its source.
type InferredFeature struct {
	Name   string `json:"name"`
	Source string `json:"source"` // "config", "metadata", "known_api", "go_ast"
}

// AnalysisResult holds the complete analysis result.
type AnalysisResult struct {
	Features  []InferredFeature `json:"features"`
	Source    string            `json:"source"`
	Hash      string            `json:"hash"`
}

// Config holds the analyzer configuration.
type Config struct {
	ProjectDir  string
	Profile     string
	ExplicitFeatures []string
}

// Analyze performs feature inference from all sources.
// Priority: Explicit Config > Generated Metadata > Known API > Go AST
func Analyze(ctx context.Context, cfg Config) (*AnalysisResult, error) {
	result := &AnalysisResult{Source: "analysis"}

	// Collect features from all sources
	allFeatures := make(map[string]string) // feature -> source

	// 1. Explicit config (highest priority)
	for _, f := range cfg.ExplicitFeatures {
		allFeatures[f] = "config"
	}

	// 2. Generated metadata
	metadataFeatures, err := readGeneratedMetadata(cfg.ProjectDir)
	if err == nil {
		for _, f := range metadataFeatures {
			if _, exists := allFeatures[f]; !exists {
				allFeatures[f] = "metadata"
			}
		}
	}

	// 3. Known API patterns
	apiFeatures, err := scanKnownAPIs(cfg.ProjectDir)
	if err == nil {
		for _, f := range apiFeatures {
			if _, exists := allFeatures[f]; !exists {
				allFeatures[f] = "known_api"
			}
		}
	}

	// 4. Go AST imports
	astFeatures, err := scanGoImports(cfg.ProjectDir)
	if err == nil {
		for _, f := range astFeatures {
			if _, exists := allFeatures[f]; !exists {
				allFeatures[f] = "go_ast"
			}
		}
	}

	// Build sorted result
	names := make([]string, 0, len(allFeatures))
	for name := range allFeatures {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		result.Features = append(result.Features, InferredFeature{
			Name:   name,
			Source: allFeatures[name],
		})
	}

	// Compute deterministic hash
	data, _ := json.Marshal(result.Features)
	h := sha256.Sum256(data)
	result.Hash = hex.EncodeToString(h[:])

	return result, nil
}

// readGeneratedMetadata reads .mcp/generated/metadata.json
func readGeneratedMetadata(projectDir string) ([]string, error) {
	path := filepath.Join(projectDir, ".mcp", "generated", "metadata.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var meta struct {
		Features []string `json:"features"`
	}
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}
	return meta.Features, nil
}

// scanKnownAPIs scans for known Configure() patterns in Go source files.
func scanKnownAPIs(projectDir string) ([]string, error) {
	var features []string
	seen := make(map[string]bool)

	err := filepath.Walk(projectDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || !strings.HasSuffix(path, ".go") {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		source := string(content)
		for pattern, feats := range KnownAPIPatterns {
			if strings.Contains(source, pattern) {
				for _, f := range feats {
					if !seen[f] {
						seen[f] = true
						features = append(features, f)
					}
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(features)
	return features, nil
}

// scanGoImports scans Go files for known import paths.
func scanGoImports(projectDir string) ([]string, error) {
	var features []string
	seen := make(map[string]bool)

	fset := token.NewFileSet()
	err := filepath.Walk(projectDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || !strings.HasSuffix(path, ".go") {
			return nil
		}
		node, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			return nil
		}
		for _, imp := range node.Imports {
			importPath := strings.Trim(imp.Path.Value, `"`)
			if feats, ok := KnownImports[importPath]; ok {
				for _, f := range feats {
					if !seen[f] {
						seen[f] = true
						features = append(features, f)
					}
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(features)
	return features, nil
}

// WriteInferredFeatures writes inferred features to .mcp/inferred-features.json
func WriteInferredFeatures(projectDir string, result *AnalysisResult) error {
	lockDir := filepath.Join(projectDir, ".mcp")
	if err := os.MkdirAll(lockDir, 0755); err != nil {
		return fmt.Errorf("failed to create .mcp directory: %w", err)
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}

	path := filepath.Join(lockDir, "inferred-features.json")
	return os.WriteFile(path, data, 0644)
}

// ResolveFeatures combines analysis results with the feature graph.
func ResolveFeatures(g *featuregraph.Graph, result *AnalysisResult) ([]string, error) {
	var enabled []string
	for _, f := range result.Features {
		if g.GetFeature(f.Name) != nil {
			enabled = append(enabled, f.Name)
		}
	}
	sort.Strings(enabled)
	return enabled, nil
}
