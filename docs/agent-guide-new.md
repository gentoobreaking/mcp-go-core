# AI Agent Guide: Building a New MCP with mcp-go-core

This guide shows an AI agent how to create a **new** MCP server from scratch using the mcp-go-core framework.

## Prerequisites

- Go 1.26+
- Basic understanding of MCP (Model Context Protocol) concepts

## Step 1: Create Project Structure

```
my-mcp-server/
├── cmd/my-mcp/
│   └── main.go          # Server entry point
├── tools/
│   └── greet.go         # Tool definitions
├── go.mod
└── go.sum
```

## Step 2: Initialize Go Module

```bash
go mod init my-mcp-server
go mod edit -require=github.com/project/mcp-go-core@v0.1.0
go mod tidy
```

## Step 3: Define a Tool

Create a file like `tools/greet.go`:

```go
package tools

import (
    "context"
    "github.com/project/mcp-go-core/core/tool"
    "github.com/project/mcp-go-core/core/protocol"
)

func GreetTool() tool.Tool {
    return tool.NewTool(
        "greet",
        "Returns a greeting message",
        tool.Schema{"type": "object"},
        func(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
            return &protocol.Response{
                JSONRPC: "2.0",
                ID:      req.ID,
                Result:  map[string]any{"message": "Hello from mcp-go-core!"},
            }, nil
        },
    )
}
```

## Step 4: Create Server Entry Point

Create `cmd/my-mcp/main.go`:

```go
package main

import (
    "context"
    "fmt"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/project/mcp-go-core/core/server"
    "github.com/project/mcp-go-core/modules/transport/stdio"
    "my-mcp-server/tools"
)

func main() {
    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop()

    srv, err := server.NewBuilder().
        WithName("my-mcp-server").
        WithTimeout(30 * time.Second).
        WithTool(tools.GreetTool()).
        WithTransport(stdio.New(os.Stdin, os.Stdout)).
        Build()

    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }

    if err := srv.Run(ctx); err != nil && err != context.Canceled {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}
```

## Step 5: Run the Server

```bash
# Stdio transport (default — works with Claude Desktop)
go run ./cmd/my-mcp

# HTTP transport
go run ./cmd/my-mcp --transport http --addr localhost:8080

# Docker
make docker-build
docker run -p 8080:8080 my-mcp-server:latest \
  mcp-go-core run --transport http --addr 0.0.0.0:8080
```

## Step 6: Test with Claude Desktop

Add to `~/.config/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "my-mcp": {
      "command": "/path/to/my-mcp"
    }
  }
}
```

## Extending the Server

Add resources:

```go
srv, _ := server.NewBuilder().
    WithTool(tools.GreetTool()).
    WithResource(resources.ReadFileResource()).
    Build()
```

Add middleware:

```go
srv, _ := server.NewBuilder().
    WithTool(tools.GreetTool()).
    WithMiddleware(core.Middleware(logging), core.Middleware(recovery)).
    Build()
```

Switch transports:

```go
// HTTP
import "github.com/project/mcp-go-core/modules/transport/http"
tr := http.New("localhost:8080")

// SSE
import "github.com/project/mcp-go-core/modules/transport/sse"
tr := sse.New("localhost:8080")

// Stdio
import "github.com/project/mcp-go-core/modules/transport/stdio"
tr := stdio.New(os.Stdin, os.Stdout)
```

## API Reference for AI Agents

### Core Types

| Type | Package | Purpose |
|---|---|---|
| `server.Builder` | `core/server` | Fluent API for building servers |
| `tool.Tool` | `core/tool` | MCP tool interface |
| `resource.Resource` | `core/resource` | MCP resource interface |
| `prompt.Prompt` | `core/prompt` | MCP prompt interface |
| `protocol.Request` | `core/protocol` | JSON-RPC request |
| `protocol.Response` | `core/protocol` | JSON-RPC response |
| `transport.Transport` | `modules/transport` | Transport interface |

### Quick Mapping

| What You Need | Code |
|---|---|
| Build server | `server.NewBuilder()` |
| Register tool | `.WithTool(tool.NewTool(...))` |
| Register resource | `.WithResource(resource.NewResource(...))` |
| Register prompt | `.WithPrompt(prompt.NewPrompt(...))` |
| Set transport | `.WithTransport(stdio.New(os.Stdin, os.Stdout))` |
| Add middleware | `.WithMiddleware(mw1, mw2)` |
| Build | `.Build()` or `.MustBuild()` |

### Error Handling

```go
import "github.com/project/mcp-go-core/core/mcperror"

err := mcperror.NewError(mcperror.ErrCodeInvalidParams, "param 'x' is required")
return nil, err  // Will be serialized as JSON-RPC error response
```
