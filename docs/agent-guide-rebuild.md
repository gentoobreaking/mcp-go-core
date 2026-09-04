# AI Agent Guide: Rebuilding an Existing MCP with mcp-go-core

This guide shows an AI agent how to **rebuild an existing MCP server** using the mcp-go-core framework, preserving all original functionality.

## Prerequisites

- Go 1.26+
- Access to the original MCP source code
- Tools, resources, prompts, auth, and transport list from the original project
- Schema definitions from the original project

## Step 1: Analyze the Existing MCP

Run analysis to identify all components:

```bash
# Analyze original MCP source code
mcp-go-core analyze --source /path/to/original-mcp/

# Or manually identify:
# - Tools: name, description, input schema, handler logic
# - Resources: URI, name, description, read function
# - Prompts: name, description, get function
# - Auth: API key, JWT, or OAuth usage
# - Transport: stdio, HTTP, or SSE
# - Middleware: logging, recovery, metrics, tracing
```

## Step 2: Map Original Components

Create a mapping table from original MCP to mcp-go-core:

| Original Component | mcp-go-core Equivalent |
|---|---|
| JSON-RPC 2.0 layer | `core/protocol` |
| Tool dispatch | `core/tool` + `core/router` |
| Resource read | `core/resource` |
| Prompt get | `core/prompt` |
| stdio transport | `modules/transport/stdio` |
| HTTP transport | `modules/transport/http` |
| SSE transport | `modules/transport/sse` |
| JWT auth | `modules/security/jwt` |
| OAuth auth | `modules/security/oauth` |
| Metrics | `modules/middleware/metrics` |
| Tracing | `modules/middleware/tracing` |

## Step 3: Rebuild Each Tool

For each tool in the original MCP, create a wrapper in `tools/`:

```go
// tools/rebuilt_tool.go
package tools

import (
    "context"
    "github.com/project/mcp-go-core/core/protocol"
    "github.com/project/mcp-go-core/core/tool"
)

// OriginalSchema is the input schema from the original tool
var OriginalSchema = tool.Schema{
    "type": "object",
    "properties": map[string]any{
        "param1": map[string]any{"type": "string"},
        // ... copy from original
    },
    "required": []string{"param1"}, // ... from original
}

// RebuiltTool wraps the original handler logic
func RebuiltTool() tool.Tool {
    return tool.NewTool(
        "original_tool_name",       // same name
        "original description",     // same description
        OriginalSchema,             // copied schema
        func(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
            // 1. Parse req.Params (same as original)
            // 2. Call original handler logic
            result := callOriginalLogic(req)

            // 3. Return result wrapped in standard response
            return &protocol.Response{
                JSONRPC: "2.0",
                ID:      req.ID,
                Result:  result,
            }, nil
        },
    )
}
```

## Step 4: Rebuild Resources

For each resource:

```go
// resources/rebuilt_resource.go
package resources

import (
    "context"
    "github.com/project/mcp-go-core/core/protocol"
    "github.com/project/mcp-go-core/core/resource"
)

func RebuiltResource() resource.Resource {
    return resource.NewResource(
        "file://path/to/resource",  // URI template
        "original_resource_name",    // name
        "original description",      // description
        func(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
            // Original resource read logic
            data := callOriginalReadLogic(req)

            return &protocol.Response{
                JSONRPC: "2.0",
                ID:      req.ID,
                Result:  data,
            }, nil
        },
    )
}
```

## Step 5: Rebuild Prompts

For each prompt:

```go
// prompts/rebuilt_prompt.go
package prompts

import (
    "context"
    "github.com/project/mcp-go-core/core/prompt"
)

func RebuiltPrompt() prompt.Prompt {
    return prompt.NewPrompt(
        "original_prompt_name",
        "original description",
        func(ctx context.Context, req prompt.PromptRequest) (prompt.PromptResponse, error) {
            // Original prompt generation logic
            messages := callOriginalPromptLogic(req)

            return messages, nil
        },
    )
}
```

## Step 6: Recreate the Server

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
    "github.com/project/mcp-go-core/core/middleware"
    "github.com/project/mcp-go-core/modules/transport/stdio"
    "github.com/project/mcp-go-core/modules/middleware/metrics"
    "github.com/project/mcp-go-core/modules/middleware/tracing"
    auth "github.com/project/mcp-go-core/modules/security/jwt"
    "my-mcp/tools"
    "my-mcp/resources"
    "my-mcp/prompts"
)

func main() {
    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop()

    // Choose transport (match original)
    var tr server.Transport
    if useHTTP {
        tr = http.New("localhost:8080")
    } else {
        tr = stdio.New(os.Stdin, os.Stdout)
    }

    builder := server.NewBuilder().
        WithName("my-mcp-rebuilt").
        WithTimeout(30 * time.Second).
        WithTool(tools.RebuiltTool()).
        WithResource(resources.RebuiltResource()).
        WithPrompt(prompts.RebuiltPrompt()).
        WithTransport(tr)

    // Add auth middleware if original used JWT
    if useAuth {
        authenticator := auth.NewAuthenticator("secret", "issuer")
        builder = builder.WithMiddleware() // wire auth into handler
    }

    // Add metrics/tracing if desired
    if enableMetrics {
        m := metrics.NewMetrics()
        m.Configure()
        builder = builder.WithMiddleware(m.Middleware())
    }

    if enableTracing {
        t := tracing.NewTracing()
        t.Configure("my-mcp-rebuilt")
        builder = builder.WithMiddleware(t.Middleware())
    }

    srv, err := builder.Build()
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

## Step 7: Verify the Rebuild

```bash
# Compile
go build -o dist/my-mcp-rebuilt ./cmd/my-mcp

# Run unit tests
go test ./... -count=1

# Static analysis
go vet ./...

# Binary verification
mcp-go-core verify --binary dist/my-mcp-rebuilt

# Compare functionality — send same requests to original and rebuilt
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | dist/original-mcp
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | dist/my-mcp-rebuilt
# Outputs should match
```

## Step 8: Deploy

```bash
# Docker
docker build -t my-mcp-rebuilt:latest .
docker run -p 8080:8080 my-mcp-rebuilt:latest \
  mcp-go-core run --transport http --addr 0.0.0.0:8080

# K8s
mcp-go-core k8s --name my-mcp-rebuilt --image my-mcp-rebuilt:latest --port 8080
```

## Checklist for AI Agents

- [ ] Analyzed all tools/resources/prompts from original
- [ ] Mapped all transports (stdio/HTTP/SSE) correctly
- [ ] Preserved all schema definitions
- [ ] Migrated authentication (JWT/OAuth/API key)
- [ ] Migrated middleware (logging/recovery/metrics/tracing)
- [ ] Verified all handler logic produces identical results
- [ ] Tested with same input payloads as original
- [ ] Confirmed outputs match (or documented intentional differences)
- [ ] Ran `go test ./...` — all pass
- [ ] Ran `go vet ./...` — no issues
- [ ] Validated with `mcp-go-core verify`

## Key Differences When Rebuilding vs New

| Aspect | New MCP | Rebuilding Existing |
|---|---|---|
| Schema origin | Write from scratch | Copy from original |
| Handler logic | Implement new logic | Wrap original logic |
| Naming | Choose freely | Match original names |
| Tool set | Define from scratch | Map all existing tools |
| Testing | New test cases | Compare against original outputs |
| Compatibility | No compatibility constraints | Must be drop-in replacement |
| Auth | Choose new scheme | Match original auth method |
