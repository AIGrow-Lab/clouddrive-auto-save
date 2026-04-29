# 统一云盘自动转存系统 (UCAS) 项目指南

## 项目概述
**统一云盘自动转存系统 (UCAS)** 是一个基于 Go 语言和 Vue 3 构建的高性能、低资源占用的云盘自动化工具。它旨在整合移动云盘(139)与夸克网盘(Quark)的核心转存能力，支持多任务并发转存、智能整理与去重，并提供现代化的实时监控后台。

- **后端技术栈**: Go 1.25+, Gin, GORM, Glebarez SQLite (无 CGO 依赖的纯 Go 实现), Cron 任务调度。
- **前端技术栈**: Vue 3.5+, Vite, Element Plus, Pinia, Vue Router, Server-Sent Events (SSE) 实时通信。
- **基础设施**: Docker 容器化部署、GitHub Actions 自动化构建与测试、GoReleaser 多架构编译发布。

## 目录结构
- `cmd/server/main.go`: 后端服务入口点。
- `internal/`: 核心后端业务代码（包括 API 路由、云盘处理引擎、数据库操作等）。
- `web/`: 独立的 Vue 3 前端工程代码。
- `e2e/`: 基于 Playwright 的端到端自动化测试套件。
- `docs/`: 详细的 API 设计与技术规范文档。
- `bin/`: 默认的构建二进制产物输出目录。
- `data/`: 默认的 SQLite 数据库和持久化配置文件存放目录（容器化时通过 Volume 挂载）。
- `conductor/`: 项目规划与任务管理工作区（Plan Mode 专用目录）。

## 构建与运行指南

本项目统一使用 `Makefile` 进行任务管理：

### 开发环境
- **启动后端服务 (DEBUG 模式)**: 
  ```bash
  make dev-server
  ```
  将在 `8080` 端口启动 Go 后端。
- **启动前端服务**:
  ```bash
  make dev-web
  ```
  将在 `5173` 端口启动 Vue 3 热更新开发服务器。

### 生产构建与打包
- **完整构建 (编译前端并内嵌到后端二进制)**:
  ```bash
  make build
  ```
  编译完成后，独立的二进制文件将生成在 `bin/ucas`。
- **Docker 容器化构建与运行**:
  ```bash
  make docker-build
  make docker-up
  ```

### 测试与质量验证
- **代码质量与单元测试**:
  ```bash
  make check
  ```
  （自动化执行 `go fmt`, `go vet`, 及 `go test -race` 流程）。
- **端到端测试**:
  ```bash
  make e2e-setup  # 仅首次运行，用于安装依赖和 Playwright 浏览器
  make e2e-test
  ```

## 开发规范与约定
1. **纯 Go 架构**: 数据库层使用 `glebarez/sqlite` 替代了传统的 CGO SQLite 驱动，以确保应用可以在不依赖本地 C 编译器的情况下进行多平台交叉编译。
2. **前后端集成**: 生产环境构建时，前端代码会被编译至 `web/dist`，随后移入 `internal/api/dist` 并通过 Go 语言的 `embed` 特性静态打包进单一的可执行文件中。
3. **中文语境强制要求**: 根据系统全局规范，所有的功能解释、需求沟通、代码注释、相关文档维护以及 Git 提交记录均必须严格使用**中文**。
4. **Git 提交规范**: 严格遵循 Angular 的 Conventional Commits 规范（如 `feat(api): ...`, `fix(core): ...`），提交内容需重点阐述“原因 (Why)”与“改动 (What)”。
5. **健壮的异常处理**: 针对各类外部网络请求与磁盘 I/O，必须编写充分的异常防御代码。不可随意吞噬错误；出现影响系统运行的严重异常时，应使用 `[Fatal]` 级别日志告警，并通过 SSE 同步至前端 UI 进行展示。
