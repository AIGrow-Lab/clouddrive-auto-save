# 任务执行监控 UI 重构计划（数据库驱动版）

**目标：** 通过将进度存储在数据库中，并结合轮询与 SSE 实时同步，将“任务执行监控”卡片转化为一个真实、持久的监控器。

### 1. 数据库模型 (`internal/db/db.go`)
- [x] 在 `Task` 结构体中添加 `Percent` (int) 和 `Stage` (string) 字段。

### 2. Worker 引擎 (`internal/core/worker/worker.go`)
- [x] 实现 `updateProgress` 辅助函数，用于持久化进度、阶段和消息。
- [x] 重构 `execute` 和 `finishTask` 方法以调用该函数。

### 3. 后端 API (`internal/api/router.go`)
- [x] 更新 `getDashboardStats` 接口，使其返回所有运行中、失败、以及最近 15 秒内成功的任务列表。

### 4. 前端 UI (`web/src/views/Dashboard.vue`)
- [x] 移除脆弱的日志回放 (`isReplay`) 和手动超时逻辑。
- [x] 在 `onMounted` 中实现 5 秒一次的自动轮询。
- [x] 在 `fetchStats` 中实现后端数据与本地状态的智能合并。
- [x] 更新进度处理函数，专门负责高频的视觉更新。
