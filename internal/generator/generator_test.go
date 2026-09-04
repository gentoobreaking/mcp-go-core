package generator

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

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
	g.AddFeature(featuregraph.FeatureDescriptor{Name: "transport"})
	g.AddFeature(featuregraph.FeatureDescriptor{Name: "jwt_auth"})
	g.AddFeature(featuregraph.FeatureDescriptor{Name: "oauth"})
	return g
}

func TestGenerateFeaturesGo(t *testing.T) {
	dir := t.TempDir()
	g := setupTestGraph()
	res, _ := featuregraph.Resolve(featuregraph.Config{
		Graph:          g,
		ExplicitEnable: []string{"http_transport", "jwt_auth"},
	})

	cfg := Config{
		OutputDir:        dir,
		Resolution:       res,
		FrameworkVersion: "1.0.0",
		Modules: []ModuleInfo{
			{Name: "http", Import: "github.com/project/mcp-go-core/modules/transport/http"},
		},
	}

	gen := New()
	if err := gen.Generate(context.Background(), cfg); err != nil {
		t.Fatal(err)
	}

	// Verify features.go exists
	data, err := os.ReadFile(filepath.Join(dir, "features.go"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "FeatureHttpTransport") {
		t.Fatal("expected feature constant")
	}
}

func TestGenerateModulesGoNotContainConfigureAll(t *testing.T) {
	dir := t.TempDir()
	g := setupTestGraph()
	res, _ := featuregraph.Resolve(featuregraph.Config{
		Graph:          g,
		ExplicitEnable: []string{"http_transport"},
	})

	cfg := Config{
		OutputDir:        dir,
		Resolution:       res,
		FrameworkVersion: "1.0.0",
		Modules: []ModuleInfo{
			{Name: "http", Import: "github.com/project/mcp-go-core/modules/transport/http"},
		},
	}

	gen := New()
	if err := gen.Generate(context.Background(), cfg); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "modules.go"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	// GEN-003: Should call module.Configure() directly, not ConfigureAll
	if !strings.Contains(content, "http.Configure()") {
		t.Fatal("expected direct module.Configure() call")
	}

	// GEN-002: oauth disabled → oauth import NOT present
	if strings.Contains(content, "oauth") {
		t.Fatal("oauth should not be in generated code when disabled")
	}
}

func TestGenerateDeterministic(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()
	g := setupTestGraph()
	res, _ := featuregraph.Resolve(featuregraph.Config{
		Graph:          g,
		ExplicitEnable: []string{"http_transport"},
	})

	cfg := Config{
		OutputDir:        dir1,
		Resolution:       res,
		FrameworkVersion: "1.0.0",
		Modules: []ModuleInfo{
			{Name: "http", Import: "github.com/project/mcp-go-core/modules/transport/http"},
		},
	}

	gen := New()
	gen.Generate(context.Background(), cfg)
	data1, _ := os.ReadFile(filepath.Join(dir1, "features.go"))

	cfg2 := cfg
	cfg2.OutputDir = dir2
	gen.Generate(context.Background(), cfg2)
	data2, _ := os.ReadFile(filepath.Join(dir2, "features.go"))

	if string(data1) != string(data2) {
		t.Fatal("generated code is not deterministic")
	}
}

func TestChecksumDeterministic(t *testing.T) {
	g := setupTestGraph()
	res, _ := featuregraph.Resolve(featuregraph.Config{
		Graph:          g,
		ExplicitEnable: []string{"http_transport"},
	})

	cfg := Config{
		Resolution:       res,
		FrameworkVersion: "1.0.0",
		Modules: []ModuleInfo{
			{Name: "http", Import: "github.com/project/mcp-go-core/modules/transport/http"},
		},
	}

	h1, _ := Checksum(cfg)
	h2, _ := Checksum(cfg)
	if h1 != h2 {
		t.Fatal("checksum not deterministic")
	}
}

func TestGenerateBuildInfo(t *testing.T) {
	dir := t.TempDir()
	g := setupTestGraph()
	res, _ := featuregraph.Resolve(featuregraph.Config{
		Graph:          g,
		ExplicitEnable: []string{"http_transport"},
	})

	cfg := Config{
		OutputDir:        dir,
		Resolution:       res,
		FrameworkVersion: "1.0.0",
	}

	gen := New()
	if err := gen.Generate(context.Background(), cfg); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "buildinfo.go"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "1.0.0") {
		t.Fatal("expected framework version in buildinfo")
	}
	if !strings.Contains(content, "GraphHash") {
		t.Fatal("expected GraphHash in buildinfo")
	}
}

func TestGenerateServerGo(t *testing.T) {
	dir := t.TempDir()
	g := setupTestGraph()
	res, _ := featuregraph.Resolve(featuregraph.Config{
		Graph:          g,
		ExplicitEnable: []string{"http_transport"},
	})

	cfg := Config{
		OutputDir:        dir,
		Resolution:       res,
		FrameworkVersion: "1.0.0",
	}

	gen := New()
	if err := gen.Generate(context.Background(), cfg); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "server.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "NewServer") {
		t.Fatal("expected NewServer in server.go")
	}
}

func TestImportOrdering(t *testing.T) {
	dir := t.TempDir()
	g := setupTestGraph()
	res, _ := featuregraph.Resolve(featuregraph.Config{
		Graph:          g,
		ExplicitEnable: []string{"http_transport"},
	})

	cfg := Config{
		OutputDir:        dir,
		Resolution:       res,
		FrameworkVersion: "1.0.0",
		Modules: []ModuleInfo{
			{Name: "http", Import: "github.com/project/mcp-go-core/modules/transport/http"},
		},
	}

	gen := New()
	gen.Generate(context.Background(), cfg)

	data, _ := os.ReadFile(filepath.Join(dir, "modules.go"))
	content := string(data)
	// Imports should be sorted
	for i := 0; i < len(strings.Split(content, "\n"))-1; i++ {
		lines := strings.Split(content, "\n")
		if strings.HasPrefix(lines[i], "\t\"") && i > 0 {
			if strings.HasPrefix(lines[i-1], "\t\"") {
				if lines[i] < lines[i-1] {
					t.Fatal("imports not sorted")
				}
			}
		}
	}
}

func TestGenerateRouterGo(t *testing.T) {
	dir := t.TempDir()
	g := setupTestGraph()
	res, _ := featuregraph.Resolve(featuregraph.Config{
		Graph:          g,
		ExplicitEnable: []string{"http_transport"},
	})

	cfg := Config{
		OutputDir:        dir,
		Resolution:       res,
		FrameworkVersion: "1.0.0",
	}

	gen := New()
	gen.Generate(context.Background(), cfg)

	data, err := os.ReadFile(filepath.Join(dir, "router.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "NewRouter") {
		t.Fatal("expected NewRouter")
	}
}

func TestGenerateMetadata(t *testing.T) {
	dir := t.TempDir()
	g := setupTestGraph()
	res, _ := featuregraph.Resolve(featuregraph.Config{
		Graph:          g,
		ExplicitEnable: []string{"http_transport"},
	})

	cfg := Config{
		OutputDir:        dir,
		Resolution:       res,
		FrameworkVersion: "1.0.0",
	}

	gen := New()
	gen.Generate(context.Background(), cfg)

	metaDir := filepath.Join(dir, "..", "generated")
	data, err := os.ReadFile(filepath.Join(metaDir, "metadata.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "http_transport") {
		t.Fatal("expected http_transport in metadata")
	}
}
