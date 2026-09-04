package kubernetes

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	v1 "k8s.io/api/core/v1"
)

func TestGenerateDeployment(t *testing.T) {
	spec := DeploymentSpec{
		Name:     "mcp-server",
		Image:    "myregistry/mcp-server:latest",
		Port:     8080,
		Profile:  "production",
		Features: []string{"http", "jwt"},
	}

	yaml, err := GenerateDeployment(spec)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(yaml, "apiVersion: apps/v1") {
		t.Fatal("expected apiVersion")
	}
	if !strings.Contains(yaml, "kind: Deployment") {
		t.Fatal("expected kind: Deployment")
	}
	if !strings.Contains(yaml, "mcp-server") {
		t.Fatal("expected deployment name")
	}
}

func TestGenerateDeploymentMissingName(t *testing.T) {
	spec := DeploymentSpec{Image: "test"}
	_, err := GenerateDeployment(spec)
	if err == nil {
		t.Fatal("expected error for missing name")
	}
}

func TestGenerateDeploymentMissingImage(t *testing.T) {
	spec := DeploymentSpec{Name: "test"}
	_, err := GenerateDeployment(spec)
	if err == nil {
		t.Fatal("expected error for missing image")
	}
}

func TestGenerateService(t *testing.T) {
	spec := DeploymentSpec{
		Name: "mcp-server",
		Port: 8080,
	}

	yaml, err := GenerateService(spec)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(yaml, "kind: Service") {
		t.Fatal("expected kind: Service")
	}
}

func TestWriteManifests(t *testing.T) {
	dir := t.TempDir()
	spec := DeploymentSpec{
		Name:    "mcp-server",
		Image:   "test:latest",
		Port:    8080,
		Profile: "production",
		Env:     []v1.EnvVar{{Name: "TEST", Value: "1"}},
	}

	if err := WriteManifests(dir, spec); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dir, "mcp-server-deployment.yaml")); err != nil {
		t.Fatal("deployment manifest not written")
	}
	if _, err := os.Stat(filepath.Join(dir, "mcp-server-service.yaml")); err != nil {
		t.Fatal("service manifest not written")
	}
}

func TestResourceRequirements(t *testing.T) {
	spec := DeploymentSpec{
		Name: "test",
		Image: "test:latest",
		Resources: ResourceRequirements{
			Requests: map[string]string{"cpu": "100m", "memory": "128Mi"},
			Limits:   map[string]string{"cpu": "500m", "memory": "256Mi"},
		},
	}
	if spec.Resources.Requests["cpu"] != "100m" {
		t.Fatal("resource request mismatch")
	}
}
