# Bark 高级设置前端适配计划

## Objective (目标)
在前端的设置页面中添加针对 Bark 推送高级特性的适配（分别控制成功与失败的提示音与级别），并将这些配置项放在折叠面板中。同时增强测试通知的功能，允许用户自定义测试参数。

## Proposed Changes (拟议改动)

### 1. 前端 UI 修改 (web/src/views/Settings.vue)
- **高级选项折叠面板**: 增加一个 `<el-collapse>` 包含 Bark 的高级配置。
- **表单字段**:
  - `bark_success_sound` (成功提示音，如下拉选择 `birdsong.caf`, `minuet.caf` 等，默认 `birdsong.caf`)
  - `bark_success_level` (成功通知级别，默认 `active`)
  - `bark_failure_sound` (失败警示音，如 `alarm.caf`，默认 `alarm.caf`)
  - `bark_failure_level` (失败通知级别，默认 `timeSensitive`)
- **测试表单增强**: 点击“发送测试消息”时，弹出一个小对话框 (`el-dialog`)，允许用户覆盖当前的 `标题`, `正文`, `级别` 和 `铃声` 来发送自定义的测试请求。
- **状态管理**: 确保上述新增字段通过 `getGlobalSettings` 和 `updateGlobalSettings` 与后端正确同步。

### 2. 测试 API 修改 (internal/api/router.go)
- 修改 `testBarkNotification` 接口接收额外的 `title`, `body`, `level`, `sound` 参数。
- 若前端没有提供，默认使用“测试通知”等回退值。

### 3. 生产逻辑修改 (internal/core/notify/bark.go)
- 修改 `SendTaskNotification` 内部，从数据库（或直接传递）读取 `bark_success_sound`, `bark_success_level` 等配置。如果数据库中没有值，则退回到现有的默认值（成功: active/birdsong.caf, 失败: timeSensitive/alarm.caf）。

## Implementation Steps (实施步骤)
1. **修改后端测试 API**: 更新 `internal/api/router.go` 中的 `testBarkNotification`。
2. **修改后端业务通知**: 更新 `internal/core/notify/bark.go`，查询具体的通知设置并应用。
3. **修改前端视图**: 重构 `web/src/views/Settings.vue`，增加字段定义、折叠面板与测试对话框。
4. **测试与验证**: 运行 `npm run build` 和 `go run`，测试保存配置、打开测试面板以及完成实际任务时的通知表现。