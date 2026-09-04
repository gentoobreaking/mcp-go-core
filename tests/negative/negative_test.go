package negative

import (
	"testing"

	"github.com/project/mcp-go-core/internal/featuregraph"
)

// TestNegativeDisabledHardDependency tests that disabling a HARD dependency of an enabled feature fails.
func TestNegativeDisabledHardDependency(t *testing.T) {
	g := featuregraph.NewGraph()
	g.AddFeature(featuregraph.FeatureDescriptor{
		Name: "app",
		Dependencies: []featuregraph.Dependency{
			{Name: "core", Type: featuregraph.DependencyHard},
		},
	})
	g.AddFeature(featuregraph.FeatureDescriptor{Name: "core"})

	_, err := featuregraph.Resolve(featuregraph.Config{
		Graph:           g,
		ExplicitEnable:  []string{"app"},
		ExplicitDisable: []string{"core"},
	})
	if err == nil {
		t.Fatal("expected FEATURE_REQUIRED error")
	}
}

// TestNegativeUnknownFeature tests that resolving an unknown feature fails.
func TestNegativeUnknownFeature(t *testing.T) {
	g := featuregraph.NewGraph()
	g.AddFeature(featuregraph.FeatureDescriptor{Name: "known"})

	_, err := featuregraph.Resolve(featuregraph.Config{
		Graph:          g,
		ExplicitEnable: []string{"unknown"},
	})
	if err == nil {
		t.Fatal("expected error for unknown feature")
	}
}

// TestNegativeConflictDetection tests that conflicting features are detected.
func TestNegativeConflictDetection(t *testing.T) {
	g := featuregraph.NewGraph()
	g.AddFeature(featuregraph.FeatureDescriptor{Name: "a", Conflicts: []string{"b"}, State: featuregraph.StateEnabled})
	g.AddFeature(featuregraph.FeatureDescriptor{Name: "b", State: featuregraph.StateEnabled})

	err := g.Validate()
	if err == nil {
		t.Fatal("expected FEATURE_CONFLICT")
	}
}
