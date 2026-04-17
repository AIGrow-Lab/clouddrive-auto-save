# 优化起始转存点选择：已存在文件视觉提示计划 (Optimized Checkpoint Selection Plan)

## 1. Objective (目标)
在“选择起始转存文件”的对话框中，智能识别并标注那些已经存在于目标网盘目录中的文件。通过视觉弱化（置灰）和标签提示（已同步），帮助用户更清晰地分辨哪些是新资源，从而做出更合理的转存点选择。

## 2. Key Files & Context (核心影响文件)
- **后端核心模型**: `internal/core/drive.go`
  - 在 `FileInfo` 结构体中新增 `IsExisted` 字段。
- **后端 API 控制器**: `internal/api/task_preview.go`
  - 增强 `parseShareLinkInfo` 函数，支持可选的 `save_path` 参数，并执行同名检查逻辑。
- **前端视图**: `web/src/views/Tasks.vue`
  - 更新选择弹窗的表格渲染，支持根据 `is_existed` 显示标签和调整行样式。

## 3. Implementation Steps (实施步骤)

### Step 3.1: 后端数据结构与逻辑增强
1. **修改 `internal/core/drive.go`**:
   - 在 `FileInfo` 结构体中添加：`IsExisted bool `json:"is_existed"``。
2. **重构 `internal/api/task_preview.go` 中的 `parseShareLinkInfo`**:
   - 接收参数中增加 `SavePath string `json:"save_path"``。
   - 如果 `SavePath` 不为空：
     - 调用 `driver.PrepareTargetPath(ctx, req.SavePath)` 获取目标目录 ID。
     - 调用 `driver.ListFiles(ctx, targetID)` 获取已存在文件列表。
     - 构建 `map[string]bool` 存储已存在的文件名。
     - 遍历解析出的分享文件列表，设置 `IsExisted` 状态。

### Step 3.2: 前端 API 与视图更新
1. **修改 `web/src/views/Tasks.vue` 中的 `openStartFileDialog`**:
   - 在调用 `parseShareLink` 时，将 `form.value.save_path` 作为参数传递。
2. **重构表格渲染**:
   - 为 `el-table` 添加 `:row-class-name="tableRowClassName"`。
   - 在文件名列增加一个 `el-tag`，当 `row.is_existed` 为真时显示“已在网盘中”。
3. **样式美化**:
   - 在 `<style scoped>` 中定义 `existed-row` 类，设置背景色微调和文字半透明。

## 4. Verification & Testing (验证与测试)
- **精准识别测试**: 选择一个包含部分已转存文件的分享链接，打开选择弹窗，确认已存在的文件是否被正确标记为“已在网盘中”。
- **性能测试**: 确认在加入目录预检逻辑后，解析分享链接的速度是否依然在可接受范围内（毫秒级增加）。
- **交互测试**: 确认置灰的文件行依然可以点击选择（保留灵活性），且标签显示清晰。
