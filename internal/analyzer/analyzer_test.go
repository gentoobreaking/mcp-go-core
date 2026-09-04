package analyzer

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/project/mcp-go-core/internal/featuregraph"
)

func TestAnalyzeWithExplicitConfig(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		ProjectDir:        dir,
		Profile:           "development",
		ExplicitFeatures:  []string{"http_transport"},
	}

	result, err := Analyze(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, f := range result.Features {
		if f.Name == "http_transport" && f.Source == "config" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected http_transport from config, got %+v", result.Features)
	}
}

func TestAnalyzeFromKnownAPI(t *testing.T) {
	dir := t.TempDir()
	appDir := filepath.Join(dir, "app")
	os.MkdirAll(appDir, 0755)
	os.WriteFile(filepath.Join(appDir, "main.go"), []byte(`
package main
func main() {
	jwt.Configure()
}
`), 0644)

	cfg := Config{
		ProjectDir: dir,
		Profile:    "development",
	}

	result, err := Analyze(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}

	foundJwt := false
	foundSecurity := false
	for _, f := range result.Features {
		if f.Name == "jwt_auth" {
			foundJwt = true
			if f.Source != "known_api" {
				t.Fatalf("expected source known_api, got %s", f.Source)
			}
		}
		if f.Name == "security" {
			foundSecurity = true
		}
	}
	if !foundJwt {
		t.Fatal("expected jwt_auth feature from known API")
	}
	if !foundSecurity {
		t.Fatal("expected security feature from known API")
	}
}

func TestAnalyzeNoOAuthInferred(t *testing.T) {
	dir := t.TempDir()
	appDir := filepath.Join(dir, "app")
	os.MkdirAll(appDir, 0755)
	os.WriteFile(filepath.Join(appDir, "main.go"), []byte(`
package main
import "fmt"
func main() { fmt.Println("hello") }
`), 0644)

	cfg := Config{
		ProjectDir: dir,
		Profile:    "development",
	}

	result, err := Analyze(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range result.Features {
		if f.Name == "oauth" || f.Name == "oauth_auth" {
			t.Fatal("oauth should NOT be inferred when not used")
		}
	}
}

func TestAnalyzeDeterminism(t *testing.T) {
	dir := t.TempDir()
	appDir := filepath.Join(dir, "app")
	os.MkdirAll(appDir, 0755)
	os.WriteFile(filepath.Join(appDir, "main.go"), []byte(`
package main
func main() { jwt.Configure() }
`), 0644)

	cfg := Config{ProjectDir: dir, Profile: "development"}

	r1, _ := Analyze(context.Background(), cfg)
	r2, _ := Analyze(context.Background(), cfg)

	if r1.Hash != r2.Hash {
		t.Fatal("analysis is not deterministic")
	}
}

func TestAnalyzePriorityConfigOverAPI(t *testing.T) {
	dir := t.TempDir()
	appDir := filepath.Join(dir, "app")
	os.MkdirAll(appDir, 0755)
	os.WriteFile(filepath.Join(appDir, "main.go"), []byte(`
package main
func main() { jwt.Configure() }
`), 0644)

	cfg := Config{
		ProjectDir:       dir,
		Profile:          "development",
		ExplicitFeatures: []string{"http_transport"},
	}

	result, err := Analyze(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}

	// http_transport should be from config, not from known_api
	for _, f := range result.Features {
		if f.Name == "http_transport" && f.Source != "config" {
			t.Fatalf("expected config source, got %s", f.Source)
		}
	}
}

func TestReadGeneratedMetadata(t *testing.T) {
	dir := t.TempDir()
	metaDir := filepath.Join(dir, ".mcp", "generated")
	os.MkdirAll(metaDir, 0755)
	meta := map[string]any{"features": []string{"http_transport", "jwt_auth"}}
	data, _ := json.MarshalIndent(meta, "", "  ")
	os.WriteFile(filepath.Join(metaDir, "metadata.json"), data, 0644)

	features, err := readGeneratedMetadata(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(features) != 2 {
		t.Fatalf("expected 2 features, got %d", len(features))
	}
}

func TestReadGeneratedMetadataMissing(t *testing.T) {
	dir := t.TempDir()
	_, err := readGeneratedMetadata(dir)
	if err == nil {
		t.Fatal("expected error for missing metadata")
	}
}

func TestWriteInferredFeatures(t *testing.T) {
	dir := t.TempDir()
	result := &AnalysisResult{
		Features: []InferredFeature{
			{Name: "a", Source: "config"},
			{Name: "b", Source: "known_api"},
		},
		Source: "analysis",
		Hash:   "abc123",
	}

	if err := WriteInferredFeatures(dir, result); err != nil {
		t.Fatal(err)
	}

	// Read back and verify
	data, err := os.ReadFile(filepath.Join(dir, ".mcp", "inferred-features.json"))
	if err != nil {
		t.Fatal(err)
	}

	var read Result
	json.Unmarshal(data, &read)
	if len(read.Features) != 2 {
		t.Fatal("expected 2 features")
	}
}

type Result struct {
	Features []InferredFeature `json:"features"`
}

func TestScanGoImports(t *testing.T) {
	dir := t.TempDir()
	appDir := filepath.Join(dir, "app")
	os.MkdirAll(appDir, 0755)
	os.WriteFile(filepath.Join(appDir, "main.go"), []byte(`
package main
import "github.com/project/mcp-go-core/modules/transport/http"
func main() {}
`), 0644)

	features, err := scanGoImports(dir)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, f := range features {
		if f == "http_transport" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected http_transport from import scan")
	}
}

func TestResolveFeatures(t *testing.T) {
	g := featuregraph.NewGraph()
	g.AddFeature(featuregraph.FeatureDescriptor{Name: "http_transport"})
	g.AddFeature(featuregraph.FeatureDescriptor{Name: "oauth"})

	result := &AnalysisResult{
		Features: []InferredFeature{
			{Name: "http_transport"},
		},
	}

	enabled, err := ResolveFeatures(g, result)
	if err != nil {
		t.Fatal(err)
	}
	if len(enabled) != 1 || enabled[0] != "http_transport" {
		t.Fatalf("expected [http_transport], got %v", enabled)
	}
}
