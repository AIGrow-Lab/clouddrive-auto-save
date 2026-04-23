# E2E 账号真实交互测试优化计划 (E2E Account Real Interaction Plan)

**Goal:** 摒弃在后端 `main.go` 中直接通过 SQL 插入测试账号的做法，改为由 Playwright 通过前端 UI 真实模拟用户的“添加账号”操作，从而实现对前后端表单校验、接口调用及 HTTP Mock 链路的 100% 真实覆盖。

---

### 问题分析 (Context)

当前的 `make e2e-test` 在启动后端服务时，会在 `main.go` 中通过 GORM 直接向数据库插入两个测试账号。
虽然这样能快速准备前置数据，但这破坏了 E2E 测试的完整性，导致前端的“添加账号弹窗”、“表单提交逻辑”、“后端的 `createAccount` 接口”未能被自动化测试覆盖。由于我们的 HTTP Mock 层已经能够完美模拟网盘接口的响应，让测试脚本通过 UI 创建账号是更符合 E2E 理念的最佳实践。

### Task 1: 移除后端的硬编码注入

**Files:**
- Modify: `cmd/server/main.go`

**Changes:**
1. 在 `isE2E` 的判断分支中，保留 `core.SetupE2EHTTPMock()` 开启 Mock 拦截器。
2. 删除 `db.DB.Create(&db.Account{...})` 注入测试账号的代码块。让 E2E 环境以一个干净的数据库启动。

### Task 2: 改造 E2E 测试脚本

**Files:**
- Modify: `e2e/tests/core.spec.ts`

**Changes:**
1. **重写“账号管理”测试用例**:
   - 验证初始的无账号状态（例如：验证页面中存在“立即绑定账号”按钮）。
   - **交互：添加 139 账号**:
     - 点击“立即绑定账号”或右上角的添加按钮。
     - 在平台选项中选择“移动云盘”。
     - 在 `Authorization` 文本框中输入 mock 数据 `mock_auth`。
     - 点击“确认添加”。
     - 断言成功提示（如“E2E测试账号(移动云盘)”出现）。
   - **交互：添加 Quark 账号**:
     - 再次打开添加账号弹窗。
     - 切换平台至“夸克网盘”。
     - 在 `Cookie 全量字符串` 文本框中输入 `mock_cookie`。
     - 点击“确认添加”。
     - 断言成功提示。
   - **保留原有的容量断言**: 
     - 在通过真实的 UI 添加后，依然断言界面上渲染出了 HTTP Mock 给出的 `512 GB / 1 TB` 等容量信息。

### Task 3: 运行验证与提交

- 执行 `pkill -9 ucas || true && make e2e-test` 观察 Playwright 能否顺利完成真实的表单提交流程。
- 提交代码。
