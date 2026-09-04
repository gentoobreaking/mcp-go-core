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
- **Runtime operability:** Health endpoints, feature flags, rate limiting, and dynamic config are available at runtime

---

## Features

### Core (always included)

| Feature | Description |
|---|---|
| MCP Protocol Types | JSON-RPC 2.0 message structures, error codes |
| Tool Registration | Register and dispatch tools via `core/tool` |
| Resource Registration | Register and dispatch resources via `core/resource` |
| Prompt Registration | Register and dispatch prompts via `core/prompt` |
| Router | Method-based dispatch (`tools/call`, `resources/read`, etc.) with advanced MCP methods (ping, complete, roots, tasks, elicitation, subscriptions, progress, message) |
| Server Builder | Fluent API: `WithName`, `WithTool`, `WithResource`, `WithPrompt`, `WithTransport`, `WithMiddleware`, `WithHealth`, `WithFlags`, `WithRateLimiter`, `WithConfig`, `Build` |
| Lifecycle Management | State machine: Created → Configured → Initialized → Started → Running → ShuttingDown → Shutdown |
| Structured Errors | JSON-RPC 2.0 error codes with `mcperror` package |

### Transports

| Feature | Package |
|---|---|
| stdio Transport | `modules/transport/stdio` |
| Streamable HTTP Transport | `modules/transport/http` (with `WithHealthRoutes` for health endpoints) |
| SSE Transport (with sessions) | `modules/transport/sse` (with `WithHealthRoutes` for health endpoints) |
| Transport Interface | `modules/transport` (unified `Transport` interface with `Serve` + `Close`) |
| Session Management | `modules/transport.SessionManager` with `NewSessionID()` |

### Feature Flags & Rate Limiting

| Feature | Package |
|---|---|
| Feature Flags | `core/feature` (runtime `Flags`, `IsDisabled`, `Snapshot`) |
| Feature Wire Middleware | `core/middleware/featurewire` (method→flag gating, `HealthStatus`) |
| Rate Limiting | `core/middleware/ratelimit` (token-bucket `Manager`, `NewLimiter`, `Allow`) |
| Health Endpoints | `core/server` (`HealthHandler()`, `WithHealth(true)`) — wired into HTTP/SSE transports |
| Dynamic Config | `core/config` (hot-reload YAML via `fsnotify`, atomic `Load()`, `Health()`) |

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
| Task Management | `core/router` (`TaskResult`, `TaskStatus`, `RegisterTask`, `tasks/get`, `tasks/list`, `tasks/cancel`) |
| Session Management | `modules/runtime/session` (Session, Manager, lifecycle integration) |
| Elicitation | `core/router` (`ElicitationCreateParams`, `SetElicitationHandler`, `ResolveElicitation`) |
| Roots | `core/router` (`Root`, `ListRootsResult`, `SetRoots`, `SetRootsHandler`) |

### Integration

| Feature | Package |
|---|---|
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
│      (Logging, Recovery, RateLimit, FeatureWire)   │
├──────────────────────────────────────────────────┤
│        Optional Modules (import-free from core)    │
│  Security: api_key, jwt, oauth                     │
│  Observability: metrics, tracing                   │
│  Storage: memory, filesystem, external             │
│  Runtime: task, session                            │
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
│   ├── server/              # Server + Builder API (health, lifecycle)
│   ├── router/              # Tool/Resource/Prompt dispatch (advanced MCP methods)
│   ├── tool/                # Tool interface + BaseTool
│   ├── resource/            # Resource interface + BaseResource
│   ├── prompt/              # Prompt interface + BasePrompt
│   ├── lifecycle/           # Lifecycle state machine
│   ├── mcperror/            # Structured error codes
│   ├── feature/             # Runtime feature flags (Flags, IsDisabled, Snapshot)
│   ├── middleware/          # Middleware chain (Logging, Recovery)
│   │   ├── featurewire/     # Feature flag gating middleware
│   │   └── ratelimit/       # Token-bucket rate limiting manager
│   └── config/              # Hot-reloadable YAML config via fsnotify
├── modules/                 # Optional implementations
│   ├── transport/           # Transport interface + SessionManager
│   │   ├── stdio/           # stdio transport
│   │   ├── http/            # Streamable HTTP transport (health routes)
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
├── internal/                
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
| `golang.org/x/time` | v0.11.0 | Token-bucket rate limiting |
| `github.com/fsnotify/fsnotify` | v1.9.x | Config file hot-reload |
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

HTTP with metrics, tracing, health, and feature flags enabled:

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
    "github.com/project/mcp-go-core/core/protocol"
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

### Health Endpoints

Enable HTTP health check endpoints on HTTP or SSE transports:

```go
srv, err := server.NewBuilder().
    WithName("my-server").
    WithTransport(http.New(addr)).
    WithHealth(true).         // Enable health routes on the transport
    WithFlags(feature.NewFlags(initialFlags)).
    WithRateLimiter(ratelimit.NewManager()).
    WithConfig(config.NewConfig("config.yaml")).
    Build()
```

The transport registers a health mux that takes priority for `/health*` paths, falling through to the main MCP handler. The following routes are available:

| Route | Description |
|---|---|
| `GET /health` | Basic liveness check (`{"status":"ok"}`) |
| `GET /health/features` | All feature flag statuses (requires `WithFlags`) |
| `GET /health/features/<name>` | Individual flag status |
| `GET /health/rate-limits` | Rate limiter statuses (requires `WithRateLimiter`) |
| `GET /health/config` | Config hot-reload metadata (requires `WithConfig`) |

You can also retrieve the health handler directly:

```go
handler := srv.HealthHandler()
// handler is nil if WithHealth(true) was not called
```

### Feature Flags

Runtime feature flags allow enabling/disabling MCP methods per-server without restart:

```go
import "github.com/project/mcp-go-core/core/feature"

flags := feature.NewFlags(map[string]bool{
    "tools/call:advanced":  true,   // enable advanced tool
    "resources/read:secret": false,  // gate a sensitive resource
})

srv, _ := server.NewBuilder().
    WithName("my-server").
    WithFlags(flags).
    WithMiddleware(featurewire.Middleware(flags, featurewire.DefaultFlagMapper)).
    Build()

// Toggle at runtime — next request sees the new state
flags.Set("tools/call:advanced", false)

// Check status
snapshots := flags.Snapshot()  // map[string]feature.Flag
enabled := flags.EnabledList() // []string of enabled flag names
```

The `featurewire` middleware maps MCP method names to flag names via a `FlagMapper` function. The default mapper converts e.g. `tools/call:advanced` → `advanced_tools`.

### Rate Limiting

Token-bucket rate limiting per MCP method, backed by `golang.org/x/time/rate`:

```go
import "github.com/project/mcp-go-core/core/middleware/ratelimit"

lim := ratelimit.NewManager()
lim.Allow("tools/call")  // returns nil if allowed, ratelimit.ErrRateLimit if exceeded

// Status for health endpoint
for _, st := range lim.Status() {
    fmt.Println(st.Name, st.Limit, st.Burst)
}
```

| Feature | Package | Methods Limited |
|---|---|---|
| `DefaultLimits` | `core/middleware/ratelimit` | `tools/call`, `tools/list`, `prompts/get`, `prompts/list`, `resources/read`, `resources/list` |

### Dynamic Configuration

Hot-reloadable YAML configuration with atomic swap via `fsnotify`:

```go
import "github.com/project/mcp-go-core/core/config"

cfg := config.NewConfig("config.yaml")
cfg.Load() // reads YAML, atomically swaps config

// Watch for file changes
watcher, _ := config.NewWatcher(cfg, func(format string, args ...any) {
    log.Printf(format, args...)  // logs reload events
})
defer watcher.Close()

// Health endpoint exposes config metadata
health := cfg.Health()
// HealthInfo{LastLoaded time.Time, LastLoadErr error, ...}
```

```yaml
# config.yaml
server:
  port: 8080
features:
  tools/call: true
rate_limits:
  tools/call:
    rate: 10
    burst: 20
```

### Resource Notifications & Subscriptions

The router tracks per-client resource subscriptions and emits change notifications:

```go
// Subscribe a client to resource changes
router.Subscribe(uri, "client-1")

// Check subscription status
router.IsSubscribed(uri)  // true if any client subscribed

// Notify subscribers of a resource update
router.NotifyResourceUpdate(uri, "update")

// Delete a resource and notify subscribers before clearing subscriptions
router.DeleteResource(uri)

// Unsubscribe a specific client
router.UnsubscribeClient(uri, "client-1")

// Remove all subscriptions for a URI
router.Unsubscribe(uri)
```

The subscription map is `map[string]map[string]bool` (URI → clientIDs). Client IDs are extracted from context via `clientIDFromContext()` with a `"default"` fallback.

Notifications are sent via `NotificationSender` — a callback function wired from the transport layer:

```go
router.SetNotificationSender(func(method string, params any) error {
    // Serialize and send to connected clients via transport
    return transport.Send(method, params)
})
```

Protocol types for resource notifications:

| Type | Description |
|---|---|
| `ResourceUpdateNotification` | `notifications/resources/update` — sent when a subscribed resource changes |
| `ResourceUpdateParams` | Params with `URI` and `ChangeType` (`"update"` or `"delete"`) |
| `ResourceDeletedNotification` | `notifications/resources/deleted` — sent when a subscribed resource is deleted |
| `ResourceDeleteParams` | Params with `URI` of the deleted resource |
| `SubscribeParams` | `resources/subscribe` request params (`URI`) |
| `Subscription` | Active subscription with `URI` and `SubscribedAt` |
| `UnsubscribeParams` | `resources/unsubscribe` request params (`URI`) |

### Advanced MCP Methods

The router implements the full suite of MCP protocol methods:

| Method | Direction | Description |
|---|---|---|
| `ping` | C→S | Liveness check, returns `PingResult{"pong"}` |
| `complete/arg` / `complete/prompt` | C→S | Argument/prompt value completion |
| `roots/list` | C→S | List root directories provided by client |
| `notifications/roots/list_changed` | C→S | Client notifies server that roots changed |
| `prompts/create` | C→S | Dynamically register a prompt at runtime |
| `notifications/prompts/list_changed` | C→S | Client notifies server prompt list changed |
| `notifications/resources/list_changed` | C→S | Client notifies server resource list changed |
| `notifications/tools/list_changed` | C→S | Client notifies server tool list changed |
| `resources/templates/list` | C→S | List resource URI templates |
| `resources/subscribe` | C→S | Subscribe to resource updates |
| `resources/unsubscribe` | C→S | Remove a subscription |
| `notifications/progress` | S→C / C→S | Progress updates with `ProgressToken` |
| `notifications/message` | S→C / C→S | Log messages from client to server |
| `notifications/resources/update` | S→C | Notify subscribers of resource changes |
| `notifications/resources/deleted` | S→C | Notify subscribers a resource was deleted |
| `elicitation/create` | S→C | Server requests input from client |
| `notifications/elicitation/complete` | C→S | Client responds to elicitation |
| `tasks/get` | C→S | Get a task by ID |
| `tasks/list` | C→S | List all tasks |
| `tasks/cancel` | C→S | Cancel a task by ID |
| `server/discover` | C→S | Discover server capabilities and protocol version |
| `subscriptions/listen` | C→S | Register a subscription listener |

#### Tool & Prompt Creation

Tools and prompts can be registered dynamically at runtime:

```go
// Dynamic tool registration via tools/create (if supported)
// Tools are registered on the router at build time via Builder.WithTool()
// or added post-build via server.AddTool()

// Dynamic prompt registration via prompts/create
router.SetPromptCreator(func(params protocol.PromptCreateParams) (prompt.Prompt, error) {
    return prompt.NewPrompt(
        params.Name,
        params.Description,
        func(ctx context.Context, req protocol.PromptRequest) (*protocol.PromptResponse, error) {
            return &protocol.PromptResponse{
                Messages: []protocol.SamplingMessage{
                    {Role: "assistant", Content: []protocol.Content{{Type: "text", Text: "Hello"}}},
                },
            }, nil
        },
    ), nil
})
```

#### Task Management

```go
// Register tasks on the router
router.RegisterTask("task-1", protocol.TaskStatusRunning, nil)
router.RegisterTask("task-2", protocol.TaskStatusCompleted, "result-data")

// Query via MCP methods:
//   tasks/list   → returns TaskListResult with all tasks
//   tasks/get    → returns TaskResult by ID
//   tasks/cancel → cancels a running task
```

#### Elicitation

Server requests information from the client (e.g., missing parameters):

```go
router.SetElicitationHandler(func(params protocol.ElicitationCreateParams) error {
    // Transport sends elicitation/create to client
    // Client responds via notifications/elicitation/complete
    return nil
})

// Resolve the client's response
result, ok := router.ResolveElicitation(elicitationID)
```

---

### Programmatic API

```go
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/project/mcp-go-core/core/server"
    "github.com/project/mcp-go-core/core/tool"
    "github.com/project/mcp-go-core/core/protocol"
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
| `JSONRPCVersion` | `"2.0"` constant |
| `JSONRPCMessage` | Union type for JSON-RPC message parsing |
| `InitializeRequest` | MCP initialize request params |
| `InitializeResponse` | MCP initialize response |
| `InitializeParams` | Initialize request parameters |
| `InitializeResult` | Initialize response data |
| `ServerCapabilities` | Server capability declaration |
| `ClientCapabilities` | Client capability declaration |
| `Implementation` | Client/server implementation info |

#### Notifications (server → client)

| Type | Description |
|---|---|
| `Notification` | JSON-RPC notification (no ID) |
| `ResourceUpdateNotification` | `notifications/resources/update` — resource changed |
| `ResourceUpdateParams` | Params with `URI` and `ChangeType` |
| `ResourceDeletedNotification` | `notifications/resources/deleted` — resource removed |
| `ResourceDeleteParams` | Params with `URI` of deleted resource |
| `ToolListChangedNotification` | `notifications/tools/list_changed` |
| `PromptListChangedNotification` | `notifications/prompts/list_changed` |
| `LoggingMessage` | `logging/message` — server log to client |
| `ProgressNotification` | `notifications/progress` — progress update |
| `ProgressNotificationParams` | Params with `ProgressToken`, `Progress`, `Total` |
| `MessageNotification` | `notifications/message` — server-logged message |
| `MessageNotificationParams` | Params with `Level`, `Logger`, `Data` |
| `CreatedNotification` | `notifications/resources/created` — resource created |
| `CreatedParams` | Params with `URI` and `Name` |

#### Notifications (client → server)

| Type | Description |
|---|---|
| `ElicitationCompleteParams` | `notifications/elicitation/complete` — client responds to elicitation |
| `RootsListChangedParams` | `notifications/roots/list_changed` — client roots changed |

#### Capabilities

| Struct | Fields |
|---|---|
| `PromptsCapability` | `ListAvailable: bool` |
| `ResourcesCapability` | `ListAvailable`, `Subscribe`, `Create` |
| `ToolsCapability` | `ListAvailable`, `Create`, `ListChanged` |
| `LoggingCapability` | `Log: bool` |
| `CompletionsCapability` | `Complete: bool` |
| `RootsCapability` | `ListAvailable: bool` |
| `SamplingCapability` | `CreateMessage: bool` |

#### Requests & Results

| Type | Description |
|---|---|
| `PromptListParams` / `PromptListResult` | List prompts |
| `ResourceListParams` / `ResourceListResult` | List resources |
| `ToolListParams` / `ToolListResult` | List tools |
| `SubscribeParams` | `resources/subscribe` params (`URI`) |
| `UnsubscribeParams` | `resources/unsubscribe` params (`URI`) |
| `SubscriptionListenParams` | `subscriptions/listen` params (`URI`) |
| `CompletionParams` / `CompleteResult` | Argument/prompt completion |
| `CreateMessageParams` / `CreateMessageResult` | `sampling/createMessage` |
| `ElicitationCreateParams` / `ElicitationResult` | `elicitation/create` |
| `TaskCancelParams` | `tasks/cancel` params (`ID`) |
| `TaskResult` | Task with `ID`, `Status`, `Result` |
| `TaskListResult` | List of `TaskResult` |
| `ListRootsResult` | `roots/list` result (`Roots []Root`) |
| `Root` | Root URI, name, description |
| `Subscription` | Active subscription (`URI`, `SubscribedAt`) |
| `ResourceTemplate` | URI template for resources |
| `ResourceTemplateListResult` | List of resource templates |
| `Prompt` | Prompt definition |
| `Resource` | Resource reference |
| `Tool` | Tool definition |
| `Argument` | Parameter/argument definition |
| `NotificationSender` | `func(method string, params any) error` — sends notifications to clients |

#### Error Types

| Constant/Type | Value/Description |
|---|---|
| `CodeProtocol` | `"protocol"` |
| `CodeValidation` | `"validation"` |
| `CodeInternal` | `"internal"` |
| `CodeTimeout` | `"timeout"` |
| `CodeCancellation` | `"cancellation"` |
| `ErrCodeParseError` | `-32700` (parse_error) |
| `ErrCodeInvalidRequest` | `-32600` (invalid_request) |
| `ErrCodeMethodNotFound` | `-32601` (method_not_found) |
| `ErrCodeInvalidParams` | `-32602` (invalid_params) |
| `ErrCodeInternalError` | `-32603` (internal_error) |
| `NewError(code int, msg string) *Error` | Constructor |
| `NewParseError()` / `NewInvalidRequestError()` / etc. | Convenience constructors |
| `JSONRPCError` | JSON-serializable error |

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

### core/feature

| Type | Description |
|---|---|
| `Flag` | `{Enabled bool}` — feature flag state |
| `Flags` | Thread-safe store: `Get`, `Set`, `IsDisabled`, `Snapshot`, `EnabledList` |
| `NewFlags(map[string]bool)` | Constructor |

### core/middleware/featurewire

| Type | Description |
|---|---|
| `FlagMapper` | `func(method string) string` — maps MCP method → flag name |
| `DefaultFlagMapper` | Default mapper (e.g. `tools/call:advanced` → `advanced_tools`) |
| `Middleware(flags, mapper)` | Middleware that gates methods by flag |
| `HealthStatus(flags)` | Returns flag statuses for health endpoints |
| `FlagStatus` | JSON-serializable flag status (`Name`, `Enabled`) |

### core/middleware/ratelimit

| Type | Description |
|---|---|
| `Limiter` | Token-bucket limiter with `name`, `limit`, `burst` |
| `NewLimiter(name, rate, burst)` | Constructor |
| `Manager` | Thread-safe manager: `NewManager`, `Init`, `Allow`, `Status`, `AllowAll`, `RejectAll` |
| `Status` | `{Name, Limit, Burst, ...}` — JSON-serializable |
| `DefaultLimits` | Default per-method limits map |

### core/config

| Type | Description |
|---|---|
| `Config` | Hot-reloadable config: `Load`, `GetServer`, `GetFeatures`, `GetRateLimits`, `Health` |
| `ServerConfig` | Server-level settings |
| `LimitConfig` | Rate limit config (`Rate`, `Burst`) |
| `LoggingConfig` | Logging settings |
| `HealthInfo` | Config health metadata |
| `NewConfig(path)` | Create config from YAML path |
| `NewWatcher(cfg, logger)` | File watcher via fsnotify |

### core/router

| Type/Interface | Description |
---|---|
| `Router` | Central dispatch: `RegisterTool`, `RegisterResource`, `RegisterPrompt`, `Dispatch` |
| `SamplingHandler` | `func(ctx, *CreateMessageParams) (*CreateMessageResult, error)` |
| `CreatedNotifier` | `func(uri, name string) error` — callback for resource creation |
| `PromptCreator` | `func(params PromptCreateParams) (Prompt, error)` — factory for `prompts/create` |
| `DeleteResource(uri)` | Delete resource + notify subscribers |
| `NotifyResourceDeleted(uri)` | Send `notifications/resources/deleted` to subscribers |
| `NotifyResourceUpdate(uri, changeType)` | Send resource update notification |
| `Subscribe(uri, clientID)` | Per-client subscription |
| `UnsubscribeClient(uri, clientID)` | Remove one client's subscription |
| `Unsubscribe(uri)` | Remove all subscriptions for a URI |
| `IsSubscribed(uri)` | Check if URI has subscribers |
| `SetNotificationSender(handler)` | Wire notification delivery to transport |
| `SetSampler(h)` | Register sampling handler |
| `SetProgressHandler(h)` | Register progress callback |
| `SetMessageHandler(h)` | Register message callback |
| `SetElicitationHandler(h)` | Register elicitation callback |
| `SetPromptCreator(fn)` | Register prompt factory for `prompts/create` |
| `SetPromptListChangedHandler(h)` | Register prompts/list_changed callback |
| `SetResourceListChangedHandler(h)` | Register resources/list_changed callback |
| `SetToolsListChangedHandler(h)` | Register tools/list_changed callback |
| `SetRootsHandler(h)` | Register roots/list_changed callback |
| `SetRoots(roots)` | Set client-provided roots |
| `RegisterTask(id, status, result)` | Register a task in the registry |
| `ResolveElicitation(id)` | Get elicitation result by ID |
| `HandleError(err)` | Error handler interface |

### core/server

| Type | Description |
|---|---|
| `Server` | MCP server with lifecycle, health, gating |
| `Builder` | Fluent server builder |
| `Option` | `func(*Server)` — server option |
| `Option` functions | `WithName`, `WithVersion`, `WithFlags`, `WithRateLimiter`, `WithHealth`, `WithConfig`, `WithTransport`, `WithMiddleware`, `WithTimeout` |
| `NewServer(opts...)` | Functional constructor |
| `NewBuilder()` | Builder constructor |
| `HealthHandler()` | Returns HTTP handler for health routes (nil if health not enabled) |
| `SendNotification(method, params)` | Broadcast notification to clients |
| `AddTool(t)` | Register a tool post-build |
| `AddResource(r)` | Register a resource post-build |
| `AddPrompt(p)` | Register a prompt post-build |
| `Run(ctx)` | Start server (blocks until ctx cancelled) |
| `Shutdown(ctx)` | Graceful shutdown |

### core/middleware

| Type | Description |
|---|---|
| `Middleware` | `func(Handler) Handler` |
| `Handler` | `interface{ Dispatch(ctx, method, params) (any, error) }` |
| `HandlerFunc` | Adapter for plain functions |
| `Chain(mw...)` | Compose middleware into a chain |

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
| `Task` | Background task with ID, status, creation/completion timestamps |
| `Manager` | Thread-safe task manager: `Create`, `Cancel`, `Status`, `GetResult`, `WaitFor`, `RunningCount` |

### modules/runtime/session

| Type | Description |
|---|---|
| `Session` | Active session with ID, metadata, lifecycle state |
| `Manager` | Session manager: `Create`, `Get`, `Destroy`, `Count`, `ActiveSessions`, `Close`, `CloseAll` |


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

### Test Counts

- **366 total tests, 0 failures** (excluding feature-flagged new tests)
- Additional tests: `core/server/health_test.go` (11 tests), `core/router/router_test.go` (3 tests for resource deletion/subscriptions)

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

## License

Apache 2.0
