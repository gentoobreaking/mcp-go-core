package smoke

import (
	"testing"

	"github.com/project/mcp-go-core/internal/featuregraph"
)

func BenchmarkDispatch(b *testing.B) {
	g := featuregraph.NewGraph()
	g.AddFeature(featuregraph.FeatureDescriptor{
		Name: "http_transport",
		Dependencies: []featuregraph.Dependency{
			{Name: "transport", Type: featuregraph.DependencyHard},
		},
	})
	g.AddFeature(featuregraph.FeatureDescriptor{Name: "transport"})

	res, _ := featuregraph.Resolve(featuregraph.Config{
		Graph:          g,
		ExplicitEnable: []string{"http_transport"},
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = res.Enabled
	}
}
