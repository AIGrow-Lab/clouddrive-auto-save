# 分享链接起始文件过滤机制重构计划 (Share File Start Point Refactoring Plan)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 重构任务的起始过滤逻辑，从手动填写“起始时间”改为输入分享链接后自动解析，用户在可视化的文件列表中直接勾选“起始文件”（使用 `start_file_id` 记录）。

**Architecture:** 
1. **数据库扩展**：在 `Task` 模型中新增 `StartFileID string` 字段，用于存储作为起始点的文件 ID。
2. **新 API 接口**：新增 `/api/tasks/parse_share`，仅用于根据链接和提取码获取真实的分享文件列表（按时间倒序）。
3. **前端重构**：移除 `start_date` 的时间选择器。当填写分享链接并选择了账号后，防抖触发 `parse_share` 接口，渲染出一个文件列表表格。用户可以在表格的行内操作中指定某文件为起始文件。
4. **Worker 引擎优化**：转存遍历文件时（由新到旧），若遇到与 `StartFileID` 匹配的文件则立刻 `break`，从而实现精确的“断点续传”。

**Tech Stack:** Go 1.23, Gin, GORM, SQLite3, Vue 3, Element Plus

---

### Task 1: 扩展数据库模型与迁移

**Files:**
- Modify: `internal/db/db.go`

- [ ] **Step 1: 在 Task 模型中添加 StartFileID 字段**
  在 `internal/db/db.go` 中，找到 `Task` 结构体。在 `StartDate` 字段下方添加 `StartFileID string \`gorm:"size:255" json:"start_file_id"\``。

### Task 2: 实现分享链接解析 API

**Files:**
- Modify: `internal/api/router.go`
- Modify: `internal/api/task_preview.go` (或新建 `share_parse.go`)

- [ ] **Step 1: 在 router.go 中注册新路由**
  在 `api` 路由组中添加 `POST /tasks/parse_share`。

- [ ] **Step 2: 编写解析分享链接的接口逻辑**
  在 `internal/api/task_preview.go` 中实现 `parseShareLinkInfo` 函数：
  - 接收包含 `account_id`, `share_url`, `extract_code` 的 JSON Payload。
  - 获取对应的云盘驱动实例。
  - 调用 `driver.ParseShare` 获取 `[]core.FileInfo`。
  - 直接返回 JSON 格式的文件列表。

### Task 3: 改造 Worker 引擎过滤逻辑

**Files:**
- Modify: `internal/core/worker/worker.go`

- [ ] **Step 1: 在 Worker 中应用 StartFileID 截断逻辑**
  在 `internal/core/worker/worker.go` 的 `execute` 方法中：
  - 遍历 `files` 时，原有的 `StartDate` 过滤可以保留作为兼容。
  - 增加对 `StartFileID` 的判断：如果 `task.StartFileID != ""` 且 `f.ID == task.StartFileID`，则跳出循环 (`break`)。
  - **注意**：由于获取到的文件列表默认是由新到旧排序的，当遇到 `StartFileID` 时，说明其后面的文件更老，应当停止处理并进行转存。

### Task 4: 前端视图与交互重构

**Files:**
- Modify: `web/src/api/task.js`
- Modify: `web/src/views/Tasks.vue`

- [ ] **Step 1: 新增 API 绑定**
  在 `web/src/api/task.js` 中增加 `parseShareLink(data)` 的请求绑定，对应 `/tasks/parse_share`。

- [ ] **Step 2: 移除起始日期选择器并添加自动解析逻辑**
  在 `Tasks.vue` 中：
  - 移除 `<el-date-picker v-model="form.start_date" ...>` 相关的模板。
  - 在 `form` 对象中移除 `start_date`，新增 `start_file_id: ''`。
  - 增加响应式变量 `shareFiles = ref([])` 和 `parsingShare = ref(false)`。
  - 对 `form.share_url`, `form.extract_code`, `form.account_id` 添加 `watch` 监听（配合 `lodash-es` 的 `debounce` 防抖）。当这三个字段有效时，自动调用 `parseShareLink` 获取文件列表。

- [ ] **Step 3: 增加可视化文件选择表格**
  在表单的“分享链接”下方（或在“整理规则”之前），增加一个表格区块展示 `shareFiles`：
  - 使用 `<el-table :data="shareFiles" max-height="300">` 展示解析出的文件列表（文件名、更新时间）。
  - 增加一列“设为起始点”操作列，或者直接使用单选框 (`el-radio` 绑定 `form.start_file_id`)。
  - 当某行被选中时，高亮该行。

- [ ] **Step 4: 测试与验证交互**
  - 重启前后端服务。
  - 在页面上填写一个含有多个文件的分享链接，验证文件列表是否能自动加载出来。
  - 点击某行的单选框，验证 `start_file_id` 是否正确绑定。
  - 保存任务并点击运行，验证后台日志是否确实在遇到该文件时 `break`，并且仅转存了它之前的文件。