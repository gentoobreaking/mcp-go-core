// Package kubernetes provides Kubernetes integration for MCP deployments.
package kubernetes

import (
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/api/core/v1"
)

// DeploymentSpec represents a Kubernetes deployment manifest.
type DeploymentSpec struct {
	Name          string
	Image         string
	Port          int32
	Profile       string
	Features      []string
	Resources     ResourceRequirements
	Args          []string
	Env           []v1.EnvVar
}

// ResourceRequirements defines compute resource requirements.
type ResourceRequirements struct {
	Requests map[string]string
	Limits   map[string]string
}

// GenerateDeployment generates a Kubernetes deployment YAML.
func GenerateDeployment(spec DeploymentSpec) (string, error) {
	if spec.Name == "" {
		return "", fmt.Errorf("deployment name is required")
	}
	if spec.Image == "" {
		return "", fmt.Errorf("image is required")
	}

	yaml := fmt.Sprintf(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: %s
  labels:
    app: %s
spec:
  replicas: 1
  selector:
    matchLabels:
      app: %s
  template:
    metadata:
      labels:
        app: %s
    spec:
      containers:
      - name: %s
        image: %s
        ports:
        - containerPort: %d
        env:
        - name: MCP_PROFILE
          value: "%s"
`, spec.Name, spec.Name, spec.Name, spec.Name, spec.Name, spec.Image, spec.Port, spec.Profile)

	return yaml, nil
}

// GenerateService generates a Kubernetes service YAML.
func GenerateService(spec DeploymentSpec) (string, error) {
	yaml := fmt.Sprintf(`apiVersion: v1
kind: Service
metadata:
  name: %s
  labels:
    app: %s
spec:
  selector:
    app: %s
  ports:
  - protocol: TCP
    port: %d
    targetPort: %d
`, spec.Name, spec.Name, spec.Name, spec.Port, spec.Port)

	return yaml, nil
}

// WriteManifests writes deployment and service manifests to a directory.
func WriteManifests(dir string, spec DeploymentSpec) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	dep, err := GenerateDeployment(spec)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, spec.Name+"-deployment.yaml"), []byte(dep), 0644); err != nil {
		return err
	}

	svc, err := GenerateService(spec)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, spec.Name+"-service.yaml"), []byte(svc), 0644); err != nil {
		return err
	}

	return nil
}
