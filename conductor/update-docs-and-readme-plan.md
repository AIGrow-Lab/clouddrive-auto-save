# 修复 README 中 CI/CD 徽章链接失败的问题 (Fix CI/CD Badge Plan)

**Goal:** 更新 `README.md` 中的 GitHub Actions 状态徽章链接，因为之前将 `docker-publish.yml` 拆分成了 `docker-publish-main.yml` 和 `docker-publish-tags.yml`，导致原有的徽章图片地址失效。

---

### Task 1: 更新 README.md 中的徽章链接

**Files:**
- Modify: `README.md`

**Changes:**
1. 将 `[![CI/CD](https://github.com/zhaocongqi/clouddrive-auto-save/actions/workflows/docker-publish.yml/badge.svg)](https://github.com/zhaocongqi/clouddrive-auto-save/actions)` 替换为指向 `docker-publish-main.yml` 的徽章链接：`[![CI/CD](https://github.com/zhaocongqi/clouddrive-auto-save/actions/workflows/docker-publish-main.yml/badge.svg)](https://github.com/zhaocongqi/clouddrive-auto-save/actions)`