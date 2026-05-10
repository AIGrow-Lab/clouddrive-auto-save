# OpenList 手动扫描功能设计文档

## 概述

在系统设置模块新增 OpenList 扫描功能，支持手动触发和转存任务完成后自动触发 OpenList 的文件扫描，实现快速生成本地 strm 文件。

## 背景

用户使用 OpenList 管理网盘文件并生成 strm 文件。当前转存任务完成后，需要手动登录 OpenList 触发扫描，流程繁琐。本功能将扫描触发集成到转存系统中，实现自动化。

## 技术方案

### 架构设计

```
┌─────────────────────────────────────────────────────────────┐
│                      前端 (Settings.vue)                     │
│  ┌─────────────────────┐  ┌─────────────────────────────┐  │
│  │  OpenList 配置卡片    │  │  手动扫描按钮                 │  │
│  └──────────┬──────────┘  └──────────────┬──────────────┘  │
└─────────────┼────────────────────────────┼─────────────────┘
              │                            │
              ▼                            ▼
┌─────────────────────────────────────────────────────────────┐
│                    API 层 (router.go)                        │
│  GET/POST /settings/global    POST /openlist/scan            │
└─────────────────────┬──────────────────────┬────────────────┘
                      │                      │
                      ▼                      ▼
┌─────────────────────────────────────────────────────────────┐
│               OpenList 模块 (internal/core/openlist/)        │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐  │
│  │   Client      │  │   Scanner    │  │   Config         │  │
│  │  (HTTP 调用)   │  │  (触发管理)   │  │  (配置读取)       │  │
│  └──────────────┘  └──────────────┘  └──────────────────┘  │
└─────────────────────────────────────────────────────────────┘
              ▲
              │ 任务完成回调
┌─────────────┴───────────────────────────────────────────────┐
│                    Worker (worker.go)                        │
│  任务完成 → 检查是否有新内容 → 调用 Scanner                    │
└─────────────────────────────────────────────────────────────┘
```

### 数据模型

复用现有 `Setting` 表的 key-value 存储：

| Key | 说明 | 示例值 |
|-----|------|--------|
| `openlist_enabled` | 是否启用 OpenList 扫描 | `true` / `false` |
| `openlist_api_url` | OpenList API 地址 | `http://127.0.0.1:23541` |
| `openlist_api_token` | 认证 Token | `openlist-913f51c8-...` |

配置通过现有 `GET/POST /settings/global` 接口读写。Token 在前端显示时做脱敏处理（显示前 8 位 + `***`）。

### 模块设计

#### 1. OpenList 客户端 (`internal/core/openlist/client.go`)

```go
type Client struct {
    baseURL    string
    token      string
    httpClient *http.Client
}

func NewClient(baseURL, token string) *Client
func (c *Client) StartScan(ctx context.Context) error
```

职责：
- 封装 HTTP POST 请求到 `/api/admin/scan/start`
- 设置 `Authorization` header
- 处理超时（默认 10 秒）和网络错误
- 不处理业务逻辑，仅负责 API 通信

错误处理：
- 网络超时 → 返回超时错误
- HTTP 非 2xx → 返回状态码和响应体错误
- 连接拒绝 → 返回连接失败错误

#### 2. 扫描管理器 (`internal/core/openlist/scanner.go`)

```go
type Scanner struct {
    mu           sync.Mutex
    client       *Client
    pendingBatch int        // 待完成的批量任务数
    scanTimer    *time.Timer // 延迟合并定时器
}

func NewScanner() *Scanner
func (s *Scanner) ReloadConfig() error
func (s *Scanner) ScanNow(ctx context.Context) error
func (s *Scanner) OnBatchStart(count int)
func (s *Scanner) OnTaskComplete(hasNewContent bool)
```

触发逻辑：

| 场景 | 行为 |
|------|------|
| 手动触发 | 立即调用 `StartScan()` |
| 单任务完成（有新内容） | 延迟 3 秒触发（合并短时间内的连续完成） |
| 批量任务 | 计数器递减，最后一个任务完成后延迟 3 秒触发 |

延迟 3 秒的原因：
- 避免短时间内多次扫描（如连续完成 5 个任务触发 5 次扫描）
- 批量场景下自动合并为一次扫描
- 3 秒足够等待同批次的其他任务完成

防重复机制：
- 通过 `sync.Mutex` 保证并发安全
- 如果扫描正在进行中，新请求被跳过（记录日志）

#### 3. Worker 集成 (`internal/core/worker/worker.go`)

单任务执行流程：
```
任务完成 → 检查是否有新转存文件 → 有则调用 scanner.OnTaskComplete(true)
```

批量任务执行流程：
```
开始批量 → scanner.OnBatchStart(taskCount)
  → 任务1完成 → scanner.OnTaskComplete(true/false)
  → 任务2完成 → scanner.OnTaskComplete(true/false)
  → ...
  → 最后一个完成 → 触发扫描
```

判断"有新内容"的依据：
- 复用现有的去重逻辑：如果任务执行后有文件被实际保存（非重复跳过），则视为有新内容
- Worker 已有保存成功的文件计数，直接复用

### API 设计

新增路由：

| 方法 | 路径 | 说明 |
|------|------|------|
| `POST` | `/api/openlist/scan` | 手动触发扫描 |

请求示例：
```bash
curl -X POST http://localhost:8080/api/openlist/scan
```

响应示例：
```json
{
  "message": "扫描已触发"
}
```

### 前端设计

Settings.vue 新增 OpenList 配置卡片：

```
┌─────────────────────────────────────────┐
│  OpenList 扫描配置                        │
├─────────────────────────────────────────┤
│  启用自动扫描    [开关]                    │
│                                         │
│  API 地址       [http://127.0.0.1:23541] │
│  API Token      [openlist-913f5***]      │
│                                         │
│  [手动扫描]                               │
└─────────────────────────────────────────┘
```

交互逻辑：
- 点击"手动扫描" → 发送请求 → 收到响应后显示 Toast 提示（成功/失败）
- 按钮不追踪状态，纯触发式操作
- 自动触发静默执行，失败时写入日志（通过 SSE 在 Dashboard 展示）

## 文件变更清单

| 文件 | 操作 | 说明 |
|------|------|------|
| `internal/core/openlist/client.go` | 新增 | OpenList HTTP 客户端 |
| `internal/core/openlist/scanner.go` | 新增 | 扫描管理器 |
| `internal/core/openlist/config.go` | 新增 | 配置读取逻辑 |
| `internal/api/router.go` | 修改 | 新增 `/api/openlist/scan` 路由 |
| `internal/core/worker/worker.go` | 修改 | 集成扫描触发逻辑 |
| `web/src/views/Settings.vue` | 修改 | 新增 OpenList 配置卡片 |
| `web/src/api/task.ts` | 修改 | 新增 `triggerOpenListScan` API |

## 测试策略

1. 单元测试：
   - Client 的 HTTP 调用和错误处理
   - Scanner 的触发逻辑和延迟合并

2. 集成测试：
   - 手动扫描 API 端点
   - 任务完成后自动触发

3. E2E 测试：
   - 设置页面配置和手动扫描交互

## 风险与限制

1. OpenList API 可用性：如果 OpenList 服务不可用，扫描会失败但不影响转存任务
2. 网络延迟：如果 OpenList 部署在远程服务器，可能有网络延迟
3. 并发扫描：OpenList 可能不支持并发扫描，需要通过 Scanner 的防重复机制避免
