package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/project/mcp-go-core/integration-kubernetes"
	"github.com/spf13/cobra"
)

var k8sCmd = &cobra.Command{
	Use:   "k8s",
	Short: "Generate Kubernetes deployment manifests",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		image, _ := cmd.Flags().GetString("image")
		port, _ := cmd.Flags().GetInt("port")
		output, _ := cmd.Flags().GetString("output")

		if name == "" {
			name = "mcp-server"
		}
		if image == "" {
			image = "localhost/mcp-server:latest"
		}
		if port == 0 {
			port = 8080
		}
		if output == "" {
			output = "."
		}

		spec := kubernetes.DeploymentSpec{
			Name:    name,
			Image:   image,
			Port:    int32(port),
			Profile: "production",
		}

		if err := os.MkdirAll(filepath.Clean(output), 0755); err != nil {
			return fmt.Errorf("create output dir: %w", err)
		}

		if err := kubernetes.WriteManifests(output, spec); err != nil {
			return fmt.Errorf("write manifests: %w", err)
		}

		fmt.Printf("Generated Kubernetes manifests in %s/\n", output)
		fmt.Printf("  - %s-deployment.yaml\n", name)
		fmt.Printf("  - %s-service.yaml\n", name)
		return nil
	},
}

func init() {
	k8sCmd.Flags().StringP("name", "n", "", "deployment name")
	k8sCmd.Flags().StringP("image", "i", "", "container image")
	k8sCmd.Flags().IntP("port", "p", 8080, "container port")
	k8sCmd.Flags().StringP("output", "o", ".", "output directory")
}
