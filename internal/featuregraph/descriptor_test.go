package featuregraph

import "testing"

func TestFeatureDescriptor(t *testing.T) {
	f := FeatureDescriptor{
		Name:        "sse_transport",
		Version:     "1.0.0",
		Description: "SSE transport module",
		Module:      "transport/sse",
		Default:     true,
	}
	if err := Validate(f); err != nil {
		t.Fatal(err)
	}
}

func TestValidateFeatureEmptyName(t *testing.T) {
	f := FeatureDescriptor{Name: ""}
	if err := Validate(f); err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestModuleDescriptor(t *testing.T) {
	m := ModuleDescriptor{
		Name:        "transport/sse",
		Version:     "1.0.0",
		Category:    "transport",
		Package:     "github.com/project/mcp-go-core/modules/transport/sse",
		RuntimeInit: "Init",
	}
	if err := ValidateModule(m); err != nil {
		t.Fatal(err)
	}
}

func TestValidateModuleEmptyName(t *testing.T) {
	m := ModuleDescriptor{Name: ""}
	if err := ValidateModule(m); err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestDependencyTypes(t *testing.T) {
	if DependencyHard != "HARD" {
		t.Fatal("hard dep type mismatch")
	}
	if DependencyOptional != "OPTIONAL" {
		t.Fatal("optional dep type mismatch")
	}
	if DependencyImplicit != "IMPLICIT" {
		t.Fatal("implicit dep type mismatch")
	}
}

func TestFeatureStates(t *testing.T) {
	if StateAuto != "AUTO" || StateEnabled != "ENABLED" || StateDisabled != "DISABLED" {
		t.Fatal("feature state mismatch")
	}
}
