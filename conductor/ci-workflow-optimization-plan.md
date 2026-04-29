# 整合 E2E 测试到 GitHub Actions CI 流水线计划 (CI Workflow E2E Integration Plan)

**Goal:** 新增 `.github/workflows/e2e.yml`，在每次提交代码或拉取合并请求时独立运行全链路 E2E 测试，并在失败时自动上传 Playwright 报告。

---

## 问题分析 (Context)

当前的 `ci.yml` 工作流主要负责运行基础的安全扫描 (`govulncheck`) 和单元测试 (`make check`)。为了使流水线职责更加清晰，且让耗时较长的 E2E 测试能与单元测试并行运行，新建一个专门的 `e2e.yml` 工作流是最佳实践。

## Task 1: 创建独立的 E2E 流水线文件

**Files:**

- Create: `.github/workflows/e2e.yml`

**Changes:**

1. 监听 `main` 分支的 `push` 和 `pull_request`，并指定只在代码或测试相关目录发生变动时触发：
   - `'**.go'`, `'go.mod'`, `'go.sum'`, `'Makefile'`
   - `'web/**'`
   - `'e2e/**'`
   - `'.github/workflows/e2e.yml'`
2. 配置 `jobs.e2e-test`:
   - 依赖 Node.js 20 和 Go 1.25 环境（利用官方 Action 提供缓存支持）。
   - 运行 `make e2e-setup` 安装 Playwright 核心依赖。
   - 运行 `npx playwright install-deps chromium` 补全 Ubuntu 上的系统级浏览器依赖。
   - 运行 `make e2e-test` 执行端到端测试。
   - 配置 `actions/upload-artifact@v4`，在测试结束后上传 `e2e/playwright-report/` 目录，保留 7 天以便排错。

## Task 2: 本地提交代码

- 提交新增的 `e2e.yml`。
- 自动推送以触发 GitHub Actions。
