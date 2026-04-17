# 🚀 Unified CloudDrive Auto-Save (UCAS)

[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Vue Version](https://img.shields.io/badge/Vue-3.5+-4FC08D?style=flat&logo=vue.js)](https://vuejs.org/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat&logo=docker)](https://www.docker.com/)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

**统一云盘自动转存系统** —— 这是一个由 Go 语言驱动的高性能、低资源占用的云盘自动化工具。它整合了**移动云盘 (139)** 与**夸克网盘 (Quark)** 的核心转存能力，并配备了现代化的 Vue 3 管理后台。

---

## ✨ 核心特性 (Features)

*   **⚡ 高性能引擎**：基于 Go Goroutine 实现的并发 Worker 池，支持多任务同时转存，效率远超传统脚本。
*   **🎨 现代化 UI**：采用 Vue 3 + Element Plus 构建的响应式管理后台，支持**暗黑模式**。
*   **📊 实时监控**：集成真实数据仪表盘，支持任务执行状态的**异步自动刷新**，实时汇总运行中任务、存储容量、活跃账号及最新转存动态。
*   **📡 深度排错**：通过 **Server-Sent Events (SSE)** 技术，将 139/Quark 网盘底层的 API 请求与响应日志实时推送到前端 Dashboard，极大方便了 Token 失效或转存异常时的故障排查。
*   **🤖 智能重命名**：
    *   **魔法变量**：支持 `{TASKNAME}`, `{YEAR}`, `{DATE}`, `{CHINESE}` 等自动提取与替换。
    *   **正则增强**：支持标准的正则表达式子组替换。
    *   **可视化起始点**：支持解析分享链接，允许用户在文件列表中直接勾选起始转存文件（支持时间排序）。
    *   **秒开回显**：优化起始转存点加载机制，实现文件名的本地持久化缓存，消除编辑任务时的解析延迟。
    *   **全自动执行**：已全面实现移动云盘 (139) 与夸克网盘 (Quark) 转存后的**自动重命名**，确保网盘文件结构始终整洁一致。
*   **🔗 深度整合**：
    *   **智能去重**：转存前自动预检目标目录，智能跳过同名文件，减少冗余操作并降低风控风险。
    *   **交互升级**：支持大屏网盘目录树形弹窗选择，**重构虚拟根节点逻辑**，支持在根目录直接创建与管理文件夹。
    *   **智能防御**：增加前端新建文件夹**同名冲突事前拦截**，抹平 139（自动重命名）与夸克（直接报错）的行为差异。
    *   **UI 优化**：存储空间进度条支持**状态感知色**（绿/黄/红），修复容量超额时的 NaN 显示错误。
    *   **安全加固**：增加账号删除时的关联任务强校验，防止误删正在运行的账号；**完善 139 容量统计模型**，支持个人与家庭空间的自动汇总。
    *   **日志同步**：仪表盘“清空日志”支持前后端同步彻底擦除。
    *   **移动云盘 (139)**：支持 Authorization/Cookie 登录，完美模拟 HCY 私有接口；**优化昵称获取逻辑**，自动跳过脱敏星号名称；**修复时间解析 Bug**，确保分享链接内的文件更新时间准确显示。
*   **📦 单体化部署**：前端资源内嵌于 Go 二进制文件中，一个文件即可启动完整服务。
*   **🛡️ 生产级可靠性**：支持 Docker 一键部署，数据库持久化，异常自动退避重试。

---

## 🏗️ 技术架构 (Architecture)

*   **Backend**: Go 1.23, Gin (Web Framework), GORM (ORM), SQLite3 (Database).
*   **Frontend**: Vue 3, Vite, Element Plus, Lucide Icons, Pinia (Store).
*   **Core**: 
    *   `internal/core/worker`: 并发调度引擎。
    *   `internal/core/renamer`: 正则与模板重命名引擎。
    *   `internal/core/cloud*`: 云盘协议适配层。

---

## 🚀 快速开始 (Quick Start)

### 方式一：Docker Compose (推荐)

这是最简单的部署方式，适合生产环境。

1.  下载 `docker-compose.yml`。
2.  在同级目录下执行：
    ```bash
    docker-compose up -d --build
    ```
3.  访问浏览器：`http://your-ip:8080`

### 方式二：源码编译

1.  **构建前端**：
    ```bash
    cd web
    npm install
    npm run build
    ```
2.  **编译后端**：
    ```bash
    cd ..
    go mod tidy
    go build -o ucas cmd/server/main.go
    ```
3.  **运行**：
    ```bash
    ./ucas
    ```

---

## 📖 使用指南 (Usage)

### 1. 账号配置
*   **139 移动云盘**：推荐通过浏览器抓包获取 `Authorization` (Basic 格式)，最为稳定。
*   **夸克网盘**：通过浏览器 F12 获取 `Cookie` 全量字符串。

### 2. 任务创建
*   **分享链接**：直接粘贴 139 或 夸克分享链接，系统会自动解析 ID。
*   **保存路径**：填写云盘内的绝对路径（如 `/电影/2024/`），系统会自动按需创建不存在的文件夹。

### 3. 重命名示例
*   **Pattern**: `^【电影TT】(.*)\.mp4$`
*   **Replacement**: `{TASKNAME}.$1.{YEAR}.mp4`
*   **效果**：将名为 `【电影TT】绝命毒师.S01E01.mp4` 的文件自动重命名为 `我的任务名.绝命毒师.S01E01.2024.mp4`。

---

## 文件夹结构 (Project Structure)

```text
.
├── cmd/server          # 编译入口
├── internal/
│   ├── api             # REST API 路由与控制器
│   ├── core/
│   │   ├── cloud139    # 139 SDK 移植
│   │   ├── quark       # 夸克 SDK 移植
│   │   ├── renamer     # 重命名引擎
│   │   └── worker      # 任务调度器
│   └── db              # 数据库模型与初始化
├── web/                # Vue 3 前端源码
├── Dockerfile          # 多阶段构建文件
└── RESTRUCTURE_PLAN.md # 重构详细规划
```

---

## 🤝 贡献与反馈

如果您在使用过程中发现 Bug 或有新的建议，欢迎提交 Issue 或 Pull Request。

**鸣谢**：本项目灵感与部分逻辑参考自原 `cloudpan-auto-save` (Node.js) 与 `quark-auto-save` (Python) 项目。

---

## 📄 开源协议

基于 [MIT License](LICENSE) 协议。
