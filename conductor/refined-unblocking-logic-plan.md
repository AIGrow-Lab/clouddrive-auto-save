# 精细化任务解封逻辑计划 (Refined Task Unblocking Plan)

## 1. Objective (目标)
优化任务更新逻辑，确保只有在用户修改了关键参数（分享链接或提取码）并保存时，系统才解除 `[Fatal]` 致命错误封锁。仅修改名称、路径等非核心字段时，保留原有的错误状态。

## 2. Key Files & Context (核心影响文件)
- `internal/api/router.go`: `updateTask` 函数。

## 3. Implementation Steps (实施步骤)
1. **引入参数比对逻辑**:
   - 在 `ShouldBindJSON` 执行前，记录数据库中现有的 `ShareURL` 和 `ExtractCode`。
   - 在执行 `Updates` 前进行比对。
2. **条件触发重置**:
   - 仅当新旧 `ShareURL` 或 `ExtractCode` 不一致时，才在 `updateData` 映射表中加入 `status: "pending"` 和 `message: ""`。
3. **编译并测试**:
   - 验证：修改任务名称保存，状态应保持 `LINK ERROR`；修改链接保存，状态应变为 `PENDING`。

## 4. Verification & Testing (验证与测试)
1. 准备一个处于 `LINK ERROR` 状态的任务。
2. 修改其“任务名称”，点击保存。确认 UI 依然显示红标。
3. 修改其“分享链接”，点击保存。确认红标消失，变为蓝色 `PENDING`。
