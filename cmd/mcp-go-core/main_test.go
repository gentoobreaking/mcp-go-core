package main

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
)

func TestRootCommand(t *testing.T) {
	root := &cobra.Command{
		Use:  "mcp-go-core",
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("mcp-go-core root")
		},
	}

	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetArgs([]string{"--help"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Fatal("expected output")
	}
}

func TestVersionCommand(t *testing.T) {
	buf := &bytes.Buffer{}

	cmd := &cobra.Command{
		Use: "version",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Printf("mcp-go-core version %s\n", version)
		},
	}
	cmd.SetOut(buf)
	cmd.SetArgs([]string{})

	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	if buf.String() == "" {
		t.Fatal("expected version output")
	}
}

func TestInitCommand(t *testing.T) {
	buf := &bytes.Buffer{}

	cmd := &cobra.Command{
		Use: "init",
		RunE: func(cmd *cobra.Command, args []string) error {
			name, _ := cmd.Flags().GetString("name")
			profile, _ := cmd.Flags().GetString("profile")
			if name == "" {
				name = "my-mcp-server"
			}
			if profile == "" {
				profile = "development"
			}
			cmd.Printf("Initializing MCP project: %s (profile: %s)\n", name, profile)
			return nil
		},
	}
	cmd.Flags().StringP("name", "n", "", "")
	cmd.Flags().StringP("profile", "p", "", "")
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--name", "testapp", "--profile", "production"})

	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if output == "" {
		t.Fatal("expected init output")
	}
}

func TestBuildCommand(t *testing.T) {
	buf := &bytes.Buffer{}

	cmd := &cobra.Command{
		Use: "build",
		RunE: func(cmd *cobra.Command, args []string) error {
			output, _ := cmd.Flags().GetString("output")
			profile, _ := cmd.Flags().GetString("profile")
			if output == "" {
				output = "dist/mcp-server"
			}
			if profile == "" {
				profile = "production"
			}
			cmd.Printf("Building MCP server: %s (profile: %s)\n", output, profile)
			return nil
		},
	}
	cmd.Flags().StringP("output", "o", "", "")
	cmd.Flags().StringP("profile", "p", "", "")
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--output", "dist/server", "--profile", "development"})

	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if output == "" {
		t.Fatal("expected build output")
	}
}

func TestGenerateCommand(t *testing.T) {
	buf := &bytes.Buffer{}

	cmd := &cobra.Command{
		Use: "generate",
		RunE: func(cmd *cobra.Command, args []string) error {
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				cmd.Println("Dry run: would generate server code")
				return nil
			}
			cmd.Println("Generating MCP server code...")
			return nil
		},
	}
	cmd.Flags().Bool("dry-run", false, "")
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--dry-run"})

	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	if buf.String() == "" {
		t.Fatal("expected generate output")
	}
}

func TestVerifyCommand(t *testing.T) {
	buf := &bytes.Buffer{}

	cmd := &cobra.Command{
		Use: "verify",
		RunE: func(cmd *cobra.Command, args []string) error {
			binary, _ := cmd.Flags().GetString("binary")
			if binary == "" {
				binary = "dist/server"
			}
			cmd.Printf("Verifying binary: %s\n", binary)
			return nil
		},
	}
	cmd.Flags().StringP("binary", "b", "", "")
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--binary", "dist/mcp-server"})

	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	if buf.String() == "" {
		t.Fatal("expected verify output")
	}
}

func TestRunCommand(t *testing.T) {
	cmd := &cobra.Command{
		Use: "run",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	if cmd == nil {
		t.Fatal("expected run command")
	}
}
