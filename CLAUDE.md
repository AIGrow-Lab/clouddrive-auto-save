# CLAUDE.md

本文件为 Claude Code (claude.ai/code) 在本仓库中工作提供指导。

## 项目概述

**统一云盘自动转存系统 (UCAS)** 是一个云盘自动化工具，整合了移动云盘 (139) 和夸克网盘 (Quark) 的转存能力。支持多任务并发转存、正则表达式文件过滤/重命名、双层 Cron 调度。后端 (Go/Gin) 将嵌入式 Vue 3 SPA 作为单一二进制文件提供服务。

## 常用命令

所有任务通过 `make` 管理（详见 `Makefile`）：

```bash
make dev-server     # 在 :8080 端口启动 Go 后端（DEBUG 模式）
make dev-web        # 在 :5173 端口启动 Vue 3 开发服务器（代理 /api 到 :8080）
make build          # 完整生产构建：前端编译 -> 嵌入 Go 二进制 -> bin/ucas
make build-web      # 仅构建前端（web/dist -> internal/api/dist）
make test           # 运行 Go 单元测试（带 -race 和覆盖率）
make lint           # gofmt 代码格式检查
make vet            # go vet ./...
make check          # 完整 CI 流水线：lint + vet + test
make e2e-setup      # 安装 Playwright + Chromium（仅首次需要）
make e2e-test       # 构建二进制，以 E2E 模式启动服务，运行 Playwright 测试
make docker-build   # 构建 Docker 镜像
make docker-up      # 通过 docker-compose 启动
make clean          # 清理 bin/、web/dist/、coverage.out
```

运行单个 Go 测试：`go test -run TestName ./internal/path/...`

## 架构

### 后端 (Go 1.25)

- **入口点**：`cmd/server/main.go` — 初始化日志 (slog)、SQLite (GORM)、工作池 (3 个 goroutine)、Cron 调度器、Gin HTTP 服务器
- **`internal/core/drive.go`** — `CloudDrive` 接口，所有驱动实现该接口。工厂模式：`RegisterDriver(platform, factory)` / `GetDriver(account)`
- **驱动注册**：通过 `init()` 副作用导入在 `internal/api/router.go` 中注册：
  - `internal/core/cloud139/client.go` — 移动云盘
  - `internal/core/quark/client.go` — 夸克网盘
- **工作池** (`internal/core/worker/`) — 带缓冲的任务队列 (容量 100)，每个 worker 执行完整流水线：解析分享链接 → 去重检查 → 正则过滤 → 保存链接 → 重命名文件 → Bark 通知
- **调度器** (`internal/core/scheduler/`) — 封装 robfig/cron，支持秒级精度。"global" 模式共享一个 cron 触发所有全局任务；"custom" 模式为每个任务独立 cron。带有 `[Fatal]` 消息的任务会被自动跳过
- **重命名器** (`internal/core/renamer/`) — 支持魔法变量 `{TASKNAME}`、`{OLDNAME}`、`{CHINESE}`、`{DATE}`、`{YEAR}`、`{EXT}`，正则捕获组 `${1}`，以及 Go `text/template` 表达式
- **SSE/事件系统** (`internal/utils/`) — `Broadcaster` 发布/订阅系统，向所有 SSE 客户端广播实时日志和 `[EVENT:task_update|task_delete|stats_update]` 结构化 JSON 事件。`DashboardLogger` 双写 slog 输出到控制台 + SSE

### 前端 (Vue 3 + Vite)

4 个页面位于 `web/src/views/`：Dashboard（实时统计 + SSE 日志）、Accounts（139/Quark 账号管理）、Tasks（任务 CRUD + 预览 + 执行）、Settings（全局调度 + Bark 配置）

### 构建标签分离

- `internal/api/fs.go`（标签 `!embed`）：从磁盘 `web/dist/` 目录提供静态文件（开发环境）
- `internal/api/fs_embed.go`（标签 `embed`）：从 `embed.FS` 提供静态文件（生产二进制）

### E2E 测试

- `E2E_TEST_MODE=true` 激活 `internal/core/mock_http.go`，替换 `http.DefaultTransport` 拦截所有云盘 API 调用并返回预设响应
- Playwright 测试用例位于 `e2e/tests/`，覆盖账号、仪表盘、任务和设置模块
- Playwright 选择器：当多个按钮的可访问名称有包含关系时（如"选择目录" vs "浏览分享内容并选择目录"），必须使用 `getByRole('button', { name: '...', exact: true })` 避免严格模式冲突

### 前端注意事项

- Element Plus `el-input` 的 `#append` 插槽中放置多个按钮时，需要显式 CSS：`:deep(.el-input-group__append) { display: flex; align-items: center; }`，按钮需设置 `margin-left: 0`
- lucide-vue-next 图标居中：按钮需设置 `display: inline-flex; align-items: center; justify-content: center`
- 139 平台分享链接 URL 不包含目录信息（不像夸克可通过 URL 中的 pdirFID 区分），子目录需通过 `share_parent_id` 字段单独存储

## 核心约定

- **语言**：所有注释、文档、提交信息和 UI 文本必须使用**中文**
- **提交规范**：严格遵循 Angular Conventional Commits — `feat(scope): ...`、`fix(scope): ...`、`docs(scope): ...`，重点阐述 "原因 (Why)" 和 "改动 (What)"
- **纯 Go 架构**：使用 `glebarez/sqlite`（无 CGO 依赖）以确保跨平台交叉编译，无需 C 编译器
- **错误处理**：不可吞噬错误。严重异常使用 `[Fatal]` 级别日志，通过 SSE 同步至前端 UI 展示
- **环境变量**：`LOG_LEVEL`（DEBUG/INFO/WARN/ERROR）、`DB_PATH`（默认 `data.db`）、`LISTEN_ADDR`（默认 `0.0.0.0:8080`）

## 数据库模型 (`internal/db/db.go`)

- **Account**：平台 (139/quark)、昵称、凭证、状态、容量
- **Task**：关联账号、分享链接、提取码、保存路径、正则表达式、Cron、状态/进度
- **CommonFolder**：每个账号的收藏文件夹路径
- **Setting**：全局配置的键值存储（调度、Bark 通知）

## API 路由

所有路由以 `/api` 为前缀（定义在 `internal/api/router.go`）：
- 账号管理：增删改查 + 校验 + 文件夹
- 任务管理：增删改查 + 执行 + 全部执行 + 预览 + 解析分享 + 忽略
- 仪表盘：统计信息 + SSE 日志流 + 历史日志 + 清空日志
- 设置：调度配置 + 全局设置 + 测试 Bark
