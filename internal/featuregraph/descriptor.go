// Package featuregraph provides feature graph types for MCP module dependency management.
package featuregraph

import "fmt"

// DependencyType represents the strength of a dependency.
type DependencyType string

const (
	DependencyHard     DependencyType = "HARD"
	DependencyOptional DependencyType = "OPTIONAL"
	DependencyImplicit DependencyType = "IMPLICIT"
)

// FeatureState represents the state of a feature.
type FeatureState string

const (
	StateAuto      FeatureState = "AUTO"
	StateEnabled   FeatureState = "ENABLED"
	StateDisabled  FeatureState = "DISABLED"
	StateRequired  FeatureState = "REQUIRED"
	StateInferred  FeatureState = "INFERRED"
)

// Dependency represents a dependency on another feature.
type Dependency struct {
	Name string
	Type DependencyType
}

// FeatureDescriptor describes a single feature.
type FeatureDescriptor struct {
	Name        string
	Version     string
	Description string

	Module      string
	Dependencies []Dependency
	Conflicts    []string
	Implies      []string
	Default      bool
	Optional     bool
	BuildOnly    bool
	Runtime      bool
	State        FeatureState
}

// ModuleDescriptor describes a module.
type ModuleDescriptor struct {
	Name        string
	Version     string
	Category    string
	Features    []FeatureDescriptor
	Dependencies []Dependency
	Package     string
	RuntimeInit string
}

// Validate checks that a FeatureDescriptor has required fields.
func Validate(f FeatureDescriptor) error {
	if f.Name == "" {
		return fmt.Errorf("feature name cannot be empty")
	}
	return nil
}

// ValidateModule checks that a ModuleDescriptor has required fields.
func ValidateModule(m ModuleDescriptor) error {
	if m.Name == "" {
		return fmt.Errorf("module name cannot be empty")
	}
	return nil
}
