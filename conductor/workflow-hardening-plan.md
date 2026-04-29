# GitHub Actions 流水线触发规则优化计划 (Workflow Hardening Plan)

**Goal:** 精简各个流水线的触发路径，移除不必要的重复运行，从而节约 CI 资源并加快发布反馈速度。

---

## Task 1: 精炼 ci.yml 触发规则

**Files:**

- Modify: `.github/workflows/ci.yml`

**Changes:**

1. 从 `paths` 列表中移除 `'**.md'`。
2. 理由：单元测试和安全扫描主要是针对 Go 代码的，纯文档改动不应运行这些耗时任务。

## Task 2: 完善 docker-publish-main.yml 触发规则

**Files:**

- Modify: `.github/workflows/docker-publish-main.yml`

**Changes:**

1. 在 `paths-ignore` 中增加以下忽略路径：
   - `- '**.md'` (替代原有的散落 md 文件)
   - `- 'e2e/**'`
2. 理由：E2E 测试脚本和文档绝对不会被打包进镜像，它们的改动不应触发耗时数分钟的多架构 Docker 构建。

## Task 3: 自动提交与验证

- 提交修改并推送到远程。
- 手工检查各流水线的触发逻辑是否符合预期。
