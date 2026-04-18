# 指挥中心仪表盘升级实施计划 (Command Center Dashboard Upgrade Implementation Plan)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将仪表盘重构为实时监控中心，包含基于 SSE 的实时日志流、任务具体执行进度展示以及全局快捷操作悬浮栏。

**Architecture:** 
1. **后端 (SSE)**：实现全局日志广播器 (Broadcaster)，新增 `/api/dashboard/logs` 接口。
2. **前端 (Console)**：重构 Dashboard 布局，新增实时滚动终端组件和任务进度卡片。
3. **交互 (FAB)**：实现全局悬浮快捷按钮组。

**Tech Stack:** Go (SSE), Vue 3, Element Plus, Lucide Icons

---

### Task 1: 实现后端全局日志广播器

**Files:**
- Create: `internal/utils/broadcaster.go`
- Modify: `internal/api/router.go`

- [ ] **Step 1: 编写 Broadcaster 逻辑**
  实现一个支持多客户端订阅、广播字符串消息的单例对象。包含 `Subscribe`, `Unsubscribe` 和 `Broadcast` 方法。

- [ ] **Step 2: 劫持标准日志输出**
  在 `main.go` 或核心初始化位置，通过自定义 `io.Writer` 将系统的 `log.Printf` 同步镜像到 Broadcaster 中。

- [ ] **Step 3: 注册 SSE 路由**
  在 `router.go` 中增加 `GET /api/dashboard/logs` 接口，使用 `c.Stream` 将广播器的消息流式推送给前端。

### Task 2: 前端 Dashboard 布局重构

**Files:**
- Modify: `web/src/views/Dashboard.vue`

- [ ] **Step 1: 实现左右分栏布局**
  使用 `el-row` 重新排列页面。左侧 65% 用于日志终端，右侧 35% 用于运行中任务列表。

- [ ] **Step 2: 开发实时终端组件**
  编写具有深色背景、等宽字体的日志显示区。实现自动滚动到底部的逻辑，并增加简单的 ANSI 颜色或关键词（[INFO], [ERROR]）着色。

- [ ] **Step 3: 对接 EventSource**
  在 `onMounted` 中建立与 `/api/dashboard/logs` 的连接，将收到的日志实时追加到终端组件中。

### Task 3: 实现任务执行监控展示

**Files:**
- Modify: `internal/core/worker/worker.go`
- Modify: `web/src/views/Dashboard.vue`

- [ ] **Step 1: 在 Worker 关键节点插入进度日志**
  修改 `worker.go`，在解析、转存、重命名开始时，向广播器发送格式化消息（如 `[PROGRESS:TaskID:Stage:Message]`）。

- [ ] **Step 2: 前端解析进度标识**
  在 Dashboard 接收日志的日志中，识别 `[PROGRESS]` 前缀的消息，不展示在终端，而是更新右侧任务列表中的状态百分比或步骤描述。

### Task 4: 实现全局快捷操作悬浮栏 (FAB)

**Files:**
- Modify: `web/src/views/Dashboard.vue` (暂放在此处作为局部功能，或放在 Layout 中作为全局功能)

- [ ] **Step 1: 编写悬浮操作组组件**
  使用固定定位（Bottom-Right）实现一个美观的悬浮球，点击展开“添加账号”、“新建任务”、“清空当前日志”三个快捷按钮。

### Task 5: 验证与测试

- [ ] **Step 1: 联调测试**
  启动任务，验证 Dashboard 左侧是否出现毫秒级刷新的 API 调用日志，右侧任务是否能显示“正在解析链接...”等实时进度。
- [ ] **Step 2: 压力测试**
  开启多个浏览器窗口同时访问 Dashboard，验证 SSE 广播在高并发连接下的稳定性。