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
	Name         string
	Version      string
	Description  string
	Module       string
	Dependencies []Dependency
	Conflicts    []string
	Implies      []string
	Default      bool
	Optional     bool
	BuildOnly    bool
	Runtime      bool
	State        FeatureState
}

// ModuleCategory represents the category of a module.
type ModuleCategory string

const (
	ModuleCore          ModuleCategory = "Core"
	ModuleTransport     ModuleCategory = "Transport"
	ModuleSecurity      ModuleCategory = "Security"
	ModuleMiddleware    ModuleCategory = "Middleware"
	ModuleRuntime       ModuleCategory = "Runtime"
	ModuleObservability ModuleCategory = "Observability"
	ModuleStorage       ModuleCategory = "Storage"
	ModuleDeveloper     ModuleCategory = "Developer"
	ModuleIntegration   ModuleCategory = "Integration"
)

// ModuleDescriptor describes a module.
type ModuleDescriptor struct {
	Name        string
	Version     string
	Category    ModuleCategory
	Features    []FeatureDescriptor
	Dependencies []Dependency
	Package     string
	RuntimeInit bool
}

// Validate checks that a FeatureDescriptor has required fields.
func (f FeatureDescriptor) Validate() error {
	if f.Name == "" {
		return fmt.Errorf("feature name cannot be empty")
	}
	return nil
}

// ValidateModule checks that a ModuleDescriptor has required fields.
func (m ModuleDescriptor) Validate() error {
	if m.Name == "" {
		return fmt.Errorf("module name cannot be empty")
	}
	return nil
}

// Error codes for graph validation.
const (
	ErrDuplicateFeature   = "DUPLICATE_FEATURE"
	ErrMissingDependency  = "MISSING_DEPENDENCY"
	ErrMissingFeature     = "MISSING_FEATURE"
	ErrMissingModule      = "MISSING_MODULE"
	ErrFeatureCycle       = "FEATURE_CYCLE"
	ErrFeatureConflict    = "FEATURE_CONFLICT"
	ErrFeatureRequired    = "FEATURE_REQUIRED"
)

// GraphError represents a graph validation error.
type GraphError struct {
	Code    string
	Message string
	Path    []string
}

func (e *GraphError) Error() string {
	if len(e.Path) > 0 {
		return fmt.Sprintf("%s: %s (path: %v)", e.Code, e.Message, e.Path)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// NewGraphError creates a new GraphError.
func NewGraphError(code, message string, path ...string) *GraphError {
	return &GraphError{Code: code, Message: message, Path: path}
}
