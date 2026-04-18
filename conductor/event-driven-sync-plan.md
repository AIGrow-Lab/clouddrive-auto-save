# 后端事件驱动同步计划

**目标：** 消除前端所有的定时轮询（setInterval），通过利用 SSE (Server-Sent Events) 将数据库和状态变更直接推送到前端，实现极致的实时同步。

**架构转变：**
目前，`Tasks.vue` 和 `Dashboard.vue` 依赖 5 秒一次的轮询来获取更新。我们将通过现有的 SSE 日志流，引入后端驱动的事件系统来替换它。

**修改内容：**

### 1. 后端事件系统 (`internal/utils/events.go`)
- 创建事件广播辅助工具，将结构化 JSON 事件（如 `task_update`, `task_delete`, `stats_update`）序列化，并通过 `GlobalBroadcaster` 发送，使用特殊前缀 `[EVENT:{...}]`。

### 2. 后端事件触发器
- **`worker.go`**: 每当进度改变时发送 `task_update`。任务完成时发送 `stats_update`。
- **`router.go`**: 在任务创建、修改、删除及手动触发执行时，发送 `task_update` 和 `stats_update`。

### 3. 前端：任务管理 (`Tasks.vue`)
- 移除 5 秒一次的 `setInterval`。
- 初始化到 `/api/dashboard/logs` 的 SSE 连接。
- 解析 `[EVENT:...]` 消息：
  - `task_update`: 在本地列表中找到任务并执行原地更新。
  - `task_delete`: 将任务从本地列表中移除。

### 4. 前端：仪表盘 (`Dashboard.vue`)
- 移除定时轮询逻辑。
- 更新 SSE 消息处理函数，识别 `[EVENT:...]`。
- `stats_update`: 触发一次 `fetchStats()` 刷新统计数字，保证在无轮询的情况下数据依然 100% 准确。
- 确保所有的系统事件消息不会显示在终端日志窗口中。

这种转变确保了 UI 对后端变更的响应延迟降至毫秒级，减少了不必要的网络流量。
