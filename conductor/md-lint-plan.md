# Markdown 格式校验集成计划 (Markdown Lint Plan)

**Goal:** 引入 Markdown 格式校验工具，并在 `Makefile` 和 `ci.yml` 中强化执行，确保代码库中的所有 `.md` 文档均符合书写规范。

---

### Task 1: 在 Makefile 中增加 `lint-md` 目标

**Files:**
- Modify: `Makefile`

**Changes:**
1. 添加 `lint-md` 目标，利用 `npx markdownlint-cli` 工具校验全目录下的 Markdown 文件，同时忽略 `node_modules` 和 `web/node_modules`。
2. 将 `lint-md` 加入到现有的 `check` 完整验证流程中（例如：`check: lint lint-md vet test`）。

### Task 2: 强化 CI 流水线 (ci.yml)

**Files:**
- Modify: `.github/workflows/ci.yml`

**Changes:**
1. **触发条件**: 在 `paths` 过滤规则中增加 `**.md`，使得任何 Markdown 文件的变动都能触发 `make check`。
2. **环境依赖**: 在 `Run checks` 步骤前，增加 `actions/setup-node@v4`，以提供执行 `npx` 命令所需的 Node.js 环境。

### Task 3: 自动提交与推送

- 提交代码更改。
- 不推送到远程。