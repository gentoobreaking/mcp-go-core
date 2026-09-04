# mcp-go-core

一個用 Go 語言構建 MCP（Model Context Protocol）伺服器的模組化框架。

**版本：** v0.1.0  
**Go：** 1.26+  
**授權：** Apache 2.0

---

## 總覽

`mcp-go-core` 提供一個最小化、低依賴的 MCP 伺服器框架，具備靜態組合與模組隔離特性。

框架遵循 **「完整建置，最小部署」** 的原則：開發環境提供完整框架；生產部署僅包含所需模組，依靠 Go 的死碼消除機制。

### 核心特性

- **核心隔離:** `core/` 擁有零外部依賴，僅定義介面與型別
- **模組化:** 選用的功能（傳輸層、安全、中介軟體）位於 `modules/`
- **無熱路徑反射:** 呼叫使用型別化函式，避免 `reflect` 或 `map[string]interface{}`
- **確定性建置:** 功能鎖定檔案確保二進制組合可重現

---

## 功能列表

### 核心功能（始終包含）

| 功能 | 說明 |
|---|---|
| MCP 協定型別 | JSON-RPC 2.0 訊息結構、錯誤碼 |
| 工具註冊 | 透過 `core/tool` 註冊與調度工具 |
| 資源註冊 | 透過 `core/resource` 註冊與調度資源 |
| Prompt 註冊 | 透過 `core/prompt` 註冊與調度 Prompt |
| 路由器 | 基於方法的派發（`tools/call`、`resources/read` 等） |
| 伺服器建造者 | 流暢 API：`WithName`、`WithTool`、`WithResource`、`WithPrompt`、`WithTransport`、`WithMiddleware`、`Build` |
| 生命週期管理 | 狀態機：Created → Configured → Initialized → Started → Running → ShuttingDown → Shutdown |
| 結構化錯誤 | 具備 `mcperror` 套件的 JSON-RPC 2.0 錯誤碼 |

### 傳輸層

| 功能 | 套件 |
|---|---|
| stdio 傳輸 | `modules/transport/stdio` |
| Streamable HTTP 傳輸 | `modules/transport/http` |
| SSE 傳輸（含會話） | `modules/transport/sse` |
| 傳輸介面 | `modules/transport`（統一 `Transport` 介面：`Serve` + `Close`） |
| 會話管理 | `modules/transport.SessionManager` 與 `NewSessionID()` |

### 安全

| 功能 | 套件 |
|---|---|
| API Key 驗證 | `modules/security/api_key` |
| JWT 驗證 | `modules/security/jwt` |
| OAuth 2.1 + PKCE | `modules/security/oauth` |

### 中介軟體與可觀察性

| 功能 | 套件 |
|---|---|
| 核心中介軟體鏈 | `core/middleware`（Logging、Recovery 透過 `Chain`)|
| 結構化日誌 | `modules/middleware/logging`（文字/JSON、層級過濾、欄位支援） |
| Panic 復原 | `modules/middleware/recovery`（RecoveryError、mcperror 整合） |
| Prometheus 監控 | `modules/middleware/metrics`（基於 `prometheus/client_golang`） |
| OpenTelemetry 追蹤 | `modules/middleware/tracing`（基於 `otel`） |

### 儲存

| 功能 | 套件 |
|---|---|
| 記憶體型儲存 | `modules/storage/memory` |
| 檔案系統儲存 | `modules/storage/filesystem`（路徑遍歷保護、context 感知） |
| Redis 儲存 | `modules/storage/external`（連線池、TTL、SCAN） |
| PostgreSQL 儲存 | `modules/storage/external`（upsert、前綴掃描、資料表支援） |

### 執行期

| 功能 | 套件 |
|---|---|
| 任務管理 | `modules/runtime/task`（Task、Manager、Status、Result、取消支援） |
| 會話管理 | `modules/runtime/session`（Session、Manager、生命週期整合） |

### 整合

| 功能 | 套件 |
|---|---|
| Kubernetes 清單生成 | `integration-kubernetes`（Deployment + Service YAML） |

### 測試

| 功能 | 套件 |
|---|---|
| 測試工具 | `testutil`（EchoServer、MockTransport、TestSession、斷言） |

### 建置與工具

| 功能 | 套件 |
|---|---|
| CLI | `cmd/mcp-go-core`（init、build、run、generate、verify、k8s、version） |
| 功能圖 | `internal/featuregraph`（解析、驗證、鎖定） |
| 分析器 | `internal/analyzer`（基於 Go AST 的功能推斷） |
| 產生器 | `internal/generator`（靜態 Go 程式碼生成） |
| 建構器 | `internal/builder`（建置管道：Config→Analyze→Resolve→Lock→Generate→Compile→Verify） |
| 清單 | `internal/manifest`（建置清單 + 校驗和） |

### 效能

| 功能 | 套件 |
|---|---|
| 效能測試 | `benchmarks`（P50/P99、吞吐量） |

---

## 架構

```
┌──────────────────────────────────────────────────┐
│                    MCP 客戶端                     │
├──────────────────────────────────────────────────┤
│              傳輸層                                │
│        (stdio, HTTP, SSE + SessionManager)        │
├──────────────────────────────────────────────────┤
│                       核心                          │
│   protocol │ server │ router │ tool │ resource    │
│   │ prompt │ lifecycle │ mcperror │ middleware     │
├──────────────────────────────────────────────────┤
│                  中介軟體                           │
│      (Logging, Recovery — 位於 core/middleware)    │
├──────────────────────────────────────────────────┤
│        選用模組（core 無向上反向依賴）             │
│  Security: api_key, jwt, oauth                     │
│  Observability: metrics, tracing                   │
│  Storage: memory                                   │
│  Integration: kubernetes                           │
└──────────────────────────────────────────────────┘
```

### 依賴方向

```
應用程式  →  模組  →  核心
```

- **Core** 毋須任何向上的依賴。它不會引入 security、observability 或 integration 模組。
- **Modules** 僅依賴 Core。跨類別模組不會互相引入。
- **CLI** 根據使用者旗標組合模組於啟動時。

### 建置管道

```
原始碼 → 配置 → 功能分析器 → 功能圖解析器 → 功能鎖定
  → 程式碼產生器 → 建置清單 → Go 編譯 → 二進制分析 → 效能測試/驗證
```

產生的成果：`features.go`、`modules.go`、`router.go`、`server.go`、`buildinfo.go`。

---

## 專案結構

```text
mcp-go-core/
├── cmd/mcp-go-core/         # CLI 工具（init、build、run、generate、verify、k8s、version）
├── core/                    # 核心型別與介面（零外部依賴）
│   ├── protocol/            # JSON-RPC 2.0 型別、錯誤碼
│   ├── server/              # Server + Builder API
│   ├── router/              # Tool/Resource/Prompt 調度
│   ├── tool/                # Tool 介面 + BaseTool
│   ├── resource/            # Resource 介面 + BaseResource
│   ├── prompt/              # Prompt 介面 + BasePrompt
│   ├── lifecycle/           # 生命週期狀態機
│   ├── mcperror/            # 結構化錯誤碼
│   └── middleware/          # 中介軟體鏈（Logging、Recovery）
├── modules/                 # 選用實作
│   ├── transport/           # 傳輸介面 + SessionManager
│   │   ├── stdio/           # stdio 傳輸
│   │   ├── http/            # Streamable HTTP 傳輸
│   │   └── sse/             # SSE 傳輸（含會話、基於 mark3labs/mcp-go）
│   ├── security/
│   │   ├── api_key/         # API Key 驗證
│   │   ├── jwt/             # JWT 驗證
│   │   ├── oauth/           # OAuth 2.1 + PKCE
│   │   └── mtls/            # mTLS（預留）
│   ├── middleware/
│   │   ├── logging/         # 結構化日誌（文字/JSON）
│   │   ├── recovery/        # Panic 復原
│   │   ├── metrics/         # Prometheus 監控
│   │   └── tracing/         # OpenTelemetry 追蹤
│   ├── storage/
│   │   ├── memory/          # 記憶體型儲存
│   │   ├── filesystem/      # 檔案系統儲存（路徑遍歏保護）
│   │   └── external/        # Redis + PostgreSQL 儲存
│   └── runtime/
│       ├── task/            # 背景任務管理
│       └── session/         # 會話生命週期管理
├── integration-kubernetes/  # Kubernetes 清單生成
├── internal/                # 建置時工具（不可被使用者引入）
│   ├── featuregraph/        # 功能描述、圖、解析、鎖定
│   ├── analyzer/            # 基於 Go AST 的原始碼分析器
│   ├── generator/           # 靜態 Go 程式碼產生器
│   ├── builder/             # 建置管道（Config→Analyze→Resolve→Lock→Generate→Compile→Verify）
│   └── manifest/            # 建置清單 + 校驗和讀寫
├── testutil/                # 測試工具（EchoServer、MockTransport、TestSession）
├── benchmarks/              # 調度與啟動效能測試
├── tests/                   # 整合、冒煙、CI、負向測試
│   ├── build/               # 二進制回歸測試
│   ├── ci/                  # CI 管道驗證測試
│   ├── smoke/               # 執行期冒煙測試（RT-001~RT-005）
│   ├── negative/            # 負向路徑測試
│   └── integration_test.go  # 端對端測試
├── examples/                # 範例 MCP 伺服器
│   └── minimal/             # 最小化 MCP 伺服器範例
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

## 需求

- Go 1.26+
- 支援 `CGO_ENABLED=0` 以確保生產建置可重現

### Go 模組依賴

| 模組 | 版本 | 用途 |
|---|---|---|
| `github.com/mark3labs/mcp-go` | v1.0.0 | SSE 傳輸後端 |
| `github.com/prometheus/client_golang` | v1.24.1 | 監控中介軟體 |
| `github.com/spf13/cobra` | v1.10.2 | CLI 框架 |
| `go.opentelemetry.io/otel` | v1.46.0 | 追蹤中介軟體 |
| `k8s.io/api` | v0.37.0 | Kubernetes 清單型別 |
| `github.com/redis/go-redis/v9` | v9.22.0 | Redis 儲存後端 |
| `github.com/lib/pq` | v1.12.3 | PostgreSQL 儲存後端 |

---

## 安裝

```bash
go install github.com/project/mcp-go-core/cmd/mcp-go-core@latest
```

從原始碼建置：

```bash
git clone <repo>
cd mcp-go-core
go build -o dist/mcp-go-core ./cmd/mcp-go-core
```

---

## 快速入門

### 執行 MCP 伺服器

預設（stdio 傳輸）：

```bash
mcp-go-core run --transport stdio
```

HTTP 傳輸：

```bash
mcp-go-core run --transport http --addr localhost:8080
```

啟用監控與追蹤的 HTTP：

```bash
mcp-go-core run \
  --transport http \
  --addr localhost:8080 \
  --metrics \
  --tracing
```

### 生成 Kubernetes 清單

```bash
mcp-go-core k8s --name my-mcp-server --image myregistry/mcp-server:v0.1 --port 8080 -o k8s/
```

會生成 `k8s/my-mcp-server-deployment.yaml` 和 `k8s/my-mcp-server-service.yaml`。

### 驗證二進制

```bash
mcp-go-core verify --binary dist/mcp-go-core
```

---

## 使用方法

### CLI 命令

| 命令 | 說明 |
|---|---|
| `mcp-go-core init` | 初始化新 MCP 專案（`--name`、`--profile`） |
| `mcp-go-core build` | 建置 MCP 伺服器二進制（`--output`、`--profile`） |
| `mcp-go-core run` | 執行 MCP 伺服器（`--transport`、`--addr`、`--metrics`、`--tracing`、`--oauth`） |
| `mcp-go-core generate` | 生成 MCP 原始碼（`--dry-run`） |
| `mcp-go-core verify` | 驗證 MCP 伺服器二進制（`--binary`） |
| `mcp-go-core k8s` | 生成 Kubernetes 清單（`--name`、`--image`、`--port`、`--output`） |
| `mcp-go-core version` | 顯示版本資訊（`-V`） |

#### `run` 旗標

| 旗標 | 預設值 | 說明 |
|---|---|---|
| `--transport` | `stdio` | 傳輸類型：`stdio`、`http` 或 `sse` |
| `--addr` | `localhost:8080` | 監聽位址（適用 http/sse） |
| `--metrics` | `false` | 啟用 Prometheus 監控端點 |
| `--tracing` | `false` | 啟用 OpenTelemetry 追蹤 |
| `--oauth` | `false` | 啟用 OAuth 2.1 驗證 |

### 程式化 API

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
    // 定義一個工具
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

    // 建置並執行伺服器
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

### 傳輸介面

所有傳輸實作統一介面：

```go
type Transport interface {
    Serve(ctx context.Context, handler Handler) error
    Close(ctx context.Context) error
}

type Handler func(ctx context.Context, msg json.RawMessage) (any, error)
```

會話管理透過 `SessionManager` 提供：

```go
sm := transport.NewSessionManager()
id := sm.RegisterSession()  // NewSessionID() 生成隨機 ID
done := sm.GetSession(id)   // 返回完成 channel
sm.UnregisterSession(id)    // 關閉並移除會話
sm.CloseAll()               // 關閉所有會話
```

### 伺服器建造者

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

### 安全功能

**API Key：**

```go
auth := apikey.NewAuthenticator(map[string]apikey.Identity{
    "secret-key": {Principal: "user1", Scopes: []string{"read"}},
})
identity, err := auth.Authenticate(ctx, apikey.HTTPRequest{r})
```

**JWT：**

```go
auth := jwt.NewAuthenticator("hmac-secret", "my-issuer")
identity, err := auth.Authenticate(ctx, jwt.HTTPRequest{r})
```

**OAuth 2.1 + PKCE：**

```go
auth := oauth.NewAuthenticator(
    "https://auth.example.com",
    "mcp-client",
    []string{"read", "write"},
)
pkce, _ := oauth.GeneratePKCE()  // RFC 7636
identity, err := auth.Authenticate(ctx, oauth.HTTPRequest{r})
```

### 儲存

**外部儲存（Redis/PostgreSQL）：**

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

// 所有儲存均實作 Store 介面
var s external.Store = store
```

### 執行期（任務與會話）

```go
// 任務管理與取消
tm := task.NewManager()
defer tm.Close()

t := tm.Create("my-task", func(ctx context.Context) (task.Result, error) {
    return task.Result{Data: []byte("done")}, nil
})

status, _ := tm.Status(t.ID)  // 檢查狀態
err := tm.Cancel(t.ID)        // 取消仍在執行的任務
```

```go
// 會話管理與生命週期
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

## API 參考

### core/protocol

| 型別 | 說明 |
|---|---|
| `Message` | JSON-RPC 2.0 訊息（請求或回應） |
| `Request` | 進來的 JSON-RPC 請求 |
| `Response` | JSON-RPC 回應 |
| `Error` | JSON-RPC 錯誤（code、message） |
| `InitializeRequest` | MCP initialize 請求參數 |
| `InitializeResponse` | MCP initialize 回應 |
| `ServerCapabilities` | 伺服器能力宣告 |
| `ClientCapabilities` | 客戶端能力宣告 |
| `JSONRPCMessage` | JSON-RPC 訊息聯合型別 |
| `JSONRPCVersion` | `"2.0"` 常數 |

### core/tool

| 型別/介面 | 說明 |
|---|---|
| `Tool` | 介面：`Name()`、`Description()`、`InputSchema()`、`Handler()` |
| `BaseTool` | 預設實作 |
| `NewTool(name, desc, schema, handler)` | 建構函式 |
| `Schema` | `map[string]any`（JSON Schema） |
| `ToolHandler` | `func(ctx, *Request) (*Response, error)` |

### core/resource

| 型別/介面 | 說明 |
|---|---|
| `Resource` | 介面：`URI()`、`Name()`、`Description()`、`Read(ctx, req)` |
| `BaseResource` | 預設實作 |
| `NewResource(uri, name, desc, readFn)` | 建構函式 |

### core/prompt

| 型別/介面 | 說明 |
|---|---|
| `Prompt` | 介面：`Name()`、`Description()`、`Get(ctx, req)` |
| `BasePrompt` | 預設實作 |
| `NewPrompt(name, desc, getFn)` | 建構函式 |
| `PromptRequest` | 請求參數 |
| `PromptResponse` | 回傳含有 messages |

### core/mcperror

| 常數/型別 | 值/說明 |
|---|---|
| `CodeProtocol` | `"protocol"` |
| `CodeValidation` | `"validation"` |
| `CodeAuth` | `"auth"` |
| `CodeAuthorization` | `"authorization"` |
| `CodeTransport` | `"transport"` |
| `CodeTool` | `"tool"` |
| `CodeInternal` | `"internal"` |
| `CodeTimeout` | `"timeout"` |
| `CodeCancellation` | `"cancellation"` |
| `ErrCodeParseError` | `-32700` |
| `ErrCodeInvalidRequest` | `-32600` |
| `ErrCodeMethodNotFound` | `-32601` |
| `ErrCodeInvalidParams` | `-32602` |
| `ErrCodeInternalError` | `-32603` |
| `NewError(code int, msg string) *Error` | 建構函式 |
| `NewParseError()` 等 | 便捷建構函式 |
| `JSONRPCError` | 可 JSON 序列化的錯誤 |

### core/lifecycle

| 狀態 | 說明 |
|---|---|
| `Created` | 初始狀態 |
| `Configured` | 選項已套用 |
| `Initialized` | 就緒啟動 |
| `Started` | 開始執行 |
| `Running` | 運行中 |
| `ShuttingDown` | 正在優雅關閉 |
| `Shutdown` | 已完全停止 |

### modules/runtime/task

| 型別 | 說明 |
|---|---|
| `Status` | 任務狀態：`pending`、`running`、`completed`、`failed`、`cancelled` |
| `Result` | 任務結果：`Data []byte`、`Err error` |
| `Task` | 背景任務：ID、狀態、建立/完成時間戳 |
| `Manager` | 執行緒安全管理器：`Create`、`Cancel`、`Status`、`GetResult`、`WaitFor`、`RunningCount` |

### modules/runtime/session

| 型別 | 說明 |
|---|---|
| `Session` | 會話：ID、中繼資料、生命週期狀態 |
| `Manager` | 會話管理器：`Create`、`Get`、`Destroy`、`Count`、`ActiveSessions`、`Close`、`CloseAll` |

## 測試

```bash
# 所有測試
go test ./... -count=1

# 啟用競態偵測器
go test -race ./... -count=1

# 靜態分析
go vet ./...

# 效能測試
go test -bench=. -benchmem ./benchmarks/...
```

### 測試數量

- **337 total tests, 0 failures**

### testutil 套件

| 型別 | 說明 |
|---|---|
| `EchoServer` | 回音伺服器 |
| `MockTransport` | 模擬 `Transport` 實作，支援 `Intercept` |
| `TestSession` | 基於會話傳輸測試的輔助工具 |
| `AssertJSONEqual`、`AssertJSONError` 等 | 斷言輔助函式 |

---

## 建置

```bash
# 建置全部
make build

# 建置特定二進制
go build -o dist/mcp-go-core ./cmd/mcp-go-core

# 可重現建置（CGO 停用）
CGO_ENABLED=0 go build -ldflags="-s -w" -o dist/mcp-go-core ./cmd/mcp-go-core
```


### Docker 建置

```bash
# 建置 Docker 映像
make docker-build

# 使用 docker-compose 運行（HTTP + Prometheus + Grafana）
make docker-run

# 推送到登錄（需設定 DOCKER_REGISTRY 環境變數）
make docker-push DOCKER_REGISTRY=ghcr.io/project

# 清理
make docker-clean
```

### 建置管道

```bash
mcp-go-core init --name my-server --profile production
mcp-go-core generate --dry-run
mcp-go-core build --output dist/ --profile production --verify
mcp-go-core verify --binary dist/mcp-go-core
```

階段順序：`config` → `analyze` → `resolve` → `lock` → `generate` → `compile` → `verify`。

---

## 部署

### Minimal 配置

僅 stdio 傳輸，無外部依賴：

```bash
mcp-go-core run --transport stdio
```

二進制僅含：`core`、`stdio` 傳輸。

### Production 配置

HTTP 傳輸：

```bash
mcp-go-core run --transport http --addr 0.0.0.0:8080
```

### Secure 配置

HTTP + JWT：

```bash
mcp-go-core run --transport http --addr 0.0.0.0:8080
```

### Observability 配置

HTTP + 監控 + 追蹤：

```bash
mcp-go-core run --transport http --addr 0.0.0.0:8080 --metrics --tracing
```

[NEEDS VERIFICATION: 容器化部署工具鏈（Dockerfile、docker-compose.yml、CI 映像建置/推送）尚未實作。K8s 清單生成功能可透過 `mcp-go-core k8s` 使用，但缺少容器映像建置、登錄推送及部署自動化。待執行期部署工具完備後提供完整文件。]

---

## 安全

### 容器化部署

使用 Docker 建置並運行：

```bash
mcp-go-core build --output dist/mcp-go-core --profile production
docker build -t mcp-go-core:v0.1.0 .
docker run -p 8080:8080 mcp-go-core:v0.1.0 \
  mcp-go-core run --transport http --addr 0.0.0.0:8080 --metrics
```

使用 docker-compose 進行本地開發（HTTP + Prometheus + Grafana）：

```bash
docker compose --profile production up -d
```

| 服務 | URL | 配置檔案 |
|---|---|---|
| MCP 伺服器（HTTP） | http://localhost:8080 | production |
| Prometheus | http://localhost:9090 | production |
| Grafana | http://localhost:3000 | production |

### 安全驗證

| 場景 | 測試案例 |
|---|---|
| API Key | 有效：PASS，無效：拒絕，缺少：拒絕 |
| JWT | 有效：PASS，過期：拒絕，無效簽章：拒絕，缺少：拒絕 |
| OAuth | PKCE 生成（RFC 7636）、Bearer Token 驗證、Token 內省 |

mTLS 模組存在為預留套件 (`modules/security/mtls`) — 完整實作延期至 v2。

---

## 效能

### 效能測試

| 效能測試 | 說明 |
|---|---|
| `BenchmarkToolDispatch` | 單次工具調度延遲 |
| `BenchmarkToolDispatchInProcess` | In-process 調度（無傳輸層） |
| `BenchmarkToolDispatchThroughput` | 吞吐量測量 |
| `BenchmarkToolDispatchP50P99` | P50 和 P99 延遲 |
| `BenchmarkStartup` | 進程啟動到就緒 |
| `BenchmarkStartupMemory` | 啟動時記憶體使用量 |

### 效能目標

| 計量 | 目標 |
|---|---|
| 調度 P50 | < 10 µs |
| 調度 P99 | < 100 µs |
| 吞吐量 | > 100k req/s |
| 啟動時間 | < 50 ms |
| 最小 RSS | < 20 MB |
| 生產 RSS | < 30 MB |

---

## 貢獻

1. 確保所有測試通過：`go test ./...`
2. 執行競態偵測器：`go test -race ./...`
3. 執行靜態分析：`go vet ./...`
4. 維持核心隔離：`core/` 不得引入 `modules/` 或 `integration-*`
5. 模組隔離：傳輸層不得互相引入

---

## 授權

Apache 2.0
