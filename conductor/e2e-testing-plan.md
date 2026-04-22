# E2E 端到端测试集成计划 (E2E Testing Implementation Plan)

**Goal:** 引入 Playwright 框架并构建 E2E 测试环境，配合后端的 E2E 专属 Mock 模式，彻底杜绝 AI (Vibecoding) 在后续开发中由于上下文丢失引发的核心功能退化。

---

### Task 1: 初始化 Playwright 测试环境

**Files:**
- Create: `e2e/` 目录及相关配置
- Modify: `package.json` (根目录或独立的 e2e package)

**Changes:**
1. 在项目根目录创建 `e2e` 文件夹，专门用于存放端到端测试。
2. 初始化 Playwright 项目配置 (`playwright.config.ts`)。
3. 编写基础的启动环境依赖（安装 `@playwright/test`）。

### Task 2: 后端支持 E2E Mock 启动模式

**Files:**
- Modify: `cmd/server/main.go`
- Modify: `internal/core/worker/mock_test.go` (如果需要复用 MockDriver，可能需要将其移出 `_test.go` 或专门为 E2E 写一个 Mock)

**Changes:**
1. 在 `main.go` 启动时读取环境变量 `E2E_TEST_MODE`。
2. 如果处于 E2E 模式：
   - 使用内存数据库 `file::memory:?cache=shared`，保证每次测试数据纯净。
   - 自动向数据库注入测试用的 Mock 账号（如平台为 `mock_139` 的账号）。
   - 注册 E2E 专用的 `MockDriver`（模拟目录树、分享列表和固定的转存成功响应）。
   - 暴露特殊的 API（如 `/api/e2e/reset`）供 Playwright 在每个用例间重置状态（可选）。

### Task 3: 编写首个核心 E2E 测试用例

**Files:**
- Create: `e2e/tests/core.spec.ts`

**Changes:**
1. 编写“账号全生命周期管理”测试：进入系统 -> 看到空状态引导 -> 点击添加账号 -> 绑定 Mock 账号 -> 验证列表展示。
2. 编写“任务创建与重命名预览”测试：输入链接 -> 输入正则 -> 验证预览表格中的 `matched` 状态和 `original_name` -> 保存任务。

### Task 4: 配置与验证

- 执行 Playwright 测试，确保无头浏览器能够顺利跑通上述流程。
- 将执行通过的代码提交至本地仓库。
