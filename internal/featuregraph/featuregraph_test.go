package featuregraph

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestGraphAddFeature(t *testing.T) {
	g := NewGraph()
	f := FeatureDescriptor{
		Name:        "feature_a",
		Description: "test feature a",
		Dependencies: []Dependency{
			{Name: "feature_b", Type: DependencyHard},
		},
	}
	if err := g.AddFeature(f); err != nil {
		t.Fatal(err)
	}
	if g.GetFeature("feature_a") == nil {
		t.Fatal("feature not found")
	}
}

func TestDuplicateFeature(t *testing.T) {
	g := NewGraph()
	g.AddFeature(FeatureDescriptor{Name: "dup"})
	err := g.AddFeature(FeatureDescriptor{Name: "dup"})
	if err == nil {
		t.Fatal("expected duplicate feature error")
	}
	ge, ok := err.(*GraphError)
	if !ok {
		t.Fatalf("expected GraphError, got %T", err)
	}
	if ge.Code != ErrDuplicateFeature {
		t.Fatal("wrong error code")
	}
}

func TestMissingDependency(t *testing.T) {
	g := NewGraph()
	g.AddFeature(FeatureDescriptor{Name: "a", Dependencies: []Dependency{{Name: "missing", Type: DependencyHard}}})
	err := g.Validate()
	if err == nil {
		t.Fatal("expected missing dependency error")
	}
	ge, ok := err.(*GraphError)
	if !ok {
		t.Fatalf("expected GraphError, got %T", err)
	}
	if ge.Code != ErrMissingDependency {
		t.Fatal("wrong error code")
	}
}

func TestCycleDetection(t *testing.T) {
	g := NewGraph()
	g.AddFeature(FeatureDescriptor{Name: "a", Dependencies: []Dependency{{Name: "b", Type: DependencyHard}}})
	g.AddFeature(FeatureDescriptor{Name: "b", Dependencies: []Dependency{{Name: "a", Type: DependencyHard}}})
	err := g.Validate()
	if err == nil {
		t.Fatal("expected cycle error")
	}
	ge, ok := err.(*GraphError)
	if !ok {
		t.Fatalf("expected GraphError, got %T", err)
	}
	if ge.Code != ErrFeatureCycle {
		t.Fatalf("expected FEATURE_CYCLE, got %s", ge.Code)
	}
}

func TestSelfDependency(t *testing.T) {
	g := NewGraph()
	g.AddFeature(FeatureDescriptor{Name: "a", Dependencies: []Dependency{{Name: "a", Type: DependencyHard}}})
	err := g.Validate()
	if err == nil {
		t.Fatal("expected cycle error for self-dependency")
	}
	ge, ok := err.(*GraphError)
	if !ok {
		t.Fatalf("expected GraphError, got %T", err)
	}
	if ge.Code != ErrFeatureCycle {
		t.Fatalf("expected FEATURE_CYCLE, got %s", ge.Code)
	}
}

func TestConflictDetection(t *testing.T) {
	g := NewGraph()
	g.AddFeature(FeatureDescriptor{Name: "a", Conflicts: []string{"b"}, State: StateEnabled})
	g.AddFeature(FeatureDescriptor{Name: "b", State: StateEnabled})
	err := g.Validate()
	if err == nil {
		t.Fatal("expected conflict error")
	}
	ge, ok := err.(*GraphError)
	if !ok {
		t.Fatalf("expected GraphError, got %T", err)
	}
	if ge.Code != ErrFeatureConflict {
		t.Fatalf("expected FEATURE_CONFLICT, got %s", ge.Code)
	}
}

func TestBasicResolution(t *testing.T) {
	g := NewGraph()
	g.AddFeature(FeatureDescriptor{Name: "a", Dependencies: []Dependency{{Name: "b", Type: DependencyHard}}})
	g.AddFeature(FeatureDescriptor{Name: "b"})

	res, err := Resolve(Config{
		Graph:          g,
		ExplicitEnable: []string{"a"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Enabled) != 1 || res.Enabled[0] != "a" {
		t.Fatalf("expected only 'a' in enabled, got %v", res.Enabled)
	}
	if len(res.Inferred) != 1 || res.Inferred[0] != "b" {
		t.Fatalf("expected 'b' in inferred, got %v", res.Inferred)
	}
}

func TestTransitiveDependency(t *testing.T) {
	g := NewGraph()
	g.AddFeature(FeatureDescriptor{Name: "a", Dependencies: []Dependency{{Name: "b", Type: DependencyHard}}})
	g.AddFeature(FeatureDescriptor{Name: "b", Dependencies: []Dependency{{Name: "c", Type: DependencyHard}}})
	g.AddFeature(FeatureDescriptor{Name: "c"})

	res, err := Resolve(Config{
		Graph:          g,
		ExplicitEnable: []string{"a"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Inferred) != 2 {
		t.Fatalf("expected 2 inferred (b, c), got %v", res.Inferred)
	}
}

func TestExplicitDisableHardDep(t *testing.T) {
	g := NewGraph()
	g.AddFeature(FeatureDescriptor{Name: "a", Dependencies: []Dependency{{Name: "b", Type: DependencyHard}}})
	g.AddFeature(FeatureDescriptor{Name: "b"})

	_, err := Resolve(Config{
		Graph:           g,
		ExplicitEnable:  []string{"a"},
		ExplicitDisable: []string{"b"},
	})
	if err == nil {
		t.Fatal("expected FEATURE_REQUIRED error")
	}
	ge, ok := err.(*GraphError)
	if !ok {
		t.Fatalf("expected GraphError, got %T", err)
	}
	if ge.Code != ErrFeatureRequired {
		t.Fatalf("expected FEATURE_REQUIRED, got %s", ge.Code)
	}
}

func TestOptionalDisableNoError(t *testing.T) {
	g := NewGraph()
	g.AddFeature(FeatureDescriptor{Name: "a", Dependencies: []Dependency{{Name: "b", Type: DependencyOptional}}})
	g.AddFeature(FeatureDescriptor{Name: "b"})

	res, err := Resolve(Config{
		Graph:           g,
		ExplicitEnable:  []string{"a"},
		ExplicitDisable: []string{"b"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Enabled) != 1 {
		t.Fatalf("expected 1 enabled, got %v", res.Enabled)
	}
}

func TestDeterministicResolution(t *testing.T) {
	g := NewGraph()
	g.AddFeature(FeatureDescriptor{Name: "a", Dependencies: []Dependency{{Name: "c", Type: DependencyHard}, {Name: "b", Type: DependencyHard}}})
	g.AddFeature(FeatureDescriptor{Name: "b"})
	g.AddFeature(FeatureDescriptor{Name: "c"})

	// Compare JSON byte output across 3 runs
	hashes := make([][]byte, 3)
	for i := 0; i < 3; i++ {
		res, err := Resolve(Config{
			Graph:          g,
			ExplicitEnable: []string{"a"},
		})
		if err != nil {
			t.Fatal(err)
		}
		data, _ := json.Marshal(res.Enabled)
		hashes[i] = data
	}

	for i := 1; i < 3; i++ {
		if !bytes.Equal(hashes[0], hashes[i]) {
			t.Fatal("resolution is not deterministic")
		}
	}
}

func TestProfileResolution(t *testing.T) {
	g := NewGraph()
	g.AddFeature(FeatureDescriptor{Name: "core_a"})
	g.AddFeature(FeatureDescriptor{Name: "core_b"})

	res, err := ResolveProfile(g, "development", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Enabled) != 2 {
		t.Fatalf("expected 2 enabled for development, got %d", len(res.Enabled))
	}
}

func TestDependencyClosure(t *testing.T) {
	g := NewGraph()
	g.AddFeature(FeatureDescriptor{
		Name: "a",
		Dependencies: []Dependency{
			{Name: "b", Type: DependencyHard},
			{Name: "c", Type: DependencyHard},
		},
	})
	g.AddFeature(FeatureDescriptor{
		Name: "b",
		Dependencies: []Dependency{
			{Name: "c", Type: DependencyHard},
		},
	})
	g.AddFeature(FeatureDescriptor{Name: "c"})

	closure := GetDependencyClosure(g, "a")
	if len(closure) != 2 {
		t.Fatalf("expected 2 deps in closure (b, c), got %v", closure)
	}
}

func TestGraphCount(t *testing.T) {
	g := NewGraph()
	g.AddFeature(FeatureDescriptor{Name: "a"})
	g.AddFeature(FeatureDescriptor{Name: "b"})
	if g.CountFeatures() != 2 {
		t.Fatal("wrong count")
	}
}

func TestListFeaturesSorted(t *testing.T) {
	g := NewGraph()
	g.AddFeature(FeatureDescriptor{Name: "z"})
	g.AddFeature(FeatureDescriptor{Name: "a"})
	features := g.ListFeatures()
	if features[0] != "a" || features[1] != "z" {
		t.Fatalf("expected sorted [a, z], got %v", features)
	}
}

func TestLockFileDeterministic(t *testing.T) {
	g := NewGraph()
	g.AddFeature(FeatureDescriptor{Name: "a", Dependencies: []Dependency{{Name: "b", Type: DependencyHard}}})
	g.AddFeature(FeatureDescriptor{Name: "b"})

	res, _ := Resolve(Config{
		Graph:          g,
		ExplicitEnable: []string{"a"},
	})

	lock1 := GenerateLock(res, "1.0.0")
	lock2 := GenerateLock(res, "1.0.0")

	if lock1.GraphHash != lock2.GraphHash {
		t.Fatal("graph_hash not deterministic")
	}

	lock3 := GenerateLock(res, "2.0.0")
	if lock1.GraphHash == lock3.GraphHash {
		t.Fatal("graph_hash should change with framework version")
	}
}

func TestComputeLockHash(t *testing.T) {
	h1 := ComputeLockHash([]string{"a", "b"}, [][]string{{"a", "b"}}, "dev", "1.0.0")
	h2 := ComputeLockHash([]string{"b", "a"}, [][]string{{"a", "b"}}, "dev", "1.0.0")
	if h1 != h2 {
		t.Fatal("hash should be deterministic regardless of input order")
	}
}

func TestLockFileWrite(t *testing.T) {
	g := NewGraph()
	g.AddFeature(FeatureDescriptor{Name: "a", Dependencies: []Dependency{{Name: "b", Type: DependencyHard}}})
	g.AddFeature(FeatureDescriptor{Name: "b"})

	res, _ := Resolve(Config{
		Graph:          g,
		ExplicitEnable: []string{"a"},
	})

	lock := GenerateLock(res, "1.0.0")

	dir := t.TempDir()
	if err := WriteLock(dir, lock); err != nil {
		t.Fatal(err)
	}

	// Read back and verify
	data, err := os.ReadFile(filepath.Join(dir, ".mcp", "features.lock"))
	if err != nil {
		t.Fatal(err)
	}
	var readLock LockFile
	json.Unmarshal(data, &readLock)
	if readLock.GraphHash != lock.GraphHash {
		t.Fatal("graph_hash mismatch after write")
	}
}

func TestValidateRequiredDependencies(t *testing.T) {
	g := NewGraph()
	g.AddFeature(FeatureDescriptor{
		Name: "a",
		Dependencies: []Dependency{{Name: "b", Type: DependencyHard}},
	})
	g.AddFeature(FeatureDescriptor{Name: "b"})

	err := ValidateRequiredDependencies(g, []string{"a", "b"})
	if err != nil {
		t.Fatal(err)
	}

	err = ValidateRequiredDependencies(g, []string{"a"})
	if err == nil {
		t.Fatal("expected error: b not enabled")
	}
}
