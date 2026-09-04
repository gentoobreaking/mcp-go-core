package featuregraph

import (
	"fmt"
	"sort"
)

// Graph represents the feature dependency graph.
type Graph struct {
	features map[string]*FeatureDescriptor
	modules  map[string]*ModuleDescriptor
}

// NewGraph creates a new Graph.
func NewGraph() *Graph {
	return &Graph{
		features: make(map[string]*FeatureDescriptor),
		modules:  make(map[string]*ModuleDescriptor),
	}
}

// AddFeature adds a feature to the graph.
func (g *Graph) AddFeature(f FeatureDescriptor) error {
	if f.Name == "" {
		return fmt.Errorf("feature name cannot be empty")
	}
	if _, ok := g.features[f.Name]; ok {
		return NewGraphError(ErrDuplicateFeature, "feature already exists: "+f.Name)
	}
	f.State = f.State
	if f.State == "" {
		f.State = StateAuto
	}
	g.features[f.Name] = &f
	return nil
}

// GetFeature retrieves a feature by name.
func (g *Graph) GetFeature(name string) *FeatureDescriptor {
	return g.features[name]
}

// AddModule adds a module to the graph.
func (g *Graph) AddModule(m ModuleDescriptor) {
	g.modules[m.Name] = &m
	for i := range m.Features {
		f := &m.Features[i]
		if _, ok := g.features[f.Name]; ok {
			continue
		}
		if f.State == "" {
			f.State = StateAuto
		}
		g.features[f.Name] = f
	}
}

// GetModule retrieves a module by name.
func (g *Graph) GetModule(name string) *ModuleDescriptor {
	return g.modules[name]
}

// ListFeatures returns all feature names sorted.
func (g *Graph) ListFeatures() []string {
	names := make([]string, 0, len(g.features))
	for name := range g.features {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// ListModules returns all module names sorted.
func (g *Graph) ListModules() []string {
	names := make([]string, 0, len(g.modules))
	for name := range g.modules {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// GetDependencies returns HARD and IMPLICIT dependencies of a feature.
func (g *Graph) GetDependencies(name string) []string {
	f := g.features[name]
	if f == nil {
		return nil
	}
	var deps []string
	for _, dep := range f.Dependencies {
		if dep.Type == DependencyHard || dep.Type == DependencyImplicit {
			deps = append(deps, dep.Name)
		}
	}
	sort.Strings(deps)
	return deps
}

// GetSoftDependencies returns OPTIONAL dependencies.
func (g *Graph) GetSoftDependencies(name string) []string {
	f := g.features[name]
	if f == nil {
		return nil
	}
	var deps []string
	for _, dep := range f.Dependencies {
		if dep.Type == DependencyOptional {
			deps = append(deps, dep.Name)
		}
	}
	sort.Strings(deps)
	return deps
}

// AllDependencies returns all dependencies.
func (g *Graph) AllDependencies(name string) []string {
	f := g.features[name]
	if f == nil {
		return nil
	}
	var deps []string
	for _, dep := range f.Dependencies {
		deps = append(deps, dep.Name)
	}
	sort.Strings(deps)
	return deps
}

// GetConflicts returns features that this feature conflicts with.
func (g *Graph) GetConflicts(name string) []string {
	f := g.features[name]
	if f == nil {
		return nil
	}
	result := make([]string, len(f.Conflicts))
	copy(result, f.Conflicts)
	sort.Strings(result)
	return result
}

// GetImplies returns features that this feature implies.
func (g *Graph) GetImplies(name string) []string {
	f := g.features[name]
	if f == nil {
		return nil
	}
	implied := make([]string, len(f.Implies))
	copy(implied, f.Implies)
	sort.Strings(implied)
	return implied
}

// Validate checks the graph for correctness.
func (g *Graph) Validate() error {
	for name, f := range g.features {
		for _, dep := range f.Dependencies {
			if g.features[dep.Name] == nil {
				return NewGraphError(ErrMissingDependency,
					fmt.Sprintf("feature %s depends on missing feature %s", name, dep.Name))
			}
		}
		for _, conflict := range f.Conflicts {
			if g.features[conflict] == nil {
				return NewGraphError(ErrMissingFeature,
					fmt.Sprintf("feature %s conflicts with missing feature %s", name, conflict))
			}
		}
		for _, impl := range f.Implies {
			if g.features[impl] == nil {
				return NewGraphError(ErrMissingFeature,
					fmt.Sprintf("feature %s implies missing feature %s", name, impl))
			}
		}
	}

	if cycle := g.detectCycles(); cycle != nil {
		return NewGraphError(ErrFeatureCycle, "cycle detected", cycle...)
	}
	if conflict := g.detectConflicts(); conflict != nil {
		return NewGraphError(ErrFeatureConflict, "conflicting features", conflict...)
	}
	return nil
}

// detectCycles uses DFS to find cycles.
func (g *Graph) detectCycles() []string {
	visiting := make(map[string]bool)
	visited := make(map[string]bool)

	var dfs func(node string, path []string) []string
	dfs = func(node string, path []string) []string {
		if visiting[node] {
			idx := -1
			for i, n := range path {
				if n == node {
					idx = i
					break
				}
			}
			if idx >= 0 {
				cycle := append([]string{}, path[idx:]...)
				cycle = append(cycle, node)
				return cycle
			}
			return append(path, node)
		}
		if visited[node] {
			return nil
		}

		visiting[node] = true
		path = append(path, node)

		for _, dep := range g.AllDependencies(node) {
			if c := dfs(dep, path); c != nil {
				return c
			}
		}

		visiting[node] = false
		visited[node] = true
		return nil
	}

	for name := range g.features {
		if !visited[name] {
			if c := dfs(name, nil); c != nil {
				return c
			}
		}
	}
	return nil
}

// detectConflicts checks for conflicting enabled features.
func (g *Graph) detectConflicts() []string {
	for name, f := range g.features {
		if f.State != StateEnabled && f.State != StateRequired {
			continue
		}
		for _, conflict := range f.Conflicts {
			cf := g.features[conflict]
			if cf != nil && (cf.State == StateEnabled || cf.State == StateRequired) {
				return []string{name, conflict}
			}
		}
	}
	return nil
}

// CountFeatures returns the number of features.
func (g *Graph) CountFeatures() int {
	return len(g.features)
}

// CountModules returns the number of modules.
func (g *Graph) CountModules() int {
	return len(g.modules)
}
