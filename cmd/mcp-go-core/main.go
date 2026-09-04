package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:   "mcp-go-core",
		Short: "Minimal MCP (Model Context Protocol) server core",
	}

	root.AddCommand(
		&cobra.Command{Use: "init", Short: "Initialize MCP project"},
		&cobra.Command{Use: "analyze", Short: "Analyze application for MCP features"},
		&cobra.Command{Use: "generate", Short: "Generate MCP server code"},
		&cobra.Command{Use: "build", Short: "Build MCP server binary"},
		&cobra.Command{Use: "test", Short: "Test MCP server"},
		&cobra.Command{Use: "benchmark", Short: "Run benchmarks"},
		&cobra.Command{Use: "doctor", Short: "Run diagnostics"},
		&cobra.Command{Use: "overview", Short: "Show project overview"},
		&cobra.Command{Use: "clean", Short: "Clean generated files"},
	)

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
