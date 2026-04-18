# 🚀 Unified CloudDrive Auto-Save (UCAS)

[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Vue Version](https://img.shields.io/badge/Vue-3.5+-4FC08D?style=flat&logo=vue.js)](https://vuejs.org/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat&logo=docker)](https://www.docker.com/)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

**统一云盘自动转存系统** —— 这是一个由 Go 语言驱动的高性能、低资源占用的云盘自动化工具。它整合了**移动云盘 (139)** 与**夸克网盘 (Quark)** 的核心转存能力，并配备了现代化的 Vue 3 管理后台。

---

## ✨ 核心特性 (Features)

*   **⚡ 高性能引擎**：基于 Go Goroutine 实现的并发 Worker 池，支持多任务同时转存，效率远超传统脚本。
*   **🎨 现代化 UI**：采用 Vue 3 + Element Plus 构建的响应式管理后台，支持**等宽字体监控视图**与暗黑模式。
*   **📊 指挥中心**：集成实时数据仪表盘，支持任务执行状态的**绝对实时同步**，一键清空日志可联动重置全站运行摘要。
*   **📡 深度排错**：通过 **Server-Sent Events (SSE)** 技术，实时推送网盘底层 API 日志，极大方便故障排查。
*   **🤖 智能整理**：
    *   **魔法变量**：支持 `{TASKNAME}`, `{YEAR}`, `{DATE}`, `{CHINESE}` 等自动提取。
    *   **可视化选择**：解析分享链接，允许在列表中直接勾选起始转存文件。
    *   **全自动执行**：转存后自动根据规则重命名，确保网盘结构整洁一致。
*   **⏰ 智能调度 (Scheduling)**：
    *   **双层调度逻辑**：支持全局默认配置 + 任务独立重写，满足精细化自动转存需求。
    *   **内置校验**：后端实时验证 Cron 表达式有效性，确保任务可靠触发。
*   **🔗 深度整合**：
    *   **智能去重**：转存前自动预检目标目录，智能跳过同名文件。
    *   **交互升级**：支持大屏网盘目录树形选择，支持在根目录直接创建文件夹。
    *   **人性化报错**：内置错误字典，将晦涩的 API 错误实时转化为可读中文。
    *   **自动自愈**：服务端启动时自动检测并重置异常中断的任务状态，杜绝幽灵锁定。
*   **🤖 AI 增强 (Roadmap)**：
    *   **AI 智能重命名**：计划接入 LLM，通过自然语言指令全自动完成整理。
    *   **AI 定时助手**：支持自然语言设定频率（如“每周五下午两点”），AI 自动生成 Cron。
*   **📦 单体化部署**：前端资源内嵌于 Go 二进制文件中，一键即可启动完整服务。

---

## 🏗️ 技术架构 (Architecture)

*   **Backend**: Go 1.23, Gin (Web Framework), GORM (ORM), SQLite3 (Database).
*   **Frontend**: Vue 3, Vite, Element Plus, Lucide Icons.
*   **Scheduler**: 基于 `robfig/cron/v3` 的高性能调度引擎。

---

## 🚀 快速开始 (Quick Start)

### 方式一：Docker Compose (推荐)

```bash
docker-compose up -d --build
```

### 方式二：源码编译

1.  **构建前端**：`cd web && npm install && npm run build`
2.  **编译后端**：`go build -o ucas cmd/server/main.go`
3.  **运行**：`./ucas`

---

## 📖 使用指南 (Usage)

### 1. 账号配置
*   **139 移动云盘**：推荐优先使用 `Authorization` (Basic 格式)。
*   **夸克网盘**：获取 `Cookie` 全量字符串。

### 2. 调度模式说明
*   **跟随全局**：受全局总开关控制，按全局统一频率运行。
*   **自定义频率**：独立运行，需输入 6 位 Cron 表达式（带秒级）。
*   **手动执行**：仅在点击“运行”按钮时触发。

---

## 🤝 贡献与反馈

如果您在使用过程中发现 Bug 或有新的建议，欢迎提交 Issue。

---

## 📄 开源协议

基于 [MIT License](LICENSE) 协议。
