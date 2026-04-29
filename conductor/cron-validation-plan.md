# 全局 Cron 表达式校验增强计划 (Enhanced Cron Validation Plan)

## 背景与目标 (Objective)

在“高级 Cron”模式下，用户输入非法的 Cron 表达式时，前端缺乏即时校验，后端在 `updateGlobalSettings` 接口中也未进行格式检查。这会导致无效配置入库并可能引发调度器错误。
目标是在后端接口增加格式校验，并在前端提供明确的错误反馈。

## 关键文件 (Key Files & Context)

- `internal/api/router.go`: `updateGlobalSettings` 接口。
- `web/src/views/Settings.vue`: 设置页面保存逻辑。

## 实施步骤 (Implementation Steps)

### 1. 后端接口校验

修改 `internal/api/router.go` 中的 `updateGlobalSettings` 函数：

- 在保存 `global_schedule_cron` 之前，判断是否同时开启了全局调度。
- 如果开启，则调用 `scheduler.ValidateCron` 进行校验。
- 校验失败则返回 `400 Bad Request` 并附带具体错误信息。

### 2. 前端保存逻辑优化

修改 `web/src/views/Settings.vue`：

- 优化 `saveSettings` 函数：当 API 返回错误时，从 `error.response.data.error` 中提取具体的错误原因并显示给用户（例如通过 `ElMessage.error`）。
- 考虑在 `advanced` 模式下增加前端初步校验（可选，优先保证后端强校验）。

## 验证与测试 (Verification & Testing)

1. 进入“系统设置” -> “高级 Cron”。
2. 输入一个明显错误的表达式（如 `abc` 或 `* * *` 少于 5 位）。
3. 失去焦点或按回车触发保存。
4. 验证是否弹出红色的错误提示，且该非法值未成功应用到系统调度中。
