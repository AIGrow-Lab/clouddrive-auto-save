# 修复移动云盘重命名接口请求失败计划 (Fix 139 Rename API Plan)

## 1. Objective (目标)
修复 139 移动云盘在转存文件后的“重命名操作”失效的问题。

## 2. Key Files & Context (核心影响文件)
- `internal/core/cloud139/client.go`：包含 139 云盘的所有 API 封装。其中 `RenameFile` 方法目前的 URL 路径和参数结构有误。

## 3. Root Cause Analysis (根本原因分析)
经过分析原版 Node.js 项目 (`cloudpan-auto-save`)，发现：
- 139 移动云盘正确的单文件重命名 API 路径为：`/hcy/file/update`。
- 请求体 JSON 的结构应为：`{"fileId": "xxx", "name": "new_name"}`。
- 我们现有的 Go 代码中，错误地猜测使用了 `/hcy/file/batchUpdate` 接口，并且构建了复杂的 `fileUpdateList` 结构，导致移动云盘服务器无法识别该请求，重命名静默失败。

## 4. Implementation Steps (实施步骤)
1. 打开 `internal/core/cloud139/client.go`。
2. 定位到 `RenameFile` 方法。
3. 将请求 URL 修正为 `PersonalKdNjsURL + "/hcy/file/update"`。
4. 将请求体 `body` 简化为 `map[string]interface{}{"fileId": fileID, "name": newName}`。

## 5. Verification & Testing (验证与测试)
1. 编译并启动服务。
2. 触发一次 139 移动云盘的转存重命名任务。
3. 确认转存成功后，目标网盘中的文件确实被改成了正则定义的名称。