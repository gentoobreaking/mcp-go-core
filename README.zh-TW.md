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
- **執行期可運作性:** 健康端點、功能旗標、速率限制與動態配置可於執行期使用

---

## 功能列表

### 核心功能（始終包含）

| 功能 | 說明 |
|---|---|
| MCP 協定型別 | JSON-RPC 2.0 訊息結構、錯誤碼 |
| 工具註冊 | 透過 `core/tool` 訂冊與調度工具 |
| 資源註冊 | 透過 `core/resource` 註冊與調度資源 |
| Prompt 註冊 | 透過 `core/prompt` 註冊與調度 Prompt |
| 路由器 | 基於方法的派發（`tools/call`、`resources/read` 等），支援進階 MCP 方法（ping、complete、roots、tasks、elicitation、subscriptions、progress、message） |
| 伺服器建造者 | 流暢 API：`WithName`、`WithTool`、`WithResource`、`WithPrompt`、`WithTransport`、`WithMiddleware`、`WithHealth`、`WithFlags`、`WithRateLimiter`、`WithConfig`、`Build` |
| 生命週期管理 | 狀態機：Created → Configured → Initialized → Started → Running → ShuttingDown → Shutdown |
| 結構化錯誤 | 具備 `mcperror` 套件的 JSON-RPC 2.0 錯誤碼 |

### 傳輸層

| 功能 | 套件 |
|---|---|
| stdio 傳輸 | `modules/transport/stdio` |
| Streamable HTTP 傳輸 | `modules/transport/http`（支援 `WithHealthRoutes` 健康端點） |
| SSE 傳輸（含會話） | `modules/transport/sse`（支援 `WithHealthRoutes` 健康端點） |
| 傳輸介面 | `modules/transport`（統一 `Transport` 介面：`Serve` + `Close`） |
| 會話管理 | `modules/transport.SessionManager` 與 `NewSessionID()` |

### 功能旗標 & 速率限制

| 功能 | 套件 |
|---|---|
| 功能旗標 | `core/feature`（執行期 `Flags`、`IsDisabled`、`Snapshot`） |
| 功能連線中介軟體 | `core/middleware/featurewire`（方法→旗標門控、`HealthStatus`） |
| 速率限制 | `core/middleware/ratelimit`（token-bucket `Manager`、`NewLimiter`、`Allow`） |
| 健康端點 | `core/server`（`HealthHandler()`、`WithHealth(true)`） — 連接 HTTP/SSE 傳輸 |
| 動態配置 | `core/config`（基於 `fsnotify` 的熱重載 YAML、原子 `Load()`、`Health()`） |

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
| 檔案系統儲存 | `modules/storage/filesystem`（路徑遍歏保護、context 感知） |
| Redis 儲存 | `modules/storage/external`（連線池、TTL、SCAN） |
| PostgreSQL 儲存 | `modules/storage/external`（upsert、前綴掃描、資料表支援） |

### 執行期

| 功能 | 套件 |
|---|---|
| 任務管理 | `core/router`（`TaskResult`、`TaskStatus`、`RegisterTask`、`tasks/get`、`tasks/list`、`tasks/cancel`） |
| 會話管理 | `modules/runtime/session`（Session、Manager、生命週期整合） |
| 資源通知 | `core/router`（`DeleteResource`、`NotifyResourceDeleted`、`NotifyResourceUpdate`、`Subscribe`、`UnsubscribeClient`） |
| 引導式提問 | `core/router`（`ElicitationCreateParams`、`SetElicitationHandler`、`ResolveElicitation`） |
| 根目錄管理 | `core/router`（`Root`、`ListRootsResult`、`SetRoots`、`SetRootsHandler`） |

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
│              傳輸層                               │
│        (stdio, HTTP, SSE + SessionManager)        │
├──────────────────────────────────────────────────┤
│                       核心                          │
│   protocol │ server │ router │ tool │ resource    │
│   │ prompt │ lifecycle │ mcperror │ middleware     │
├──────────────────────────────────────────────────┤
│                  中介軟體                           │
│  (Logging, Recovery, FeatureWire, RateLimit)      │
├──────────────────────────────────────────────────┤
│        選用模組（core 無向上反向依賴）             │
│  Security: api_key, jwt, oauth                     │
│  Observability: metrics, tracing                   │
│  Storage: memory, filesystem, external             │
│  Runtime: task, session                            │
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
│   ├── server/              # Server + Builder API（包含健康端點）
│   ├── router/              # Tool/Resource/Prompt 調度（進階 MCP 方法）
│   ├── tool/                # Tool 介面 + BaseTool
│   ├── resource/            # Resource 介面 + BaseResource
│   ├── prompt/              # Prompt 介面 + BasePrompt
│   ├── lifecycle/           # 生命週期狀態機
│   ├── mcperror/            # 結構化錯誤碼
│   ├── feature/             # 執行期功能旗標
│   └── middleware/          # 中介軟體鏈（Logging、Recovery）
│       ├── featurewire/     # 功能旗標門控中介軟體
│       └── ratelimit/       # token-bucket 速率限制
│   ├── config/              # 動態配置（熱重載 YAML）
├── modules/                 # 選用實作
│   ├── transport/           # 傳輸介面 + SessionManager
│   │   ├── stdio/           # stdio 傳輸
│   │   ├── http/            # Streamable HTTP 傳輸（健康路由）
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
| `golang.org/x/oauth2` | v0.36.0 | OAuth PKCE 支援 |
| `github.com/redis/go-redis/v9` | v9.22.0 | Redis 儲存後端 |
| `github.com/lib/pq` | v1.12.3 | PostgreSQL 儲存後端 |
| `k8s.io/api` | v0.37.0 | Kubernetes 清單型別 |
| `golang.org/x/time` | v0.11.0 | token-bucket 速率限制 |
| `github.com/fsnotify/fsnotify` | v1.9.x | 配置檔熱重載 |
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
    "github.com/project/mcp-go-core/core/protocol"
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

### 健康端點

透過 `WithHealth(true)` 啟用 HTTP/SSE 傳輸上的健康檢查端點：

```go
srv, err := server.NewBuilder().
    WithName("my-server").
    WithTransport(http.New(addr)).
    WithHealth(true).         // 在傳輸層啟用健康路由
    WithFlags(feature.NewFlags(initialFlags)).
    WithRateLimiter(ratelimit.NewManager()).
    WithConfig(config.NewConfig("config.yaml")).
    Build()
```

傳輸層會註冊一個健康多路路由（mux），優先處理 `/health*` 路徑，其他路徑則回落到主要 MCP 處理器。以下路由可用：

| 路由 | 說明 |
|---|---|
| `GET /health` | 基本存活檢查（`{"status":"ok"}`） |
| `GET /health/features` | 所有功能旗標狀態（需 `WithFlags`） |
| `GET /health/features/<name>` | 單一旗標狀態 |
| `GET /health/rate-limits` | 速率限制狀態（需 `WithRateLimiter`） |
| `GET /health/config` | 配置熱重載中繼資料（需 `WithConfig`） |

也可以直接取得健康處理器：

```go
handler := srv.HealthHandler()
// 若未呼叫 WithHealth(true)，handler 為 nil
```

### 功能旗標

執行期功能旗標允許在無需重啟的情況下啟用/停用 MCP 方法：

```go
import "github.com/project/mcp-go-core/core/feature"

flags := feature.NewFlags(map[string]bool{
    "tools/call:advanced":  true,   // 啟用進階工具
    "resources/read:secret": false,  // 閉門敏感資源
})

srv, _ := server.NewBuilder().
    WithName("my-server").
    WithFlags(flags).
    WithMiddleware(featurewire.Middleware(flags, featurewire.DefaultFlagMapper)).
    Build()

// 執行期切換 — 下一個請求即見新狀態
flags.Set("tools/call:advanced", false)

// 檢查狀態
snap := flags.Snapshot()  // map[string]feature.Flag
enabled := flags.EnabledList() // []string 啟用的旗標名稱
```

`featurewire` 中介軟體透過 `FlagMapper` 函式將 MCP 方法名稱映射到旗標名稱。預設映射器會轉換例如 `tools/call:advanced` → `advanced_tools`。

### 速率限制

基於 `golang.org/x/time/rate` 的 token-bucket 速率限制：

```go
import "github.com/project/mcp-go-core/core/middleware/ratelimit"

lim := ratelimit.NewManager()
lim.Allow("tools/call")  // 允許則返回 nil，超過則返回 ratelimit.ErrRateLimit

// 提供給健康端點
for _, st := range lim.Status() {
    fmt.Println(st.Name, st.Limit, st.Burst)
}
```

| 功能 | 套件 | 受限方法 |
|---|---|---|
| `DefaultLimits` | `core/middleware/ratelimit` | `tools/call`、`tools/list`、`prompts/get`、`prompts/list`、`resources/read`、`resources/list` |

### 動態配置

基於 `fsnotify` 的熱重載 YAML 配置，支援原子交換：

```go
import "github.com/project/mcp-go-core/core/config"

cfg := config.NewConfig("config.yaml")
cfg.Load() // 讀取 YAML，原子交換配置

// 監聽文件變更
watcher, _ := config.NewWatcher(cfg, func(format string, args ...any) {
    log.Printf(format, args...)  // 記錄重載事件
})
defer watcher.Close()

// 健康端點暴露配置中繼資料
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

### 資源通知與訂閱

路由器追蹤每個客戶端的資源訂閱，並發送變更通知：

```go
// 訂閱資源變更
router.Subscribe(uri, "client-1")

// 檢查訂閱狀態
router.IsSubscribed(uri)  // 若有任何客戶端訂閱則為 true

// 通知訂閱者資源已更新
router.NotifyResourceUpdate(uri, "update")

// 刪除資源並在清除訂閱前通知訂閱者
router.DeleteResource(uri)

// 取消特定客戶端的訂閱
router.UnsubscribeClient(uri, "client-1")

// 移除 URI 的所有訂閱
router.Unsubscribe(uri)
```

訂閱映射表為 `map[string]map[string]bool`（URI → clientIDs）。客戶端 ID 透過 `clientIDFromContext()` 從 context 中提取，預設為 `"default"`。

通知透過 `NotificationSender` 發送 — 這是一個回調函數，從傳輸層連接：

```go
router.SetNotificationSender(func(method string, params any) error {
    // 序列化並透過傳輸層發送給已連接的客戶端
    return transport.Send(method, params)
})
```

資源通知的協定型別：

| 型別 | 說明 |
|---|---|
| `ResourceUpdateNotification` | `notifications/resources/update` — 資源變更通知 |
| `ResourceUpdateParams` | 參數含 `URI` 和 `ChangeType` |
| `ResourceDeletedNotification` | `notifications/resources/deleted` — 資源刪除通知 |
| `ResourceDeleteParams` | 參數含被刪除資源的 `URI` |
| `SubscribeParams` | `resources/subscribe` 請求參數（`URI`） |
| `Subscription` | 主動訂閱（`URI`、`SubscribedAt`） |
| `UnsubscribeParams` | `resources/unsubscribe` 請求參數（`URI`） |

### 進階 MCP 方法

路由器實作完整的 MCP 協定方法集：

| 方法 | 方向 | 說明 |
|---|---|---|
| `ping` | C→S | 存活檢查，返回 `PingResult{"pong"}` |
| `complete/arg` / `complete/prompt` | C→S | 引數/Prompt 補全 |
| `roots/list` | C→S | 列出客戶端提供的根目錄 |
| `notifications/roots/list_changed` | C→S | 客戶端通知根目錄已更改 |
| `prompts/create` | C→S | 執行期動態註冊 Prompt |
| `notifications/prompts/list_changed` | C→S | 客戶端通知 Prompt 清單已更改 |
| `notifications/resources/list_changed` | C→S | 客戶端通知資源清單已更改 |
| `notifications/tools/list_changed` | C→S | 客戶端通知工具清單已更改 |
| `resources/templates/list` | C→S | 列出資源 URI 模板 |
| `resources/subscribe` | C→S | 訂閱資源更新通知 |
| `resources/unsubscribe` | C→S | 移除訂閱 |
| `notifications/progress` | S→C / C→S | 進度更新（`ProgressToken`） |
| `notifications/message` | S→C / C→S | 客戶端發送記錄訊息 |
| `notifications/resources/update` | S→C | 通知訂閱者資源已變更 |
| `notifications/resources/deleted` | S→C | 通知訂閱者資源已刪除 |
| `elicitation/create` | S→C | 伺服器向客戶端請求輸入 |
| `notifications/elicitation/complete` | C→S | 客戶端回應引導式提問 |
| `tasks/get` | C→S | 依 ID 取得任務 |
| `tasks/list` | C→S | 列出所有任務 |
| `tasks/cancel` | C→S | 依 ID 取消任務 |
| `server/discover` | C→S | 探索伺服器能力與協定版本 |
| `subscriptions/listen` | C→S | 註冊訂閱監聽器 |

#### 工具與 Prompt 動態註冊

透過 `prompts/create` 在執行期動態註冊 Prompt：

```go
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

#### 任務管理

```go
// 在路由器上註冊任務
router.RegisterTask("task-1", protocol.TaskStatusRunning, nil)
router.RegisterTask("task-2", protocol.TaskStatusCompleted, "result-data")

// 透過 MCP 方法查詢：
//   tasks/list   → 返回含所有任務的 TaskListResult
//   tasks/get    → 依 ID 返回 TaskResult
//   tasks/cancel → 取消執行中的任務
```

#### 引導式提問

伺服器向客戶端請求資訊（例如缺少的參數）：

```go
router.SetElicitationHandler(func(params protocol.ElicitationCreateParams) error {
    // 傳輸層透過 elicitation/create 發送給客戶端
    // 客戶端透過 notifications/elicitation/complete 回應
    return nil
})

// 解析客戶端回應
result, ok := router.ResolveElicitation(elicitationID)
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
| `JSONRPCVersion` | `"2.0"` 常數 |
| `JSONRPCMessage` | JSON-RPC 訊息聯合型別 |
| `InitializeRequest` | MCP initialize 請求參數 |
| `InitializeResponse` | MCP initialize 回應 |
| `InitializeParams` | Initialize 請求參數 |
| `InitializeResult` | Initialize 回應資料 |
| `ServerCapabilities` | 伺服器能力宣告 |
| `ClientCapabilities` | 客戶端能力宣告 |
| `Implementation` | 客戶端/伺服器實作資訊 |

#### 通知（伺服器 → 客戶端）

| 型別 | 說明 |
|---|---|
| `Notification` | JSON-RPC 通知（無 ID） |
| `ResourceUpdateNotification` | `notifications/resources/update` — 資源變更通知 |
| `ResourceUpdateParams` | 參數含 `URI` 和 `ChangeType`（`"update"` 或 `"delete"`） |
| `ResourceDeletedNotification` | `notifications/resources/deleted` — 資源刪除通知 |
| `ResourceDeleteParams` | 參數含被刪除資源的 `URI` |
| `ToolListChangedNotification` | `notifications/tools/list_changed` |
| `PromptListChangedNotification` | `notifications/prompts/list_changed` |
| `LoggingMessage` | `logging/message` — 伺服器日誌通知 |
| `ProgressNotification` | `notifications/progress` — 進度更新 |
| `ProgressNotificationParams` | 參數含 `ProgressToken`、`Progress`、`Total` |
| `MessageNotification` | `notifications/message` — 伺服器記錄訊息 |
| `MessageNotificationParams` | 參數含 `Level`、`Logger`、`Data` |
| `CreatedNotification` | `notifications/resources/created` — 資源已建立通知 |
| `CreatedParams` | 參數含 `URI` 和 `Name` |

#### 通知（客戶端 → 伺服器）

| 型別 | 說明 |
|---|---|
| `ElicitationCompleteParams` | `notifications/elicitation/complete` — 客戶端回應引導式提問 |
| `RootsListChangedParams` | `notifications/roots/list_changed` — 客戶端根目錄已更改 |

#### 能力宣告

| 結構 | 欄位 |
|---|---|
| `PromptsCapability` | `ListAvailable: bool` |
| `ResourcesCapability` | `ListAvailable`, `Subscribe`, `Create` |
| `ToolsCapability` | `ListAvailable`, `Create`, `ListChanged` |
| `LoggingCapability` | `Log: bool` |
| `CompletionsCapability` | `Complete: bool` |
| `RootsCapability` | `ListAvailable: bool` |
| `SamplingCapability` | `CreateMessage: bool` |

#### 請求與結果

| 型別 | 說明 |
|---|---|
| `PromptListParams` / `PromptListResult` | 列出 Prompts |
| `ResourceListParams` / `ResourceListResult` | 列出資源 |
| `ToolListParams` / `ToolListResult` | 列出工具 |
| `SubscribeParams` | `resources/subscribe` 參數（`URI`） |
| `UnsubscribeParams` | `resources/unsubscribe` 參數（`URI`） |
| `SubscriptionListenParams` | `subscriptions/listen` 參數（`URI`） |
| `CompletionParams` / `CompleteResult` | 補全請求與結果 |
| `CreateMessageParams` / `CreateMessageResult` | `sampling/createMessage` |
| `ElicitationCreateParams` / `ElicitationResult` | `elicitation/create` |
| `TaskCancelParams` | `tasks/cancel` 參數（`ID`） |
| `TaskResult` | 任務結果（`ID`、`Status`、`Result`） |
| `TaskListResult` | 任務列表 |
| `ListRootsResult` | `roots/list` 結果（`Roots []Root`） |
| `Root` | 根目錄（`URI`、`Name`、`Description`） |
| `Subscription` | 主動訂閱（`URI`、`SubscribedAt`） |
| `ResourceTemplate` | 資源 URI 模板 |
| `ResourceTemplateListResult` | 資源模板列表 |
| `Prompt` | Prompt 定義 |
| `Resource` | 資源參考 |
| `Tool` | 工具定義 |
| `Argument` | 參數定義 |
| `NotificationSender` | `func(method string, params any) error` — 發送通知給客戶端 |

#### 錯誤類型

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
| `NewParseError()` / `NewInvalidRequestError()` 等 | 便捷建構函式 |
| `JSONRPCError` | 可 JSON 序列化的錯誤 |

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

### core/feature

| 型別 | 說明 |
|---|---|
| `Flag` | `{Enabled bool}` — 功能旗標狀態 |
| `Flags` | 執行緒安全儲存：`Get`、`Set`、`IsDisabled`、`Snapshot`、`EnabledList` |
| `NewFlags(map[string]bool)` | 建構函式 |

### core/middleware/featurewire

| 型別 | 說明 |
|---|---|
| `FlagMapper` | `func(method string) string` — 映射 MCP 方法到旗標名稱 |
| `DefaultFlagMapper` | 預設映射器（例如 `tools/call:advanced` → `advanced_tools`） |
| `Middleware(flags, mapper)` | 基於旗標門控的方法中介軟體 |
| `HealthStatus(flags)` | 返回旗標狀態供健康端點使用 |
| `FlagStatus` | 可 JSON 序列化的旗標狀態（`Name`、`Enabled`） |

### core/middleware/ratelimit

| 型別 | 說明 |
|---|---|
| `Limiter` | token-bucket 限制器，含 `name`、`limit`、`burst` |
| `NewLimiter(name, rate, burst)` | 建構函式 |
| `Manager` | 執行緒安全管理器：`NewManager`、`Init`、`Allow`、`Status`、`AllowAll`、`RejectAll` |
| `Status` | `{Name, Limit, Burst, ...}` — 可 JSON 序列化 |
| `DefaultLimits` | 預設每方法限制映射表 |

### core/config

| 型別 | 說明 |
|---|---|
| `Config` | 熱重載配置：`Load`、`GetServer`、`GetFeatures`、`GetRateLimits`、`Health` |
| `ServerConfig` | 伺服器層級設定 |
| `LimitConfig` | 速率限制配置（`Rate`、`Burst`） |
| `LoggingConfig` | 日誌設定 |
| `HealthInfo` | 配置健康中繼資料 |
| `NewConfig(path)` | 從 YAML 路徑建立配置 |
| `NewWatcher(cfg, logger)` | 基於 fsnotify 的文件監視器 |

### core/router

| 型別/方法 | 說明 |
|---|---|
| `Router` | 中央派發：`RegisterTool`、`RegisterResource`、`RegisterPrompt`、`Dispatch` |
| `SamplingHandler` | `func(ctx, *CreateMessageParams) (*CreateMessageResult, error)` |
| `CreatedNotifier` | `func(uri, name string) error` — 資源建立回呼 |
| `PromptCreator` | `func(params PromptCreateParams) (Prompt, error)` — `prompts/create` 工廠 |
| `DeleteResource(uri)` | 刪除資源並通知訂閱者 |
| `NotifyResourceDeleted(uri)` | 發送 `notifications/resources/deleted` 給訂閱者 |
| `NotifyResourceUpdate(uri, changeType)` | 發送資源變更通知 |
| `Subscribe(uri, clientID)` | 逐客戶端訂閱 |
| `UnsubscribeClient(uri, clientID)` | 移除單一客戶端訂閱 |
| `Unsubscribe(uri)` | 移除 URI 的所有訂閱 |
| `IsSubscribed(uri)` | 檢查是否有訂閱者 |
| `SetNotificationSender(handler)` | 連接通知發送到傳輸層 |
| `SetSampler(h)` | 註冊 sampling 處理器 |
| `SetProgressHandler(h)` | 註冊 progress 回呼 |
| `SetMessageHandler(h)` | 註冊 message 回呼 |
| `SetElicitationHandler(h)` | 註冊 elicitation 回呼 |
| `SetPromptCreator(fn)` | 註冊 `prompts/create` 工廠 |
| `SetPromptListChangedHandler(h)` | 註冊 `prompts/list_changed` 回呼 |
| `SetResourceListChangedHandler(h)` | 註冊 `resources/list_changed` 回呼 |
| `SetToolsListChangedHandler(h)` | 註冊 `tools/list_changed` 回呼 |
| `SetRootsHandler(h)` | 註冊 `roots/list_changed` 回呼 |
| `SetRoots(roots)` | 設定客戶端提供的根目錄 |
| `RegisterTask(id, status, result)` | 註冊任務到登記表 |
| `ResolveElicitation(id)` | 依 ID 取得引導式提問結果 |
| `HandleError(err)` | 錯誤處理介面 |

### core/server

| 型別 | 說明 |
|---|---|
| `Server` | 具備生命週期、健康端點、門控功能的 MCP 伺服器 |
| `Builder` | 流暢式伺服器建造者 |
| `Option` | `func(*Server)` — 伺服器選項 |
| Option 函式 | `WithName`、`WithVersion`、`WithFlags`、`WithRateLimiter`、`WithHealth`、`WithConfig`、`WithTransport`、`WithMiddleware`、`WithTimeout` |
| `NewServer(opts...)` | 功能建構函式 |
| `NewBuilder()` | 建造者建構函式 |
| `HealthHandler()` | 返回健康端點 HTTP 處理器（若未啟用則為 nil） |
| `SendNotification(method, params)` | 廣播通知給已連接客戶端 |
| `AddTool(t)` | 建置後註冊工具 |
| `AddResource(r)` | 建置後註冊資源 |
| `AddPrompt(p)` | 建置後註冊 Prompt |
| `Run(ctx)` | 啟動伺服器（阻塞直到 ctx 取消） |
| `Shutdown(ctx)` | 優雅關閉 |

### core/middleware

| 型別 | 說明 |
|---|---|
| `Middleware` | `func(Handler) Handler` |
| `Handler` | `interface{ Dispatch(ctx, method, params) (any, error) }` |
| `HandlerFunc` | 適配器，讓普通函式實作 Handler |
| `Chain(mw...)` | 組合中介軟體鏈 |

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

---

## 測試

```bash
# 所有測試
go test ./... -count=1

# 啟用競態偵測器
go test -race ./... -count=1

# 特定套件
go test ./core/...
go test ./modules/...
go test ./internal/...

# 效能測試
go test -bench=. -benchmem ./benchmarks/...
```

### 測試數量

- **366 + 14 額外測試，0 失敗**
- `core/server/health_test.go`: 11 測試（健康端點）
- `core/router/router_test.go`: 3 測試（資源刪除與逐客戶端訂閱追踪）

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

| 服務 | URL | 配置 |
|---|---|---|
| MCP 伺服器（HTTP） | http://localhost:8080 | production |
| Prometheus | http://localhost:9090 | production |
| Grafana | http://localhost:3000 | production |

---

## 安全

### 設計原則

- Core 無需任何驗證要求 — 驗證為選用模組
- API Key 驗證透過 `modules/security/api_key`
- JWT 驗證透過 `modules/security/jwt` (HMAC-SHA256)
- OAuth 2.1 + PKCE 透過 `modules/security/oauth` (RFC 7636)

### 安全驗證

| 場景 | 測試案例 |
|---|---|
| API Key | 有效：PASS，無效：拒絕，缺少：拒絕 |
| JWT | 有效：PASS，過期：拒絕，無效簽章：拒絕，缺少：拒絕 |
| OAuth | PKCE 生成（RFC 7636）、Bearer Token 驗證、Token 內省 |

mTLS 模組存在為預留套件 (`modules/security/mtls`) — 完整實作延期至 v2。

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
