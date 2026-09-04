# Architecture Documentation

## System Architecture

### Layer 1: Core (`core/`)
- `core/protocol` - JSON-RPC message types
- `core/request` - Request types for tools/resources/prompts
- `core/response` - Response types for tools/resources/prompts
- `core/mcperror` - Structured error types with codes
- `core/tool` - Tool interface and generic helpers
- `core/resource` - Resource interface
- `core/prompt` - Prompt interface
- `core/router` - Request dispatch routing
- `core/server` - Server lifecycle management
- `core/lifecycle` - Lifecycle state machine
- `core/middleware` - Middleware abstraction (logging, recovery)

Core never depends on any external runtime library.

### Layer 2: Internal (`internal/`)
- `internal/featuregraph` - Feature graph types, resolution engine, lock file generation
- `internal/analyzer` - Application source analysis for feature inference
- `internal/generator` - Static composition code generator
- `internal/builder` - Build pipeline orchestrator
- `internal/manifest` - Build manifest operations

Internal packages are only importable by the CLI and build tools.

### Layer 3: Modules (`modules/`)
- `modules/transport/{stdio,http,sse}` - Transport implementations
- `modules/security/{api_key,jwt,oauth,mtls}` - Authentication modules
- `modules/middleware/{logging,recovery,metrics,tracing}` - Middleware implementations
- `modules/storage/memory` - In-memory storage

Modules depend downward on Core only. No inter-module imports unless explicitly needed.

### Layer 4: CLI (`cmd/`)
- `cmd/mcp-go-core` - CLI tool with subcommands

### Layer 5: Examples (`examples/`)
- `examples/minimal` - Minimal MCP server example

## Feature Graph

Features are resolved at build time using a dependency graph:
- Explicit config in `mcp.yaml` has highest priority
- Generated metadata is inferred from `.mcp/generated/metadata.json`
- Known API patterns (e.g., `jwt.Configure()`)
- Go AST import scanning

Priority: Explicit Config > Generated Metadata > Known API > Go AST

## Build Pipeline

```
Config → Analyze → Resolve → Lock → Generate → Compile → Verify → Benchmark
```

1. **Config**: Read `mcp.yaml`
2. **Analyze**: Infer features from source
3. **Resolve**: Resolve features to deterministic closure
4. **Lock**: Write `.mcp/features.lock` with SHA256 hash
5. **Generate**: Generate static composition code
6. **Compile**: Build binary with only enabled modules
7. **Verify**: Binary audit (expected/unexpected modules)
8. **Benchmark**: Performance verification

## Dependency Rules

- Core → no dependencies on modules or external frameworks
- Module → depends on Core only
- Internal → not importable by applications
- Generated code → not importable by internal/analyzer
