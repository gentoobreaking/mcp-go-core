package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/project/mcp-go-core/core/middleware"
	"github.com/project/mcp-go-core/core/server"
	"github.com/project/mcp-go-core/modules/transport/stdio"
)

var version = "0.1.0"

func main() {
	root := &cobra.Command{
		Use:   "mcp-go-core",
		Short: "Minimal MCP (Model Context Protocol) server core",
		Long:  "mcp-go-core is a minimal MCP server with static composition and module isolation.",
	}

	root.AddCommand(initCmd)
	root.AddCommand(buildCmd)
	root.AddCommand(runCmd)
	root.AddCommand(generateCmd)
	root.AddCommand(verifyCmd)

	// version command
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print version info",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("mcp-go-core version %s\n", version)
		},
	}
	root.AddCommand(versionCmd)

	// root --version flag
	root.PersistentFlags().BoolP("version", "V", false, "print version")
	root.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if v, _ := root.PersistentFlags().GetBool("version"); v {
			fmt.Printf("mcp-go-core version %s\n", version)
			os.Exit(0)
		}
	}

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize MCP project",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		profile, _ := cmd.Flags().GetString("profile")
		if name == "" {
			name = "my-mcp-server"
		}
		if profile == "" {
			profile = "development"
		}
		fmt.Printf("Initializing MCP project: %s (profile: %s)\n", name, profile)
		return nil
	},
}

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build MCP server binary",
	RunE: func(cmd *cobra.Command, args []string) error {
		output, _ := cmd.Flags().GetString("output")
		profile, _ := cmd.Flags().GetString("profile")
		if output == "" {
			output = "dist/mcp-server"
		}
		if profile == "" {
			profile = "production"
		}
		fmt.Printf("Building MCP server: %s (profile: %s)\n", output, profile)
		return nil
	},
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run MCP server",
	RunE: func(cmd *cobra.Command, args []string) error {
		addr, _ := cmd.Flags().GetString("addr")
		transport, _ := cmd.Flags().GetString("transport")
		if addr == "" {
			addr = "localhost:8080"
		}
		if transport == "" {
			transport = "stdio"
		}

		fmt.Printf("Running MCP server (transport: %s, addr: %s)\n", transport, addr)

		// Use builder pattern to construct server
		var tr server.Transport
		switch transport {
		case "stdio":
			tr = stdio.New(nil, nil)
		default:
			tr = stdio.New(nil, nil)
		}

		srv := server.NewBuilder().
			WithName("mcp-server").
			WithTimeout(30 * time.Second).
			WithTransport(tr).
			WithMiddleware(middleware.Recovery()).
			MustBuild()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		return srv.Run(ctx)
	},
}

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate MCP server code",
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			fmt.Println("Dry run: would generate server code")
			return nil
		}
		fmt.Println("Generating MCP server code...")
		return nil
	},
}

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify MCP server binary",
	RunE: func(cmd *cobra.Command, args []string) error {
		binary, _ := cmd.Flags().GetString("binary")
		if binary == "" {
			binary = "dist/server"
		}
		fmt.Printf("Verifying binary: %s\n", binary)
		return nil
	},
}

func init() {
	initCmd.Flags().StringP("name", "n", "", "project name")
	initCmd.Flags().StringP("profile", "p", "", "profile (development/production)")

	buildCmd.Flags().StringP("output", "o", "", "output binary path")
	buildCmd.Flags().StringP("profile", "p", "", "profile (development/production)")

	runCmd.Flags().StringP("addr", "a", "", "address")
	runCmd.Flags().StringP("transport", "t", "", "transport (stdio/http/sse)")

	generateCmd.Flags().Bool("dry-run", false, "show what would be generated")

	verifyCmd.Flags().StringP("binary", "b", "", "binary path")
}
