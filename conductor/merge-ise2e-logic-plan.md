# 合并 isE2E 逻辑块计划 (Merge isE2E Logic Plan)

**Goal:** 优化 `cmd/server/main.go` 中初始化逻辑，将两处 `if isE2E` 代码块合并为一处，提升代码整洁度。

---

### 问题分析 (Context)
此前 E2E 模式初始化被分为两部分，是因为其中包含 `db.DB.Create()` 注入数据库测试账号的逻辑，所以必须等 `db.InitDB` 执行完后才能继续。
由于现在我们将 E2E 模式改为“完全从前端真实点击并走 HTTP Mock 进行全链路录入”，数据库注入代码已被删除。剩下的 `core.SetupE2EHTTPMock()` 只是全局 HTTP 拦截器赋值，完全可以和数据库初始化前的 `isE2E` 环境设置部分合并。

### Task 1: 合并 main.go 中的 isE2E 逻辑

**Files:**
- Modify: `cmd/server/main.go`

**Changes:**
1. 找到初始化阶段关于 `isE2E` 的判断。
2. 将 `core.SetupE2EHTTPMock()` 的调用上移到第一个 `if isE2E` 分支中。
3. 删除原来位于 `InitDB` 后面的第二个 `if isE2E` 代码块。

### Task 2: 提交代码

- 自动提交修改并推送到远端仓库。