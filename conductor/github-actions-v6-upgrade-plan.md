# GitHub Actions 版本全面升级计划 (GitHub Actions V6 Upgrade Plan)

**Goal:** 将项目中所有官方 GitHub Actions 从旧版 (v4/v5) 升级到最新的 **v6** 版本，以原生支持 Node.js 24 运行时，并彻底移除临时的环境变量补丁。

---

## Task 1: 升级所有流水线中的 Actions 版本

**涉及 Actions 映射表:**

- `actions/checkout@v4` -> `actions/checkout@v6`
- `actions/setup-go@v5` -> `actions/setup-go@v6`
- `actions/setup-node@v4` -> `actions/setup-node@v6`
- `actions/cache@v4` -> `actions/cache@v6`
- `actions/upload-artifact@v4` -> `actions/upload-artifact@v6` (注意: v6 可能与 v4 有一些配置差异，需检查)

**Files:**

- Modify: `.github/workflows/ci.yml`
- Modify: `.github/workflows/e2e.yml`
- Modify: `.github/workflows/docker-publish-main.yml`
- Modify: `.github/workflows/docker-publish-tags.yml`
- Modify: `.github/workflows/release.yml`

## Task 2: 移除环境变量补丁

**Files:**

- Modify: 所有上述 5 个 Workflow 文件

**Changes:**

1. 移除文件顶部的 `env: FORCE_JAVASCRIPT_ACTIONS_TO_NODE24: true`。
2. 理由：升级到 v6 后，Actions 会原生使用 Node 24 运行，不再需要强制转换。

## Task 3: 验证与提交

- 提交更改并推送到远程，观察警告是否消失。
