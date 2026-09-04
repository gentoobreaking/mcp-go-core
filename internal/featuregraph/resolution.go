package featuregraph

import (
	"fmt"
	"sort"
)

// Config holds the configuration for feature resolution.
type Config struct {
	Graph           *Graph
	Profile         string
	ExplicitEnable  []string
	ExplicitDisable []string
}

// Resolution holds the result of feature resolution.
type Resolution struct {
	Graph       *Graph
	Enabled     []string
	Disabled    []string
	Inferred    []string
	Profile     string
	Conflicts   []string
	MissingDeps []string
}

// Resolve performs feature resolution from the config.
func Resolve(cfg Config) (*Resolution, error) {
	g := cfg.Graph
	if g == nil {
		return nil, fmt.Errorf("graph is required")
	}

	if err := g.Validate(); err != nil {
		return nil, err
	}

	enableSet := make(map[string]bool)
	for _, f := range cfg.ExplicitEnable {
		enableSet[f] = true
	}
	disableSet := make(map[string]bool)
	for _, f := range cfg.ExplicitDisable {
		disableSet[f] = true
	}

	for f := range enableSet {
		if disableSet[f] {
			return nil, NewGraphError(ErrFeatureConflict, "feature both enabled and disabled: "+f)
		}
	}

	if err := ValidateExplicitDisable(g, cfg.ExplicitEnable, cfg.ExplicitDisable); err != nil {
		return nil, err
	}

	enabled := make(map[string]bool)
	disabled := make(map[string]bool)
	inferred := make(map[string]bool)

	for _, f := range cfg.ExplicitDisable {
		disabled[f] = true
	}

	for _, name := range cfg.ExplicitEnable {
		if disabled[name] {
			continue
		}
		f := g.GetFeature(name)
		if f == nil {
			return nil, NewGraphError(ErrMissingFeature, "explicitly enabled feature not found: "+name)
		}
		enabled[name] = true
		inferDependencies(g, name, enabled, disabled, inferred)
	}

	for name := range enabled {
		for _, impl := range g.GetImplies(name) {
			if !disabled[impl] {
				inferred[impl] = true
			}
		}
	}

	enabledList := sortedKeys(enabled)
	disabledList := sortedKeys(disabled)
	inferredList := sortedKeys(inferred)

	enabledFinal := make([]string, 0, len(enabledList))
	for _, f := range enabledList {
		if !inferred[f] {
			enabledFinal = append(enabledFinal, f)
		}
	}

	return &Resolution{
		Graph:    g,
		Enabled:  enabledFinal,
		Disabled: disabledList,
		Inferred: inferredList,
		Profile:  cfg.Profile,
	}, nil
}

// ValidateExplicitDisable checks that no explicitly disabled feature is a HARD dependency of an enabled feature.
func ValidateExplicitDisable(g *Graph, explicitlyEnabled, explicitlyDisabled []string) error {
	disabledSet := make(map[string]bool)
	for _, f := range explicitlyDisabled {
		disabledSet[f] = true
	}
	visiting := map[string]bool{}
	for _, name := range explicitlyEnabled {
		if err := checkNoDisabledHardDeps(g, name, disabledSet, visiting); err != nil {
			return err
		}
	}
	return nil
}

func checkNoDisabledHardDeps(g *Graph, name string, disabled, visiting map[string]bool) error {
	if visiting[name] {
		return nil
	}
	visiting[name] = true

	f := g.GetFeature(name)
	if f == nil {
		return nil
	}

	for _, dep := range f.Dependencies {
		if dep.Type != DependencyHard && dep.Type != DependencyImplicit {
			continue
		}
		if disabled[dep.Name] {
			return NewGraphError(ErrFeatureRequired,
				fmt.Sprintf("feature %s requires hard dependency %s which is disabled", name, dep.Name))
		}
		if err := checkNoDisabledHardDeps(g, dep.Name, disabled, visiting); err != nil {
			return err
		}
	}
	return nil
}

// ValidateRequiredDependencies checks the dependency closure invariant.
func ValidateRequiredDependencies(g *Graph, enabled []string) error {
	enabledSet := make(map[string]bool)
	for _, f := range enabled {
		enabledSet[f] = true
	}
	for _, name := range enabled {
		if err := checkRequiredDeps(g, name, enabledSet, map[string]bool{}); err != nil {
			return err
		}
	}
	return nil
}

func checkRequiredDeps(g *Graph, name string, enabled, visiting map[string]bool) error {
	if visiting[name] {
		return nil
	}
	visiting[name] = true

	f := g.GetFeature(name)
	if f == nil {
		return nil
	}

	for _, dep := range f.Dependencies {
		if dep.Type == DependencyHard || dep.Type == DependencyImplicit {
			if !enabled[dep.Name] {
				return NewGraphError(ErrFeatureRequired,
					fmt.Sprintf("feature %s requires dependency %s which is not enabled", name, dep.Name))
			}
			if err := checkRequiredDeps(g, dep.Name, enabled, visiting); err != nil {
				return err
			}
		}
	}
	return nil
}

// inferDependencies recursively adds HARD dependencies.
func inferDependencies(g *Graph, name string, enabled, disabled, inferred map[string]bool) {
	f := g.GetFeature(name)
	if f == nil {
		return
	}
	for _, dep := range f.Dependencies {
		if dep.Type == DependencyHard || dep.Type == DependencyImplicit {
			if disabled[dep.Name] {
				continue
			}
			if !enabled[dep.Name] {
				inferred[dep.Name] = true
				inferDependencies(g, dep.Name, enabled, disabled, inferred)
			}
		}
	}
}

// sortedKeys returns sorted keys from a map.
func sortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// ResolveProfile resolves features for a named profile.
func ResolveProfile(g *Graph, profile string, explicitlyDisabled []string) (*Resolution, error) {
	var enableAll bool
	switch profile {
	case "development", "full":
		enableAll = true
	case "minimal", "production":
		enableAll = false
	default:
		enableAll = false
	}

	var explicitEnable []string
	if enableAll {
		explicitEnable = g.ListFeatures()
	}

	return Resolve(Config{
		Graph:           g,
		Profile:         profile,
		ExplicitEnable:  explicitEnable,
		ExplicitDisable: explicitlyDisabled,
	})
}

// GetDependencyClosure returns all hard dependencies of a feature.
func GetDependencyClosure(g *Graph, name string) []string {
	closure := make(map[string]bool)
	collectHardDeps(g, name, closure, map[string]bool{})
	return sortedKeys(closure)
}

func collectHardDeps(g *Graph, name string, closure, visiting map[string]bool) {
	if visiting[name] {
		return
	}
	visiting[name] = true

	f := g.GetFeature(name)
	if f == nil {
		return
	}

	for _, dep := range f.Dependencies {
		if dep.Type == DependencyHard || dep.Type == DependencyImplicit {
			closure[dep.Name] = true
			collectHardDeps(g, dep.Name, closure, visiting)
		}
	}
}
