# 任务调度实施计划

**目标：** 利用 `robfig/cron/v3` 实现“全局默认 + 局部重写”模式的任务调度系统和全局开关。

**架构：**
- 使用新的 `Setting` 模型持久化 `global_schedule_enabled` 和 `global_schedule_cron`。
- `Task` 模型更新 `ScheduleMode` 字段（`global`, `custom`, `off`）。
- 调度器维护一个全局 Entry（触发所有“跟随全局”的任务）和多个独立 Entry（用于“自定义”任务）。
- 前端包含全局设置面板，并为每个任务提供调度模式选择。

**技术栈：** Go, Gin, GORM, `robfig/cron/v3`, Vue 3, Element Plus.

---

### 任务 1：更新数据库模型

**涉及文件：**
- 修改：`internal/db/db.go`

- [x] **步骤 1：为 Task 增加 ScheduleMode**
在 `Task` 结构体中添加 `ScheduleMode` 字段。保留现有的 `Cron` 字段用于存储自定义表达式。
- [x] **步骤 2：确保 Setting 结构体存在**
检查 `db.go` 中是否存在 `Setting` 结构体，并确保其已加入 `AutoMigrate` 自动迁移。

### 任务 2：核心调度器更新

**涉及文件：**
- 修改：`internal/core/scheduler/scheduler.go`

- [x] **步骤 1：更新 Scheduler 结构体**
更新结构体以支持 `CustomEntryIDs` 映射和单个 `GlobalEntryID`。
- [x] **步骤 2：实现全局与自定义更新逻辑**
实现 `UpdateGlobalSchedule` 和 `UpdateTask` 方法。更新 `RemoveTask` 以准确清理自定义条目。

### 任务 3：API 层集成

**涉及文件：**
- 修改：`internal/api/router.go`
- 修改：`cmd/server/main.go`

- [x] **步骤 1：添加系统设置接口**
新增 `GET /api/settings/schedule` 和 `POST /api/settings/schedule` 接口。
- [x] **步骤 2：更新任务相关接口**
更新创建和更新任务的接口，在保存后同步调用调度器的更新方法。
- [x] **步骤 3：生命周期管理**
在 `main.go` 启动时，从数据库加载全局设置并初始化所有任务的调度。

### 任务 4：前端 UI 更新

**涉及文件：**
- 修改：`web/src/views/Tasks.vue`
- 修改：`web/src/api/task.js`

- [x] **步骤 1：添加全局调度设置**
在页面顶部增加全局调度配置卡片（包含开关和 Cron 输入）。
- [x] **步骤 2：更新任务编辑弹窗**
将调度配置改为单选组（跟随全局、自定义、关闭），并适配相应的输入框显示逻辑。
- [x] **步骤 3：更新表格视图**
在表格列中直观展示任务的调度状态和具体频率。

### 任务 5：构建与验证

- [x] **步骤 1：** 运行 Go 单元测试。
- [x] **步骤 2：** 编译后端并运行，确保数据库迁移成功。
- [x] **步骤 3：** 运行前端，测试全局开关和单个任务重写逻辑。
