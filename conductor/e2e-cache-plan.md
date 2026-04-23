# E2E 缓存优化计划 (E2E Cache Optimization Plan)

**Goal:** 优化 GitHub Actions 的 `e2e.yml` 流水线，利用 `actions/cache` 缓存 Playwright 下载的无头浏览器二进制文件（约 100+ MB），从而避免每次运行都重复下载，大幅缩短 CI 执行时间并节约网络资源。

---

### 问题分析 (Context)

当前 `e2e.yml` 在每次触发时都会执行 `make e2e-setup`，这会拉取最新的 Playwright Chromium 浏览器，这通常需要消耗 10-30 秒不等的时间。
我们可以使用 GitHub Actions 官方的缓存机制，以 Playwright 的版本号作为 Cache Key，将浏览器缓存目录 `~/.cache/ms-playwright` 存储起来。如果命中缓存，则跳过浏览器下载步骤。

### Task 1: 优化 E2E 流水线的依赖安装步骤

**Files:**
- Modify: `.github/workflows/e2e.yml`

**Changes:**
1. 将原有的 `Install E2E dependencies` 步骤拆分为更精细的步骤：
   - **安装 Node 依赖**: 运行 `cd e2e && npm install`。
   - **获取 Playwright 版本**: 通过 `npx playwright --version` 动态获取当前版本号，作为缓存的 Key。
   - **缓存浏览器**: 引入 `actions/cache@v4`，缓存路径为 `~/.cache/ms-playwright`，Key 使用 `${{ runner.os }}-playwright-${{ env.VERSION }}`。
   - **安装浏览器**: 仅在 `steps.playwright-cache.outputs.cache-hit != 'true'` (未命中缓存) 时，执行 `npx playwright install chromium`。
   - **安装系统依赖**: 运行 `npx playwright install-deps chromium` (OS 级的依赖使用 apt，通常很快且依赖 GitHub Runner 的状态，不建议硬缓存)。

### Task 2: 自动提交与验证

- 将修改提交至本地并推送到远程仓库。
- 让 GitHub Actions 触发一次构建以建立初始缓存，后续提交即可命中缓存。
