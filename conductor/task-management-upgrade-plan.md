# 任务管理模块升级实施计划 (Task Management Upgrade Implementation Plan)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 增强任务管理功能，支持网盘目录树形选择与新建、定时配置图形化、按文件日期过滤转存，以及基于真实分享链接的全量重命名效果预览。

**Architecture:** 前端使用 Element Plus 的 TreeSelect, DatePicker 和弹窗表格实现交互升级；后端在 Task 模型中增加 StartDate 字段，并新增三个 API（获取目录树、新建文件夹、效果预览）供前端调用；底层 Worker 在执行转存时增加按日期过滤和按需创建保存目录的逻辑。

**Tech Stack:** Go 1.23, Gin, GORM, SQLite3, Vue 3, Element Plus

---

### Task 1: 扩展后端数据库模型与数据迁移

**Files:**
- Modify: `internal/db/db.go`

- [ ] **Step 1: 在 Task 模型中添加时间过滤字段**
  - 在 `internal/db/db.go` 中，找到 `Task` 结构体。
  - 在 `Cron string` 字段下方添加 `StartDate *time.Time \`json:"start_date"\``（使用指针以支持 NULL 值，表示不进行日期过滤）。

- [ ] **Step 2: 重启后端服务以触发 GORM 的 AutoMigrate**
  - 因为 `cmd/server/main.go` 中调用了 `db.DB.AutoMigrate`，只需重新编译运行服务，SQLite 数据库即可自动增加 `start_date` 列。

### Task 2: 实现后端目录树形选择 API

**Files:**
- Modify: `internal/api/router.go`
- Create (or Modify): `internal/api/folder.go` (新建文件专门处理目录 API)

- [ ] **Step 1: 在 router.go 中注册新路由**
  - 在 `api` 路由组中添加 `GET /accounts/:id/folders`。
  - 在 `api` 路由组中添加 `POST /accounts/:id/folders`。

- [ ] **Step 2: 编写获取网盘目录树的接口函数**
  - 提取 `account_id`，查询 `db.Account` 并获取对应的 `CloudDrive` 驱动实例。
  - 接收 Query 参数 `parent_id` (默认为根目录 `"/"` 或云盘根节点 ID)。
  - 调用驱动的类似 `ListFiles(parent_id)` 方法获取目录列表。
  - 将结果转换为前端 TreeSelect 需要的格式 `[{ id: "x", label: "文件夹名", isLeaf: false }]`，并返回 JSON。

- [ ] **Step 3: 编写新建网盘文件夹的接口函数**
  - 提取 `account_id`，获取驱动实例。
  - 接收 JSON Payload `{ "parent_id": "...", "name": "新文件夹" }`。
  - 调用驱动的类似 `CreateFolder(parent_id, name)` 方法。
  - 返回新建成功后的文件夹信息 `[{ id: "new_id", label: "新文件夹", isLeaf: false }]`。

### Task 3: 实现后端全量重命名预览 API

**Files:**
- Modify: `internal/api/router.go`
- Create (or Modify): `internal/api/task_preview.go`

- [ ] **Step 1: 在 router.go 中注册预览路由**
  - 在 `api` 路由组中添加 `POST /tasks/preview`。

- [ ] **Step 2: 定义预览请求的数据结构**
  - 定义一个结构体接收前端发来的临时任务数据：包含 `account_id`, `share_url`, `extract_code`, `start_date`, `pattern`, `replacement`。

- [ ] **Step 3: 编写重命名预览的核心逻辑**
  - 根据 `account_id` 获取对应的云盘驱动实例。
  - 调用驱动的类似 `ParseShare(share_url, extract_code)` 方法，获取分享链接内的全量文件列表信息（包含文件名和更新时间）。
  - 遍历文件列表：
    - 如果 `start_date` 不为空且文件的更新时间早于 `start_date`，则跳过（表示被过滤）。
    - 对剩余的文件，调用现有的 `renamer` 引擎，使用 `pattern` 和 `replacement` 生成新文件名。
  - 将每个文件的原始名称、新名称和匹配状态打包成切片，通过 JSON 响应给前端。

### Task 4: 升级底层转存引擎 (Worker)

**Files:**
- Modify: `internal/core/worker/worker.go`

- [ ] **Step 1: 在转存执行逻辑中增加日期过滤**
  - 找到遍历分享链接文件并准备发起转存 API 的循环处。
  - 增加判断：`if task.StartDate != nil && file.UpdateTime.Before(*task.StartDate) { continue }`，跳过过期文件。

- [ ] **Step 2: 增加目标目录按需创建逻辑**
  - 在正式发起文件的转存/保存请求前，检查 `task.SavePath`。
  - 调用云盘驱动的接口校验该路径是否存在。如果不存在，则调用创建文件夹接口自动创建它。

### Task 5: 前端 UI 重构与组件升级

**Files:**
- Modify: `web/src/views/Tasks.vue`
- Modify: `web/src/api/task.js`
- Modify: `web/src/api/account.js` (若新增目录接口放此处)

- [ ] **Step 1: 在 api 文件中添加新接口绑定**
  - `getFolders(accountId, parentId)`
  - `createFolder(accountId, parentId, name)`
  - `previewTask(previewData)`

- [ ] **Step 2: 将原文本框替换为树形选择器与日期选择器**
  - 在 `Tasks.vue` 的创建对话框中，找到“保存路径”字段。将其替换为 `<el-tree-select v-model="form.save_path" lazy :load="loadFolders" />`。
  - 编写 `loadFolders` 方法，调用后端的 `/accounts/:id/folders` 实现懒加载。
  - 在表单的定时配置区域，增加下拉框，用于选择预设的图形化定时规则（例如：不使用、每天、每周），根据选择同步生成底层的 Cron 字符串赋值给 `form.cron`。
  - 增加 `<el-date-picker v-model="form.start_date" type="datetime" placeholder="选择仅转存此时间之后的文件" />`。

- [ ] **Step 3: 添加全量预览表格与交互逻辑**
  - 在表单最下方添加一个“全量效果预览”按钮。
  - 点击按钮触发后端 `previewTask` 接口，展示 Loading 状态。
  - 使用 `<el-dialog>` 或在表单底部展开一个区域，嵌入 `<el-table :data="previewData">`，展示“原始文件名”、“重命名后名称”及是否匹配成功等信息。
  - 用户确认预览效果无误后，再点击底部的“确认并保存任务”按钮。
