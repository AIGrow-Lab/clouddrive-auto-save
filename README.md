# 🚀 Unified CloudDrive Auto-Save (UCAS)

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Vue Version](https://img.shields.io/badge/Vue-3.5+-4FC08D?style=flat&logo=vue.js)](https://vuejs.org/)
[![Docker Image](https://img.shields.io/badge/Docker-zcq98%2Fucas-2496ED?style=flat&logo=docker)](https://hub.docker.com/r/zcq98/clouddrive-auto-save)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![CI/CD](https://github.com/zhaocongqi/clouddrive-auto-save/actions/workflows/docker-publish-main.yml/badge.svg)](https://github.com/zhaocongqi/clouddrive-auto-save/actions)

**统一云盘自动转存系统 (UCAS)** —— 这是一个由 Go 语言驱动的高性能、低资源占用的云盘自动化工具。它整合了**移动云盘 (139)** 与**夸克网盘 (Quark)** 的核心转存能力，并配备了现代化的 Vue 3 管理后台。

---

## ✨ 核心特性 (Features)

*   **⚡ 高性能引擎**：基于 Go Goroutine 实现的并发 Worker 池，支持多任务同时转存。
*   **🛠️ 跨平台兼容**：采用 **CGO-free** 的纯 Go SQLite 驱动，支持 Windows/Linux/macOS 零依赖部署。支持通过 GitHub Releases 下载预编译好的二进制程序。
*   **🎨 现代化 UI**：采用 Vue 3 + Element Plus 构建的响应式后台，支持暗黑模式与等宽日志视图。
*   **📊 实时指挥中心**：集成实时数据仪表盘，通过 **Server-Sent Events (SSE)** 实现任务状态与日志的绝对实时同步。
*   **🤖 智能整理与去重**：
    *   **正则重命名**：支持强大的正则匹配与替换（含 `{TASKNAME}`, `{DATE}` 等魔法变量）。
    *   **智能去重**：转存前自动预检目标目录，智能跳过同名文件，防止产生冗余副本。
    *   **可视化解析**：支持解析分享链接，允许手动选择起始转存点。
*   **⏰ 灵活调度**：支持“全局默认”与“任务自定义”双层 Cron 调度逻辑。
*   **📦 容器化优先**：提供官方 Docker 镜像，支持 **多架构 / 异构镜像 (amd64 / arm64)**，适配树莓派、Mac M系列等 ARM 设备。支持 GitHub Actions 自动构建与发布。

---

## 🚀 快速开始 (Quick Start)

### 方式一：Docker (推荐)

直接使用 Docker Hub 上的预编译镜像：

```bash
docker run -d \
  --name ucas \
  -p 8080:8080 \
  -v $(pwd)/data:/app/data \
  -e TZ=Asia/Shanghai \
  zcq98/clouddrive-auto-save:latest
```

或使用 `docker-compose.yml`:

```yaml
services:
  ucas:
    image: zcq98/clouddrive-auto-save:latest
    container_name: clouddrive-auto-save
    ports:
      - "8080:8080"
    volumes:
      - ./data:/app/data
    environment:
      - LOG_LEVEL=INFO
      - TZ=Asia/Shanghai
```

### 方式二：直接下载

前往 [GitHub Releases](https://github.com/zhaocongqi/clouddrive-auto-save/releases) 页面，下载适合您操作系统和架构的预编译压缩包，解压后即可运行。

### 方式三：本地编译

1.  **克隆仓库**: `git clone https://github.com/zhaocongqi/clouddrive-auto-save.git`
2.  **一键构建**: `make build` (依赖 Node.js 和 Go 环境)
3.  **运行**: `./bin/ucas`

---

## 🏗️ 技术架构 (Architecture)

*   **Backend**: Go 1.25, Gin, GORM, **Glebarez SQLite (CGO-free)**.
*   **Frontend**: Vue 3, Vite, Element Plus.
*   **CI/CD**: GitHub Actions (使用 Docker Buildx 构建多架构镜像，并使用 **GoReleaser** 自动发布多平台二进制产物)。

---

## ⚙️ 配置说明 (Configuration)

| 变量 | 说明 | 默认值 |
| :--- | :--- | :--- |
| `LOG_LEVEL` | 日志等级 (`DEBUG`, `INFO`, `WARN`, `ERROR`) | `INFO` |
| `DB_PATH` | SQLite 数据库文件路径 | `/app/data/data.db` |
| `LISTEN_ADDR` | 服务监听地址与端口 | `0.0.0.0:8080` |
| `TZ` | 系统时区（主要用于 Docker 内部日志时间对齐） | `Asia/Shanghai` |

---

## 🛠️ Makefile 指令

*   **`make build`**：完整构建（前端构建 + 后端内嵌编译）。
*   **`make check`**：执行 lint、vet 和单元测试。
*   **`make docker-build`**：本地构建 Docker 镜像。

---

## 📖 详细文档

更多细节请参考 [docs/](docs/) 目录：
*   [API 接口文档](docs/api/README.md)
*   [云盘底层 API 手册](docs/cloud_drive_apis.md)
*   [数据库设计](docs/api/database.md)
