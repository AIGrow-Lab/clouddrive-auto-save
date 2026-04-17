# 彻底修复任务状态死锁计划 (Final Fix Task Lock Bug Plan)

## 1. Objective (目标)
彻底解决当任务出现 `LINK ERROR` (致命错误) 后，用户编辑并更新链接也无法恢复任务状态、无法重新运行的严重交互 Bug。

## 2. Key Files & Context (核心影响文件)
- `internal/api/router.go`: `updateTask` 函数。

## 3. Root Cause Analysis (根本原因分析)
- 现有的 `updateTask` 逻辑仅执行了简单的 `db.DB.Save(&task)`。
- GORM 的 `Save` 方法在处理包含旧数据的结构体时，可能因为字段值相同或特殊原因未能正确清除 `message` 字段中的 `[Fatal]` 标记。
- 必须要显式地通过 `Updates(map)` 方式，强制将 `status` 刷回 `pending`，将 `message` 刷为空字符串 `""`。

## 4. Implementation Steps (实施步骤)
1. **重构后端 `updateTask` 逻辑**:
   - 接收 JSON 绑定后，不使用 `Save`。
   - 手动构建一个 `map[string]interface{}`，包含所有可变字段。
   - 显式加入 `"status": "pending"` 和 `"message": ""`。
   - 调用 `db.DB.Model(&task).Updates(updateData)` 执行物理更新。
2. **编译后端**:
   - 运行 `go build` 确保逻辑无误。

## 5. Verification & Testing (验证与测试)
1. 编辑一个红标任务，点击保存。
2. 确认其状态立刻变为蓝色的 `PENDING`，红标消失，“运行”按钮恢复高亮可用状态。
