# 修复 E2E 测试不稳定性与竞态条件 (Fix E2E Instability & Race Conditions)

## 1. 背景与问题

用户反馈 `手动执行 139 转存任务` E2E 测试存在偶发性超时（60s）。经过排查，发现两个核心稳定性隐患：

- **消息丢失风险**: 后端 SSE 广播器给订阅客户端的缓冲区仅为 10。在任务结束时，由于会瞬间爆发多条日志与事件消息（任务完成、进度更新、状态更新、统计更新），极易导致关键的 `task_update` 事件被丢弃。
- **前端竞态条件**: 在 `Tasks.vue` 中，手动点击“运行”后，前端会异步调用 API 并在回调中设置 `row.status = 'running'`。如果任务执行极快（如在 E2E Mock 环境下），后端发来的 `success` 事件可能在 API 回调执行前就已更新了 UI。随后 API 回调执行，又将状态强制改回了 `running`，导致 UI 状态停滞，Playwright 无法检测到 `SUCCESS` 标记。

## 2. 修复方案

### 后端修复 (`internal/utils/broadcaster.go`)

- 将 `NewBroadcaster` 的总消息缓冲区从 100 增加到 1000。
- 将 `Subscribe` 返回的客户端通道缓冲区从 10 增加到 100。
- 确保在高并发或消息突发场景下，事件能稳健送达前端。

### 前端修复 (`web/src/views/Tasks.vue`)

- **逻辑加固**: 在 `handleRun` 和 `handleRunAll` 的 API 回调中增加状态检查。
- **修复竞态**: 仅在任务当前状态不是 `success` 时才手动设置为 `running`。这样如果 SSE 已经先一步将状态更新为成功，前端的回调将不会覆盖它。

## 3. 验证步骤

- 执行 `make check` 确保无代码风格和基础逻辑问题。
- 重复运行 `make e2e-test`。目前 139 执行测试已恢复稳健，能在 4s 内准确捕捉到状态变更，不再出现 60s 超时。
