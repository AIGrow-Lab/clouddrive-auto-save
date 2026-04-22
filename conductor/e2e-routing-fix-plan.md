# E2E 路由导航修复计划 (E2E Routing Fix Plan)

**Goal:** 修复 Playwright E2E 测试脚本中错误的路由导航方式，使其适配当前前端项目实际采用的 HTML5 History 模式。

---

### 问题分析 (Root Cause)

根据测试报错日志和页面快照 (`Error Context`)，当 Playwright 执行 `await page.goto('/#/accounts')` 时，页面实际渲染的是“仪表盘 (Dashboard)”组件。
检查 `web/src/router/index.js` 发现，项目前端使用的路由模式是 `createWebHistory`，而非 Hash 模式 (`createWebHashHistory`)。
因此，带有 `/#/` 的 URL 会被 Vue Router 直接忽略 Hash 部分，从而默认匹配到根路径 `/`（即 Dashboard 组件）。这导致后续所有针对账号页面的断言（如寻找账号列表元素）全部因为“元素未找到”而失败。

### Task 1: 修正 E2E 测试脚本的导航路径

**Files:**
- Modify: `e2e/tests/core.spec.ts`

**Changes:**
1. 将账号管理测试中的导航路径从 `/#/accounts` 更改为 `/accounts`。
2. 将任务管理测试中的导航路径从 `/#/tasks` 更改为 `/tasks`。

### Task 2: 验证与提交

- 运行测试验证修改效果。
- （先不执行本地提交）