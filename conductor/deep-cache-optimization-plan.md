# GitHub Actions 深度缓存与资源节约优化计划 (Deep Cache Optimization Plan)

**Goal:** 进一步挖掘 CI 流水线中的潜在性能瓶颈，完善多目录依赖缓存，并剔除流水线中不必要的运行环境依赖，挑战“零冗余”的极限执行速度。

---

## 问题分析 (Context)

您刚才在编辑器中可能看到了尚未自动刷新（或本地旧版本）的内容，实际上，在上一个提交中，我们已经成功引入了 `actions/cache@v4` 来专门缓存 100+ MB 的 Playwright Chromium 浏览器，避免了重复下载的漫长等待。

但在对所有流水线进行深度诊断后，我发现了两处隐藏的“资源浪费点”可以进一步利用缓存和精简来加速：

1. **`ci.yml` 的“空转”环境**:
   由于我们在早期的优化中，把依赖 Node.js 的 Markdown 格式检查 (`lint-md`) 从基础的 `make check` 中剥离了出去，现在的 `make check` 是一个纯纯的 Go 语言验证流程（跑 `gofmt`, `go vet`, `go test`）。但是，`ci.yml` 中仍然残留了 `actions/setup-node@v4` 步骤，每次都会白白耗费几秒钟去下载配置 Node 环境和恢复 NPM 缓存，这完全是无用功。
2. **`e2e.yml` 缓存覆盖不全**:
   我们的 E2E 测试其实包含了两个 Node.js 项目：前端的 `web` 和测试专用的 `e2e`。
   目前 `actions/setup-node@v4` 的 `cache-dependency-path` 仅配置了 `web/package-lock.json`。这意味着 GitHub Actions 在进行 NPM 全局缓存校验时，会“视而不见” `e2e` 目录下的依赖变更，导致缓存命中率和恢复效率打折扣。

## Task 1: 精简 ci.yml (移除无用的 Node 环境)

**Files:**

- Modify: `.github/workflows/ci.yml`

**Changes:**

1. 直接删除 `Set up Node.js` 整个步骤（因为后续步骤均不依赖前端编译或 npm 环境）。

## Task 2: 完善 e2e.yml 的双目录依赖缓存

**Files:**

- Modify: `.github/workflows/e2e.yml`

**Changes:**

1. 将 `actions/setup-node@v4` 步骤中的 `cache-dependency-path` 改造为多路径格式，同时监控前后端两份锁文件：

   ```yaml
          cache-dependency-path: |
            web/package-lock.json
            e2e/package-lock.json
   ```

## Task 3: 自动提交并验证

- 提交代码更改并推送到远程，观察 CI 流水线（尤其是基础代码校验）的运行耗时是否进一步缩短。
