# mcp-go-core

A minimal MCP (Model Context Protocol) server framework for Go.

## Overview

mcp-go-core provides a minimal, dependency-light implementation of the MCP protocol for building MCP servers in Go. It follows a strict architecture where Core types define interfaces and Modules provide concrete implementations.

## Architecture

```
cmd/mcp-go-core/     → CLI tool
core/                → Core types and interfaces (no external deps)
internal/            → Build-time tools (analyzer, generator, builder, featuregraph)
modules/             → Concrete module implementations (transport, security, storage, middleware)
benchmarks/          → Performance benchmarks
tests/               → Integration and verification tests
```

## Usage

### Building an MCP Server

```go
package main

import (
    "context"
    "fmt"
    "os"
    "os/signal"
    "syscall"

    "github.com/project/mcp-go-core/core/server"
    "github.com/project/mcp-go-core/core/tool"
    "github.com/project/mcp-go-core/modules/transport/stdio"
)

func main() {
    srv := server.NewServer(server.WithName("my-server"))

    // Register a tool
    srv.AddTool(tool.NewTool("greet", "Say hello",
        tool.Schema{"type": "object"},
        func(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
            return &protocol.Response{Result: "Hello!"}, nil
        },
    ))

    // Serve over stdio
    transport := stdio.New(nil, nil)
    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop()

    if err := transport.Serve(ctx, handler); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}
```

### CLI

```
mcp-go-core init        Initialize MCP project
mcp-go-core analyze     Analyze application for MCP features
mcp-go-core generate    Generate MCP server code
mcp-go-core build       Build MCP server binary
mcp-go-core test        Test MCP server
mcp-go-core benchmark   Run benchmarks
mcp-go-core doctor      Run diagnostics
mcp-go-core overview    Show project overview
mcp-go-core clean       Clean generated files
```

## License

MIT
