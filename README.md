# 🚀 Unified CloudDrive Auto-Save (UCAS)

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Vue Version](https://img.shields.io/badge/Vue-3.5+-4FC08D?style=flat&logo=vue.js)](https://vuejs.org/)
[![Docker Image](https://img.shields.io/badge/Docker-zcq98%2Fucas-2496ED?style=flat&logo=docker)](https://hub.docker.com/r/zcq98/clouddrive-auto-save)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![CI/CD](https://github.com/zhaocongqi/clouddrive-auto-save/actions/workflows/docker-publish-main.yml/badge.svg)](https://github.com/zhaocongqi/clouddrive-auto-save/actions)

**统一云盘自动转存系统 (UCAS)** —— 这是一个由 Go 语言驱动的高性能、低资源占用的云盘自动化工具。它整合了**移动云盘 (139)** 与**夸克网盘 (Quark)** 的核心转存能力，并配备了现代化的 Vue 3 管理后台。

---

## ✨ 核心特性 (Features)

* **⚡ 高性能引擎**：基于 Go Goroutine 实现的并发 Worker 池，支持多任务同时转存。
* **🛠️ 跨平台兼容**：采用 **CGO-free** 的纯 Go SQLite 驱动，支持 Windows/Linux/macOS 零依赖部署。支持通过 GitHub Releases 下载预编译好的二进制程序。
* **🎨 现代化 UI**：采用 Vue 3 + Element Plus 构建的响应式后台，支持暗黑模式与等宽日志视图。
* **📊 实时指挥中心**：集成实时数据仪表盘，通过 **Server-Sent Events (SSE)** 实现任务状态与日志的实时同步，并支持任务执行耗时统计。
* **🤖 智能整理与去重**：
  * **正则重命名**：支持强大的正则匹配与替换（含 `{TASKNAME}`, `{DATE}` 等魔法变量）。
  * **智能去重**：转存前自动预检目标目录，智能跳过同名文件，防止产生冗余副本。
  * **可视化解析**：支持解析分享链接，允许手动选择起始转存点。跳转外链时支持自动拼接并复制提取码至剪贴板。
  * **双重过滤机制**：系统仅转存 **“在起始文件之后（含）更新”** 且 **“符合正则匹配”** 的文件。
  * **消息通知**：集成 Bark 推送，支持分级提醒、自定义铃声及**智能批量汇总通知**（避免执行所有任务时产生过量推送）。
* **🛡️ 健壮的错误处理**：针对网盘接口错误码实现人性化清洗（如提取码错误、链接失效等），支持 `[Fatal]` 致命错误自动阻断与 UI 警示。
* **⏰ 灵活调度**：支持“全局默认”与“任务自定义”双层 Cron 调度逻辑，全局调度提供“简易/高级”双模式配置。
* **📦 容器化优先**：提供官方 Docker 镜像，支持 **多架构 / 异构镜像 (amd64 / arm64)**，适配树莓派、Mac M系列等 ARM 设备。

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
  zcq98/clouddrive-auto-save:v1.1.0
```

或使用 `docker-compose.yml`:

```yaml
services:
  ucas:
    image: zcq98/clouddrive-auto-save:v1.1.0
    container_name: clouddrive-auto-save
    ports:
      - "8080:8080"
    volumes:
      - ./data:/app/data
    environment:
      - LOG_LEVEL=INFO
      - TZ=Asia/Shanghai
```

### 方式二：直接下载产物

前往 [GitHub Releases](https://github.com/zhaocongqi/clouddrive-auto-save/releases) 页面，下载适合您操作系统和架构的预编译压缩包，解压后即可直接运行。

### 方式三：源码编译

1. **克隆仓库**: `git clone https://github.com/zhaocongqi/clouddrive-auto-save.git`
2. **一键构建**: `make build` (依赖 Node.js 和 Go 环境)
3. **运行**: `./bin/ucas`

---

## ⚙️ 配置说明 (Configuration)

系统支持通过环境变量进行微调：

| 变量 | 说明 | 默认值 |
| :--- | :--- | :--- |
| `LOG_LEVEL` | 日志等级 (`DEBUG`, `INFO`, `WARN`, `ERROR`) | `INFO` |
| `DB_PATH` | SQLite 数据库文件路径 | `/app/data/data.db` |
| `LISTEN_ADDR` | 服务监听地址与端口 | `0.0.0.0:8080` |
| `TZ` | 系统时区（主要用于日志时间与定时任务对齐） | `Asia/Shanghai` |

---

## 🔮 魔法变量与重命名示例

魔法变量允许您通过简单的占位符实现自动整理。其核心区别在于：`{TASKNAME}` 是您**手动定义**的，而其他变量是**自动从原文件名中抓取**的。

### 变量提取说明 (假设任务名为：`经典电影`)

| 变量 | 来源/逻辑 | 示例 (原名: `[2024.04.20] 功夫熊猫4.mp4`) |
| :--- | :--- | :--- |
| `{TASKNAME}` | **任务显示名称** | `经典电影` (您在创建任务时填写的名字) |
| `{OLDNAME}` | **原始完整文件名** | `[2024.04.20] 功夫熊猫4.mp4` |
| `{CHINESE}` | 提取文件名中的连续中文 | `功夫熊猫` |
| `{DATE}` | 提取日期并清洗为纯数字 | `20240420` |
| `{YEAR}` | 提取 4 位年份数字 | `2024` |
| `{EXT}` | 提取原始文件后缀 | `mp4` |

### 常用组合示例

* **场景 1：简单加前缀（无需正则匹配）**
  * **替换规则**: `[已转存]_{OLDNAME}`
  * **效果**: `movie.mp4` ➡️ `[已转存]_movie.mp4`
* **场景 2：统一标题归档**
  * **替换规则**: `[{DATE}] {TASKNAME}.{EXT}`
  * **效果**: `[极速版] 2024.04.20 功夫熊猫4.mp4` ➡️ `[20240420] 经典电影.mp4`
* **场景 3：正则捕获组（高级用法）**
  * **匹配正则**: `.* (.*)\.mp4`
  * **替换规则**: `${1}_高清版.mp4`
  * **效果**: `[2024.04.20] 功夫熊猫4.mp4` ➡️ `功夫熊猫4_高清版.mp4` (使用 `${1}` 引用正则子组)

---

## 🏗️ 技术架构 (Architecture)

* **Backend**: Go 1.25, Gin, GORM, **Glebarez SQLite (CGO-free)**.
* **Frontend**: Vue 3, Vite, Element Plus.
* **CI/CD**: GitHub Actions (使用 Docker Buildx 构建多架构镜像，并使用 **GoReleaser** 自动发布多平台产物)。

---

## 📖 详细文档

更多细节请参考 [docs/](docs/) 目录：

* [API 接口文档](docs/api/README.md)
* [移动云盘 (139) API 手册](docs/cloud139_apis.md)
* [夸克网盘 (Quark) API 手册](docs/quark_apis.md)
* [数据库设计](docs/api/database.md)

---

## 🛠️ Makefile 指令

* **`make build`**：完整构建（前端构建 + 后端内嵌编译）。
* **`make check`**：执行 lint、vet 和单元测试。
* **`make docker-build`**：本地构建 Docker 镜像。

---

## 🤝 贡献与反馈

如果您在使用过程中发现 Bug 或有新的建议，欢迎提交 Issue。

---

## 📄 开源协议

基于 [MIT License](LICENSE) 协议。
