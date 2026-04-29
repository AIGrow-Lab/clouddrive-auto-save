# 升级 GitHub Actions 流水线以支持 Node.js 24 (Node.js 24 Upgrade Plan)

**Goal:** 消除 GitHub Actions 关于 Node.js 20 弃用的警告。通过全局设置环境变量 `FORCE_JAVASCRIPT_ACTIONS_TO_NODE24=true` 来让旧版 Actions 使用 Node.js 24 运行，同时将 `actions/setup-node` 中的 Node 版本同步升级至 24。

---

## Task 1: 升级 CI 流水线 (ci.yml & e2e.yml)

**Files:**

- Modify: `.github/workflows/ci.yml`
- Modify: `.github/workflows/e2e.yml`

**Changes:**

1. 在文件顶部（例如 `jobs:` 之上）添加全局环境变量 `env: FORCE_JAVASCRIPT_ACTIONS_TO_NODE24: true`。
2. 将 `actions/setup-node` 步骤中的 `node-version: '20'` 更改为 `node-version: '24'`。

## Task 2: 升级 Docker 发布流水线

**Files:**

- Modify: `.github/workflows/docker-publish-main.yml`
- Modify: `.github/workflows/docker-publish-tags.yml`

**Changes:**

1. 在文件顶部添加全局环境变量 `env: FORCE_JAVASCRIPT_ACTIONS_TO_NODE24: true`。

## Task 3: 升级 Release 流水线

**Files:**

- Modify: `.github/workflows/release.yml`

**Changes:**

1. 在文件顶部添加全局环境变量 `env: FORCE_JAVASCRIPT_ACTIONS_TO_NODE24: true`。
2. 将 `actions/setup-node` 步骤中的 `node-version: 20` 更改为 `node-version: 24`。

## Task 4: 提交代码

- 提交对所有的 Workflow 文件的更改并推送到远程仓库。
