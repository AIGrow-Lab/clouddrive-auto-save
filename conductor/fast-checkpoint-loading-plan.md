# 起始转存点加载体验优化计划 (Fast Checkpoint Loading Plan)

## 1. Objective (目标)
优化任务编辑页面中“起始转存点 (Start Checkpoint)”文件名的回显速度。通过将选定的文件名缓存至数据库，消除每次点击“编辑”时请求网盘接口解析分享链接带来的延迟，实现秒开体验。

## 2. Key Files & Context (核心影响文件)
- **后端模型**: `internal/db/db.go`
  - 需在 `Task` 结构体中新增字段以持久化文件名。由于使用了 GORM，数据库表结构会自动迁移（AutoMigrate）。
- **前端视图**: `web/src/views/Tasks.vue`
  - 需修改状态管理、保存逻辑及回显逻辑，移除原有的异步 `parseShareLink` 请求等待。

## 3. Implementation Steps (实施步骤)

### Step 3.1: 后端模型层更新
1. 打开 `internal/db/db.go`。
2. 在 `Task` 结构体中的 `StartFileID` 字段下方，新增 `StartFileName` 字段：
   ```go
   StartFileName string `gorm:"size:255" json:"start_file_name"` // 起始文件名称 (用于前端快速回显)
   ```

### Step 3.2: 前端数据模型与交互优化
1. 打开 `web/src/views/Tasks.vue`。
2. **初始化数据结构**: 在 `form` 的响应式对象中，追加 `start_file_name: ''` 属性。
3. **选择逻辑更新 (`confirmStartFileSelection`)**: 当用户在弹窗中确认选择了起始文件时，不仅保存其 ID，同时将文件名赋值给 `form.value.start_file_name`。
4. **重置逻辑更新 (`handleUrlChange` & `clearStartFile`)**: 当分享链接改变或用户点击清除时，同步清空 `form.value.start_file_name`。

### Step 3.3: 前端回显逻辑重构 (核心提速点)
1. 定位 `Tasks.vue` 中的 `handleEdit(row)` 函数。
2. **直接赋值回显**: 修改逻辑，若 `row.start_file_id` 和 `row.start_file_name` 存在，直接将 `row.start_file_name` 赋值给 `selectedStartFileName.value`。
3. **移除冗余请求**: 删除原有那段为了获取文件名而调用 `parseShareLink` 的 `try...catch` 异步代码块，彻底切断对后端及外部 API 的依赖。
4. **兼容性处理**: 对于历史存量任务（只有 ID 没有记录文件名），可以降级显示为 `ID: {id}`。

## 4. Verification & Testing (验证与测试)
- **数据库迁移测试**: 启动后端服务，观察 SQLite 数据库表 `tasks` 是否成功新增了 `start_file_name` 列。
- **保存测试**: 在界面上新建或编辑一个任务，选择起始转存点后保存，通过接口或直接查库确认 `start_file_name` 被正确持久化。
- **秒开测试**: 刷新页面，点击刚才保存的任务的“编辑”按钮，确认“起始转存点”输入框立刻显示文件名，且 Network 面板没有发起 `parseShareLink` 请求。
- **清除测试**: 更改分享链接或点击“清除选择”，确认相关状态被同步重置。