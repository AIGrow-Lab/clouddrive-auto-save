# Bark 消息推送与设置页面重构计划 (Bark Notification & Settings Refactor)

## 1. 背景与目标 (Background & Objective)

当前系统缺乏自动化转存完成后的即时通知机制。为了提升用户体验，本项目计划添加基于 **Bark** 的消息通知模块，并在 Web 管理后台中新增一个专门的“系统设置”页面，用于统一管理全局定时任务、Bark 推送配置等。

## 2. 详细实施步骤 (Detailed Implementation Steps)

### 第一阶段：后端通知引擎开发 (Backend Implementation)

1. **数据库配置扩展**:
    * 在 `Setting` 表中引入以下配置项：
        * `bark_enabled`: 推送开关 ("true"/"false")
        * `bark_server`: Bark 服务器地址 (默认为 `https://api.day.app`)
        * `bark_device_key`: Bark 设备 Key
        * `bark_notify_success`: 成功是否推送
        * `bark_notify_failed`: 失败是否推送
2. **开发 `internal/core/notify` 模块**:
    * 实现 `bark.go`，封装对 Bark API 的调用。
    * 支持 POST 方式推送到 `push` 接口，以便携带更多元数据（如 `group`, `icon`, `url`）。
3. **Worker 逻辑改造 (`internal/core/worker/worker.go`)**:
    * 在 `SaveLink` 和正则重命名过程中，记录成功转存的文件名列表。
    * 改造 `finishTask` 函数或新增通知钩子，在任务进入最终状态时，根据配置触发 Bark 推送。
    * 推送消息主体结构：
        * 标题：`转存任务完成: {TaskName}`
        * 正文：列出转存成功的文件名（支持截断防止超长），显示成功/失败总数。

### 第二阶段：API 与 设置系统重构 (API & Settings Refactor)

1. **后端 API 升级 (`internal/api/router.go`)**:
    * 新增 `GET /api/settings` 和 `POST /api/settings`，支持以 KV 对象形式批量读取和保存全局设置。
    * 移除旧的 `getScheduleSettings` / `updateScheduleSettings` 路由（或保持兼容）。
2. **前端路由与菜单更新**:
    * `web/src/router/index.js`: 新增 `/settings` 路由，关联 `views/Settings.vue`。
    * `web/src/layout/MainLayout.vue`: 在侧边栏增加“系统设置”菜单项。

### 第三阶段：前端设置页面开发 (Frontend Implementation)

1. **创建 `web/src/views/Settings.vue`**:
    * **全局调度卡片**: 迁移原 `Tasks.vue` 顶部的定时开关和 Cron 编辑器。
    * **Bark 推送卡片**:
        * 推送总开关。
        * 服务器地址、Key、推送等级配置。
        * 增加“发送测试消息”按钮。
2. **清理 `web/src/views/Tasks.vue`**:
    * 移除页面顶部的全局定时设置卡片，使任务列表页面更加纯粹。

## 3. 验证与测试 (Verification & Testing)

1. **后端单元测试**: 模拟 Bark Server，验证推送请求的格式（标题、正文及文件列表）。
2. **联调测试**:
    * 在 Web 界面修改 Bark 配置并保存，检查数据库存储。
    * 点击“发送测试消息”，手机端确认收到推送。
    * 运行一个真实转存任务，确认任务结束后收到包含文件名的推送。

## 4. 迁移建议

由于配置项存储在 `Setting` 表中，新版本将自动向下兼容。前端界面的迁移对用户来说是无感的功能升级。
