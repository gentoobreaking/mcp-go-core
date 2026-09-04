package ci

import (
	"testing"

	"github.com/project/mcp-go-core/internal/featuregraph"
)

// TestProfileVerification tests all build profiles.
func TestProfileVerification(t *testing.T) {
	profiles := []struct {
		name     string
		enable   []string
		disable  []string
	}{
		{"minimal", []string{}, []string{"http_transport", "jwt_auth", "oauth"}},
		{"development", nil, nil},
		{"production", nil, nil},
		{"secure", []string{}, []string{"oauth"}},
		{"observable", []string{}, []string{}},
		{"full", nil, nil},
	}

	g := featuregraph.NewGraph()
	g.AddFeature(featuregraph.FeatureDescriptor{Name: "http_transport"})
	g.AddFeature(featuregraph.FeatureDescriptor{Name: "jwt_auth"})
	g.AddFeature(featuregraph.FeatureDescriptor{Name: "oauth"})

	for _, p := range profiles {
		var enable []string
		if p.enable != nil {
			enable = p.enable
		} else {
			enable = g.ListFeatures()
		}

		res, err := featuregraph.Resolve(featuregraph.Config{
			Graph:           g,
			Profile:         p.name,
			ExplicitEnable:  enable,
			ExplicitDisable: p.disable,
		})
		if err != nil {
			t.Errorf("profile %s failed: %v", p.name, err)
		}
		_ = res
	}
}

// TestMinimalProfileExcludesOauth tests that minimal profile excludes oauth.
func TestMinimalProfileExcludesOauth(t *testing.T) {
	g := featuregraph.NewGraph()
	g.AddFeature(featuregraph.FeatureDescriptor{Name: "http_transport"})
	g.AddFeature(featuregraph.FeatureDescriptor{Name: "oauth"})

	res, _ := featuregraph.ResolveProfile(g, "minimal", nil)
	for _, f := range res.Enabled {
		if f == "oauth" {
			t.Fatal("oauth should not be enabled in minimal profile")
		}
	}
}
