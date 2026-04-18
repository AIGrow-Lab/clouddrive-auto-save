# 数据库驱动的任务进度持久化计划

**目标：** 通过将任务进度（百分比和阶段）存储在数据库中，解决所有进度条竞争、任务丢失以及刷新页面后重新出现的 Bug。

**架构转变：**
目前，进度信息（Percent, Stage）仅存在于瞬时的 SSE 消息中。如果发生页面刷新，前端会尝试通过回放日志来重建状态，这种方式非常脆弱且容易引起时间冲突。我们将把进度状态移至 `Task` 数据库模型中。

**修改内容：**

### 1. 数据库模型 (`internal/db/db.go`)
- 在 `Task` 结构体中增加 `Percent int` 和 `Stage string` 字段，实现进度状态的持久化。

### 2. Worker 引擎 (`internal/core/worker/worker.go`)
- 创建辅助函数 `updateProgress(task, percent, stage, message)`，用于更新数据库并发送 SSE 日志。
- 将 `execute` 函数中原本分散的进度日志替换为对 `updateProgress` 的调用。
- 更新 `finishTask` 方法，在结束时将进度设为 100% 并在数据库中记录最终阶段。

### 3. 后端 API (`internal/api/router.go`)
- 更新 `getDashboardStats` 接口中的 `runningTasksList` 查询：
  - 包含 `status = "running"` 的任务。
  - 包含 `status = "failed"` 的任务（确保异常状态持续可见）。
  - 包含 `status = "success" 且最后运行时间在 10 秒内` 的任务（实现成功任务的自然过期消失）。

### 4. 前端 UI (`web/src/views/Dashboard.vue`)
- **移除** 脆弱的 `setTimeout` / `dismissTask` 和 `isReplay` 逻辑。
- **移除** 针对进度日志的历史日志回放逻辑。
- 在 `onMounted` 中增加 5 秒一次的 `setInterval` 轮询。
- 实现智能合并算法，将 API 返回的任务列表与本地状态合并（更新现有项、添加新项、移除过期项），防止 UI 闪烁。

此方案将数据库作为唯一的“真理源”，从根本上简化了前端状态管理并保证了 100% 的准确性。
