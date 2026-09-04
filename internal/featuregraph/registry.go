package featuregraph

import "fmt"

// Registry provides internal-only feature and module registration.
type Registry struct {
	graph *Graph
}

// NewRegistry creates a new Registry.
func NewRegistry() *Registry {
	return &Registry{graph: NewGraph()}
}

// Register adds a feature to the registry.
func (r *Registry) Register(f FeatureDescriptor) error {
	return r.graph.AddFeature(f)
}

// RegisterModule adds a module with its features to the registry.
func (r *Registry) RegisterModule(m ModuleDescriptor) {
	r.graph.AddModule(m)
}

// Get retrieves a feature by name.
func (r *Registry) Get(name string) *FeatureDescriptor {
	return r.graph.GetFeature(name)
}

// GetModule retrieves a module by name.
func (r *Registry) GetModule(name string) *ModuleDescriptor {
	return r.graph.GetModule(name)
}

// List returns all registered feature names.
func (r *Registry) List() []string {
	return r.graph.ListFeatures()
}

// ListModules returns all registered module names.
func (r *Registry) ListModules() []string {
	return r.graph.ListModules()
}

// Validate validates the registry graph.
func (r *Registry) Validate() error {
	return r.graph.Validate()
}

// Graph returns the underlying graph.
func (r *Registry) Graph() *Graph {
	return r.graph
}

// Count returns the number of registered features.
func (r *Registry) Count() int {
	return r.graph.CountFeatures()
}

// MustRegister panics if registration fails.
func (r *Registry) MustRegister(f FeatureDescriptor) {
	if err := r.Register(f); err != nil {
		panic(fmt.Sprintf("failed to register feature %s: %v", f.Name, err))
	}
}
