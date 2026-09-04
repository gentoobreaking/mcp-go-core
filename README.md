# mcp-go-core

A modular Go framework for building MCP (Model Context Protocol) servers.

**Version:** v0.1.0  
**Go:** 1.26+  
**License:** Apache 2.0

---

## Overview

`mcp-go-core` provides a minimal, dependency-light MCP server framework with static composition and module isolation.

The framework follows a **Build Complete, Deploy Minimal** principle: the development environment provides the full framework; production builds include only selected modules through Go's dead-code elimination.

### Key Characteristics

- **Core isolation:** `core/` has zero external dependencies; it defines interfaces and types only
- **Module-based:** Optional capabilities (transports, security, middleware) live in `modules/`
- **No runtime reflection in hot paths:** Dispatch uses typed function calls, not `reflect` or `map[string]interface{}`
- **Deterministic builds:** Feature lock files ensure reproducible binary composition

---

## Features

### Core (always included)

| Feature | Description |
|---|---|
| MCP Protocol Types | JSON-RPC 2.0 message structures, error codes |
| Tool Registration | Register and dispatch tools via `core/tool` |
| Resource Registration | Register and dispatch resources via `core/resource` |
| Prompt Registration | Register and dispatch prompts via `core/prompt` |
| Router | Method-based dispatch (`tools/call`, `resources/read`, etc.) |
| Server Builder | Fluent API: `WithName`, `WithTool`, `WithResource`, `WithPrompt`, `WithTransport`, `WithMiddleware`, `Build` |
| Lifecycle Management | State machine: Created → Configured → Initialized → Started → Running → ShuttingDown → Shutdown |
| Structured Errors | JSON-RPC 2.0 error codes with `mcperror` package |

### Transports

| Feature | Package |
|---|---|
| stdio Transport | `modules/transport/stdio` |
| Streamable HTTP Transport | `modules/transport/http` |
| SSE Transport (with sessions) | `modules/transport/sse` |
| Transport Interface | `modules/transport` (unified `Transport` interface with `Serve` + `Close`) |
| Session Management | `modules/transport.SessionManager` with `NewSessionID()` |

### Security

| Feature | Package |
|---|---|
| API Key Authentication | `modules/security/api_key` |
| JWT Authentication | `modules/security/jwt` |
| OAuth 2.1 with PKCE | `modules/security/oauth` |

### Middleware & Observability

| Feature | Package |
|---|---|
| Core Middleware Chain | `core/middleware` (Logging, Recovery via `Chain`) |
| Structured Logging | `modules/middleware/logging` (text/JSON, level filtering, field support) |
| Recovery | `modules/middleware/recovery` (panic recovery, RecoveryError, mcperror integration) |
| Prometheus Metrics | `modules/middleware/metrics` (via `prometheus/client_golang`) |
| OpenTelemetry Tracing | `modules/middleware/tracing` (via `otel`) |

### Runtime

| Feature | Package |
|---|---|
| Task Management | `modules/runtime/task` (Task, Manager, Status, Result, cancellation) |
| Session Management | `modules/runtime/session` (Session, Manager, lifecycle integration) |

### Integration

| Feature | Package |
|||---|
| Kubernetes Manifests | `integration-kubernetes` (Deployment + Service YAML generation) |
### Testing

| Feature | Package |
|---|---|
| Test Utilities | `testutil` (EchoServer, MockTransport, TestSession, assertions) |

### Build & Tooling

| Feature | Package |
|---|---|
| CLI | `cmd/mcp-go-core` (init, build, run, generate, verify, k8s, version) |
| Feature Graph | `internal/featuregraph` (resolution, validation, lock) |
| Analyzer | `internal/analyzer` (AST-based feature inference) |
| Generator | `internal/generator` (static code generation) |
| Builder | `internal/builder` (build pipeline stages) |
| Manifest | `internal/manifest` (build manifest + checksums) |

### Performance

| Feature | Package |
|---|---|
| Dispatch Benchmark | `benchmarks` (P50/P99, throughput) |

---

## Architecture

```
┌──────────────────────────────────────────────────┐
│                    MCP Client                     │
├──────────────────────────────────────────────────┤
│              Transport Layer                       │
│        (stdio, HTTP, SSE + SessionManager)        │
├──────────────────────────────────────────────────┤
│                       CORE                          │
│   protocol │ server │ router │ tool │ resource    │
│   │ prompt │ lifecycle │ mcperror │ middleware     │
├──────────────────────────────────────────────────┤
│                  Middlewares                       │
│      (Logging, Recovery — in core/middleware)      │
├──────────────────────────────────────────────────┤
│        Optional Modules (import-free from core)   │
│  Security: api_key, jwt, oauth                     │
│  Observability: metrics, tracing                   │
│  Storage: memory                                   │
│  Integration: kubernetes                           │
└──────────────────────────────────────────────────┘
```

### Dependency Direction

```
Application  →  Modules  →  Core
```

- **Core** has no upward dependencies. It does NOT import security, observability, or integration modules.
- **Modules** depend on Core only. They never import each other across categories.
- **CLI** composes modules at startup based on user flags.

### Build Pipeline

```
Source → Configuration → Feature Analyzer → Feature Graph Resolver → Feature Lock
  → Code Generator → Build Manifest → Go Build → Binary Analyzer → Benchmark/Verification
```

Generated artifacts: `features.go`, `modules.go`, `router.go`, `server.go`, `buildinfo.go`.

---

## Project Structure

```text
mcp-go-core/
├── cmd/mcp-go-core/         # CLI tool (init, build, run, generate, verify, k8s, version)
├── core/                    # Core types & interfaces (zero external deps)
│   ├── protocol/            # JSON-RPC 2.0 types, error codes
│   ├── server/              # Server + Builder API
│   ├── router/              # Tool/Resource/Prompt dispatch
│   ├── tool/                # Tool interface + BaseTool
│   ├── resource/            # Resource interface + BaseResource
│   ├── prompt/              # Prompt interface + BasePrompt
│   ├── lifecycle/           # Lifecycle state machine
│   ├── mcperror/            # Structured error codes
│   └── middleware/          # Middleware chain (Logging, Recovery)
├── modules/                 # Optional implementations
│   ├── transport/           # Transport interface + SessionManager
│   │   ├── stdio/           # stdio transport
│   │   ├── http/            # Streamable HTTP transport
│   │   └── sse/             # SSE transport (with sessions, mark3labs/mcp-go)
│   ├── security/
│   │   ├── api_key/         # API Key authentication
│   │   ├── jwt/             # JWT authentication
│   │   ├── oauth/           # OAuth 2.1 + PKCE
│   │   └── mtls/            # mTLS (stub)
│   ├── middleware/
│   │   ├── logging/         # Structured logging (text/JSON)
│   │   ├── recovery/        # Panic recovery
│   │   ├── metrics/         # Prometheus metrics
│   │   └── tracing/         # OpenTelemetry tracing
│   ├── storage/
│   │   ├── memory/          # In-memory store
│   │   ├── filesystem/      # Filesystem store (path traversal protection)
│   │   └── external/        # Redis + PostgreSQL stores
│   └── runtime/
│       ├── task/            # Background task management
│       └── session/         # Session lifecycle management
│   ├── featuregraph/        # Feature descriptor, graph, resolution, lock
│   ├── analyzer/            # AST-based Go source analyzer
│   ├── generator/           # Static Go code generator
│   ├── builder/             # Build pipeline (Config→Analyze→Resolve→Lock→Generate→Compile→Verify)
│   └── manifest/            # Build manifest + checksum writer/reader
├── testutil/                # Test utilities (EchoServer, MockTransport, TestSession)
├── benchmarks/              # Dispatch & startup benchmarks
├── tests/                   # Integration, smoke, CI, negative tests
│   ├── build/               # Binary regression tests
│   ├── ci/                  # CI pipeline verification tests
│   ├── smoke/               # Runtime smoke tests (RT-001~RT-005)
│   ├── negative/            # Negative path tests
│   └── integration_test.go  # End-to-end tests
├── examples/                # Example MCP servers
│   └── minimal/             # Minimal MCP server example
├── go.mod
├── go.sum
├── Dockerfile
├── docker-compose.yml
├── .dockerignore
├── Makefile
├── LICENSE
└── README.md
```

---

## Requirements

- Go 1.26+
- Go toolchain supports CGO_ENABLED=0 for reproducible production builds

### Go Module Dependencies

| Module | Version | Purpose |
|---|---|---|
| `github.com/mark3labs/mcp-go` | v1.0.0 | SSE transport backend |
| `github.com/prometheus/client_golang` | v1.24.1 | Metrics middleware |
| `github.com/spf13/cobra` | v1.10.2 | CLI framework |
| `go.opentelemetry.io/otel` | v1.46.0 | Tracing middleware |
| `golang.org/x/oauth2` | v0.36.0 | OAuth PKCE support |
| `github.com/redis/go-redis/v9` | v9.22.0 | Redis storage backend |
| `github.com/lib/pq` | v1.12.3 | PostgreSQL storage backend |
| `k8s.io/api` | v0.37.0 | Kubernetes manifest types |
---

## Installation

```bash
go install github.com/project/mcp-go-core/cmd/mcp-go-core@latest
```

Build from source:

```bash
git clone <repo>
cd mcp-go-core
go build -o dist/mcp-go-core ./cmd/mcp-go-core
```

---

## Quick Start

### Run an MCP Server

Default (stdio transport):

```bash
mcp-go-core run --transport stdio
```

HTTP transport:

```bash
mcp-go-core run --transport http --addr localhost:8080
```

HTTP with metrics and tracing enabled:

```bash
mcp-go-core run \
  --transport http \
  --addr localhost:8080 \
  --metrics \
  --tracing
```

### Generate Kubernetes Manifests

```bash
mcp-go-core k8s --name my-mcp-server --image myregistry/mcp-server:v0.1 --port 8080 -o k8s/
```

This generates `k8s/my-mcp-server-deployment.yaml` and `k8s/my-mcp-server-service.yaml`.

### Verify a Binary

```bash
mcp-go-core verify --binary dist/mcp-go-core
```

---

## Usage

### CLI Commands

| Command | Description |
|---|---|
| `mcp-go-core init` | Initialize a new MCP project (`--name`, `--profile`) |
| `mcp-go-core build` | Build an MCP server binary (`--output`, `--profile`) |
| `mcp-go-core run` | Run an MCP server (`--transport`, `--addr`, `--metrics`, `--tracing`, `--oauth`) |
| `mcp-go-core generate` | Generate MCP source files (`--dry-run`) |
| `mcp-go-core verify` | Verify an MCP server binary (`--binary`) |
| `mcp-go-core k8s` | Generate Kubernetes manifests (`--name`, `--image`, `--port`, `--output`) |
| `mcp-go-core version` | Print version info (`-V`) |

#### `run` Flags

| Flag | Default | Description |
|---|---|---|
| `--transport` | `stdio` | Transport type: `stdio`, `http`, or `sse` |
| `--addr` | `localhost:8080` | Listen address (for http/sse) |
| `--metrics` | `false` | Enable Prometheus metrics endpoint |
| `--tracing` | `false` | Enable OpenTelemetry tracing |
| `--oauth` | `false` | Enable OAuth 2.1 authentication |

### Programmatic API

```go
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/project/mcp-go-core/core/server"
    "github.com/project/mcp-go-core/core/tool"
    "github.com/project/mcp-go-core/modules/transport/stdio"
)

func main() {
    // Define a tool
    greetTool := tool.NewTool(
        "greet",
        "Returns a greeting message",
        tool.Schema{"type": "object"},
        func(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
            return &protocol.Response{
                JSONRPC: "2.0",
                Result:  map[string]any{"message": "Hello from mcp-go-core!"},
            }, nil
        },
    )

    // Build and run server
    srv, err := server.NewBuilder().
        WithName("my-server").
        WithTool(greetTool).
        WithTransport(stdio.New(os.Stdin, os.Stdout)).
        Build()
    if err != nil {
        panic(err)
    }

    if err := srv.Run(context.Background()); err != nil {
        fmt.Fprintln(os.Stderr, err)
    }
}
```

### Transport Interface

All transports implement a unified interface:

```go
type Transport interface {
    Serve(ctx context.Context, handler Handler) error
    Close(ctx context.Context) error
}

type Handler func(ctx context.Context, msg json.RawMessage) (any, error)
```

Session management is available via `SessionManager`:

```go
sm := transport.NewSessionManager()
id := sm.RegisterSession()  // NewSessionID() generates random IDs
done := sm.GetSession(id)   // Returns done channel
sm.UnregisterSession(id)    // Closes and removes session
sm.CloseAll()               // Closes all sessions
```

### Server Builder

```go
srv, err := server.NewBuilder().
    WithName("my-server").
    WithTimeout(30 * time.Second).
    WithTool(myTool).
    WithResource(myResource).
    WithPrompt(myPrompt).
    WithTransport(stdio.New(os.Stdin, os.Stdout)).
    WithMiddleware(loggingMiddleware, recoveryMiddleware).
    Build()
```

### Security

**API Key:**

```go
auth := apikey.NewAuthenticator(map[string]apikey.Identity{
    "secret-key": {Principal: "user1", Scopes: []string{"read"}},
})
identity, err := auth.Authenticate(ctx, apikey.HTTPRequest{r})
```

**JWT:**

```go
auth := jwt.NewAuthenticator("hmac-secret", "my-issuer")
identity, err := auth.Authenticate(ctx, jwt.HTTPRequest{r})
```

**OAuth 2.1 with PKCE:**

```go
auth := oauth.NewAuthenticator(
    "https://auth.example.com",
    "mcp-client",
    []string{"read", "write"},
)
pkce, _ := oauth.GeneratePKCE()  // RFC 7636
```

### Storage

**External Storage (Redis/PostgreSQL):**

```go
// Redis
store := external.NewRedis(external.RedisConfig{
    Addr:     "localhost:6379",
    PoolSize: 10,
})

// PostgreSQL
pg, _ := external.NewPostgreSQL(external.PostgreSQLConfig{
    DSN:             "postgres://user:pass@localhost/db?sslmode=disable",
    MaxOpenConns:    25,
})

// All stores implement the Store interface
var s external.Store = store
```

### Runtime (Tasks & Sessions)

```go
// Task management with cancellation
tm := task.NewManager()
defer tm.Close()

t := tm.Create("my-task", func(ctx context.Context) (task.Result, error) {
    return task.Result{Data: []byte("done")}, nil
})

status, _ := tm.Status(t.ID)  // Check status
err := tm.Cancel(t.ID)        // Cancel if still running
```

```go
// Session management with lifecycle
sm := session.NewManager()
defer sm.CloseAll()

s, _ := sm.Create("client-session", map[string]any{"client_id": "claude"})
defer sm.Destroy(s.ID)
```

```go
spec := kubernetes.DeploymentSpec{
    Name:    "mcp-server",
    Image:   "localhost/mcp-server:latest",
    Port:    8080,
    Profile: "production",
}
kubernetes.WriteManifests(".", spec)
```

---

## API

### core/protocol

| Type | Description |
|---|---|
| `Message` | JSON-RPC 2.0 message (request or response) |
| `Request` | Incoming JSON-RPC request |
| `Response` | JSON-RPC response |
| `Error` | JSON-RPC error (code, message) |
| `InitializeRequest` | MCP initialize request params |
| `InitializeResponse` | MCP initialize response |
| `ServerCapabilities` | Server capability declaration |
| `ClientCapabilities` | Client capability declaration |
| `JSONRPCMessage` | Union type for JSON-RPC message parsing |
| `JSONRPCVersion` | `"2.0"` constant |

### core/tool

| Type/Interface | Description |
---|---|
| `Tool` | Interface: `Name()`, `Description()`, `InputSchema()`, `Handler()` |
| `BaseTool` | Default implementation |
| `NewTool(name, desc, schema, handler)` | Constructor |
| `Schema` | `map[string]any` (JSON Schema) |
| `ToolHandler` | `func(ctx, *Request) (*Response, error)` |

### core/resource

| Type/Interface | Description |
---|---|
| `Resource` | Interface: `URI()`, `Name()`, `Description()`, `Read(ctx, req)` |
| `BaseResource` | Default implementation |
| `NewResource(uri, name, desc, readFn)` | Constructor |

### core/prompt

| Type/Interface | Description |
---|---|
| `Prompt` | Interface: `Name()`, `Description()`, `Get(ctx, req)` |
| `BasePrompt` | Default implementation |
| `NewPrompt(name, desc, getFn)` | Constructor |
| `PromptRequest` | Request params |
| `PromptResponse` | Response with messages |

### core/mcperror

| Constant/Type | Value/Description |
---|---|
| `CodeProtocol` | `"protocol"` |
| `CodeValidation` | `"validation"` |
| `CodeInternal` | `"internal"` |
| `CodeTimeout` | `"timeout"` |
| `CodeCancellation` | `"cancellation"` |
| `ErrCodeParseError` | `-32700` |
| `ErrCodeInvalidRequest` | `-32600` |
| `ErrCodeMethodNotFound` | `-32601` |
| `ErrCodeInvalidParams` | `-32602` |
| `ErrCodeInternalError` | `-32603` |
| `NewError(code int, msg string) *Error` | Constructor |
| `NewParseError()` / `NewInvalidRequestError()` / etc. | Convenience constructors |
| `JSONRPCError` | JSON-serializable error |

### core/lifecycle

| State | Description |
|---|---|
| `Created` | Initial state |
| `Configured` | Options applied |
| `Initialized` | Ready to start |
| `Started` | Running |
| `Running` | Active |
| `ShuttingDown` | Graceful shutdown in progress |
| `Shutdown` | Fully stopped |

### modules/runtime/task

| Type | Description |
|---|---|
| `Status` | Task state: `pending`, `running`, `completed`, `failed`, `cancelled` |
| `Result` | Task outcome: `Data []byte`, `Err error` |
| `Task` | Background task with ID, status, creation/ completion timestamps |
| `Manager` | Thread-safe task manager: `Create`, `Cancel`, `Status`, `GetResult`, `WaitFor`, `RunningCount` |

### modules/runtime/session

| Type | Description |
|---|---|
| `Session` | Active session with ID, metadata, lifecycle state |
| `Manager` | Session manager: `Create`, `Get`, `Destroy`, `Count`, `ActiveSessions`, `Close`, `CloseAll` |


## Testing

```bash
# All tests
go test ./...

# With race detector
go test -race ./...

# Specific packages
go test ./core/...
go test ./modules/...
go test ./internal/...

# Benchmarks
go test -bench=. -benchmem ./benchmarks/...
```

### Test Counts

- **353 total tests, 0 failures**

| Type | Description |
|---|---|
| `EchoServer` | Test server that echoes messages |
| `MockTransport` | Mock `Transport` implementation with `Intercept` |
| `TestSession` | Test harness for session-based transport testing |
| `AssertJSONEqual`, `AssertJSONError`, etc. | Assertion helpers |

---

## Build

```bash
# Build everything
make build

# Build specific binary
go build -o dist/mcp-go-core ./cmd/mcp-go-core

# Reproducible build (CGO disabled)
CGO_ENABLED=0 go build -ldflags="-s -w" -o dist/mcp-go-core ./cmd/mcp-go-core
```

### Docker Build

```bash
# Build Docker image
make docker-build

# Run with docker-compose (HTTP + Prometheus + Grafana)
make docker-run

# Push to registry (requires DOCKER_REGISTRY and DOCKER_REGISTRY env)
make docker-push DOCKER_REGISTRY=ghcr.io/project

# Clean up
make docker-clean
```

### Build Pipeline

```bash
mcp-go-core init --name my-server --profile production
mcp-go-core generate --dry-run
mcp-go-core build --output dist/ --profile production --verify
mcp-go-core verify --binary dist/mcp-go-core
```

Stages: `config` → `analyze` → `resolve` → `lock` → `generate` → `compile` → `verify`.

---

## Deployment

### Minimal Profile

Stdio transport with no external dependencies:

```bash
mcp-go-core run --transport stdio
```

Binary includes only: `core`, `stdio` transport.

### Production Profile

HTTP transport:

```bash
mcp-go-core run --transport http --addr 0.0.0.0:8080
```

### Secure Profile

HTTP + JWT:

```bash
mcp-go-core run --transport http --addr 0.0.0.0:8080
```

### Observable Profile

HTTP + Metrics + Tracing:

```bash
mcp-go-core run --transport http --addr 0.0.0.0:8080 --metrics --tracing
```

### Containerized Deployment

Build and run with Docker:

```bash
mcp-go-core build --output dist/mcp-go-core --profile production
docker build -t mcp-go-core:v0.1.0 .
docker run -p 8080:8080 mcp-go-core:v0.1.0 \
  mcp-go-core run --transport http --addr 0.0.0.0:8080 --metrics
```

Local development with docker-compose (HTTP + Prometheus + Grafana):

```bash
docker compose --profile production up -d
```

| Service | URL | Profile |
|---|---|---|
| MCP Server (HTTP) | http://localhost:8080 | production |
| Prometheus | http://localhost:9090 | production |
| Grafana | http://localhost:3000 | production |

---

## Security

### Design Principles

- Core has **no authentication requirements** — auth is modular and opt-in
- API Key authentication via `modules/security/api_key`
- JWT validation via `modules/security/jwt` (HMAC-SHA256)
- OAuth 2.1 with PKCE via `modules/security/oauth` (RFC 7636)

### Security Verification

| Scenario | Test Cases |
|---|---|
| API Key | Valid: PASS, Invalid: Reject, Missing: Reject |
| JWT | Valid: PASS, Expired: Reject, Invalid signature: Reject, Missing: Reject |
| OAuth | PKCE generation (RFC 7636), Bearer token validation, token introspection |

mTLS module exists as a stub package (`modules/security/mtls`) — full implementation deferred to v2.

---

## Testing

```bash
# Full test suite
go test ./... -count=1

# With race detector
go test -race ./... -count=1

# Static analysis
go vet ./...

# Benchmarks
go test -bench=. -benchmem ./benchmarks/...
```

### Benchmarks

| Benchmark | Description |
|---|---|
| `BenchmarkToolDispatch` | Single tool dispatch latency |
| `BenchmarkToolDispatchInProcess` | In-process dispatch (no transport) |
| `BenchmarkToolDispatchThroughput` | Throughput measurement |
| `BenchmarkToolDispatchP50P99` | P50 and P99 latency |
| `BenchmarkStartup` | Process start to ready |
| `BenchmarkStartupMemory` | Memory usage at startup |

### Performance Targets

| Metric | Target |
|---|---|
| Dispatch P50 | < 10 µs |
| Dispatch P99 | < 100 µs |
| Throughput | > 100k req/s |
| Startup time | < 50 ms |
| Minimal RSS | < 20 MB |
| Production RSS | < 30 MB |

---

## Contributing

1. Ensure all tests pass: `go test ./...`
2. Run race detector: `go test -race ./...`
3. Run static analysis: `go vet ./...`
4. Maintain core isolation: `core/` must not import `modules/` or `integration-*`
5. Module isolation: transports must not import each other

---

## License

Apache 2.0
