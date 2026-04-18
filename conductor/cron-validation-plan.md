# 优化 Cron 输入校验计划

**目标：** 通过在前端提供引导提示以及在后端增加 API 拦截，验证自定义 Cron 表达式，提升用户体验和系统稳定性。

**修改内容：**

### 1. 后端逻辑 (`internal/core/scheduler/scheduler.go`)
- 实现 `ValidateCron(cronExpr string) error` 函数。通过临时实例化 `cron.New(cron.WithSeconds())` 并尝试添加该表达式来捕获解析错误。

### 2. 后端 API (`internal/api/router.go`)
- 在创建、更新任务及修改全局设置时，调用该校验函数。如果格式非法，返回 `400 Bad Request` 并提供具体的错误描述。

### 3. 前端 UI (`web/src/views/Tasks.vue`)
- 为 Cron 输入框增加清晰的占位符（如 `秒 分 时 日 月 周`）。
- 增加 Tooltip 提示，解释 6 位表达式的格式，极大降低使用门槛。
