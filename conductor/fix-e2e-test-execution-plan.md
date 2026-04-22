# 修复 E2E 测试指令执行与挂起问题计划 (Fix E2E Test Execution Plan)

**Goal:** 解决 `make e2e-test` 指令无法正常执行以及测试失败后终端挂起（一直 serving HTML report）的问题。

---

### 问题分析 (Root Cause)

1. **无法正常执行**: 在 `Makefile` 中启动后端服务 `E2E_TEST_MODE=true ./bin/ucas &` 之后，没有等待服务完全启动就立即执行了 Playwright 测试，导致前端页面请求 `http://localhost:8080` 被拒绝连接。
2. **终端挂起 (Hang)**: Playwright 默认的 `html` 报告器在非 CI 环境下，一旦测试失败，会自动在本地启动一个 HTTP 服务器并打开浏览器展示报告，这会导致 `make` 进程阻塞，无法执行后续的 `kill $PID` 清理逻辑。

### Task 1: 优化 Makefile 中的 e2e-test 指令

**Files:**
- Modify: `Makefile`

**Changes:**
1. 在启动后端服务之后、执行测试之前，加入 `sleep 2;`，给 Gin 框架和数据库初始化留出充足的时间绑定端口。
2. 在运行 Playwright 测试的指令前加上 `CI=true` 环境变量（即 `CI=true npx playwright test`）。Playwright 在检测到 CI 环境时，会生成报告但**不会自动启动服务器挂起终端**，保证流水线能够继续往下走，完成清理工作。

### Task 2: 验证效果

- 运行 `make e2e-test`，观察是否能正常等待 2 秒后开始测试。
- 观察测试失败或成功后，终端是否能顺利退出并执行 `Cleaning up backend` 逻辑。
