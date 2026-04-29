# 发布 v1.0.0 版本计划

## 目标 (Objective)

通过给当前最新代码打上 `v1.0.0` 的 Git Tag 并推送到远程仓库，以触发配置好的 GitHub Actions 流水线，完成自动化构建、发布 Release 产物以及推送到 Docker Hub。

## 具体改动 (Changes)

1. 执行 `git tag v1.0.0`，给当前的提交打上发行版标签。
2. 执行 `git push origin v1.0.0`，将标签推送到远程仓库。

## 验证方式 (Verification)

推送成功后，可以在 GitHub 的 Actions 面板中查看正在执行的 Release 流程，确认发版流水线被成功触发。
