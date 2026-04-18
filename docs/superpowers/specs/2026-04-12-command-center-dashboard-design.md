# 实时指挥中心仪表盘升级 (Command Center Dashboard Upgrade)

## 1. 背景与动机
当前的仪表盘仅显示静态统计数字，无法实时观察系统的运行过程。用户需要一个能够实时看到“系统正在做什么”、“哪个文件正在存”、“API 调用是否成功”的监控大屏，同时具备快速处理异常任务的能力。

## 2. 目标范围
*   **实时日志系统**：建立基于 SSE (Server-Sent Events) 的双向同步机制，将后端日志实时推送到前端终端窗口。
*   **任务执行监控**：右侧栏实时展示当前运行中的任务及其具体阶段（解析/转存/重命名）。
*   **一键式指挥交互**：支持日志清空、失败任务原地重试。
*   **全局快捷导航**：新增悬浮操作栏，快速跳转至账号添加或任务创建。

## 3. 系统设计细节

### 3.1 后端架构 (Go)
*   **Log Broadcaster**：在 `internal/utils` 或 `internal/core/worker` 中实现一个全局广播器。
    *   通过 `chan string` 接收所有模块的日志输入。
    *   维护一个 `clients map[chan string]bool` 列表，支持多个 Web 页面同时在线监听。
*   **SSE 接口 (`GET /api/dashboard/logs`)**：
    *   使用 Gin 的流式响应功能，将 Broadcaster 中的日志推送给前端。
*   **任务状态细化**：
    *   在 `worker.execute` 的各个关键节点发送特殊的“状态更新日志”（如 `[PROGRESS:TaskID:Stage:Message]`）。

### 3.2 前端架构 (Vue 3)
*   **Log Console 组件**：
    *   深色等宽字体样式。
    *   实现自动滚动至底部逻辑。
    *   关键词染色渲染（[INFO]->绿, [ERROR]->红, [WARN]->黄）。
*   **Running Tasks 组件**：
    *   监听 SSE 中的进度标识，更新本地任务卡片的状态。
*   **快捷浮动操作 (FAB)**：
    *   使用 Element Plus 的 `el-affix` 或自定义固定定位，实现右下角的快捷操作球。

## 4. 交互流程
1.  用户打开 Dashboard。
2.  前端请求 `/api/dashboard/logs/recent` 填充前 50 条记录。
3.  前端建立 `EventSource` 连接。
4.  Worker 执行任务，产生日志 -> 进入广播器 -> 通过 SSE 传至前端。
5.  日志包含进度标签时，右侧任务列表同步更新图标和文字。

## 5. 后续扩展
*   支持日志级别过滤（如仅显示 Error）。
*   支持点击日志中的任务 ID 直接跳转到任务详情。