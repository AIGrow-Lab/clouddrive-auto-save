# Bark 推送信息格式优化计划

## Objective (目标)
优化 Bark 推送通知的视觉呈现与听觉提醒，增强通知的信息密度，并根据任务状态（成功/失败）自动调整通知的优先级与提醒音。

## Proposed Changes (拟议改动)

### 1. 后端通知模块 (internal/core/notify/bark.go)
- **扩展 Payload 结构**: 在 `BarkPayload` 中增加 `Level`, `Sound`, `Badge`, `IsArchive` 等字段。
- **重构 SendBarkDirect**: 支持解析并发送新增的字段。
- **增强通知模板**: 
    - **成功**: 标题前缀 ✅，使用普通级别 (`active`)，提示音设为清脆声音。
    - **失败**: 标题前缀 ❌，使用时效性级别 (`timeSensitive`)，提示音设为警示音。
    - **正文**: 加入耗时信息、详细状态描述。

### 2. 任务执行引擎 (internal/core/worker/worker.go)
- **耗时统计**: 在 `execute` 函数开始时记录时间戳。
- **传递上下文**: 修改 `finishTask` 和 `SendTaskNotification` 的签名，将任务执行耗时传递给通知模块。

## Implementation Steps (实施步骤)

### 第一阶段：后端结构扩展
1. 修改 `internal/core/notify/bark.go`:
    - 更新 `BarkPayload` 结构体。
    - 修改 `SendBarkDirect` 以应用新字段。
    - 修改 `SendBark` 和 `SendTaskNotification` 的签名，增加 `duration` 参数。

### 第二阶段：逻辑与样式优化
1. 修改 `internal/core/notify/bark.go` 中的 `SendTaskNotification` 逻辑：
    - 根据 `status` 动态设置 `Level` 和 `Sound`。
    - 格式化 `body` 内容（包含耗时信息）。
2. 修改 `internal/core/worker/worker.go`:
    - 在 `execute` 方法中记录 `startTime`。
    - 修改 `finishTask` 调用，计算并传入 `duration`。

## Verification & Testing (验证与测试)
1. 运行任务，观察成功推送的样式、级别和声音。
2. 模拟任务失败，验证是否触发了时效性通知 (Time Sensitive) 和警示音。
3. 检查推送内容中的耗时信息是否准确。