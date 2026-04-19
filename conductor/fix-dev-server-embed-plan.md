# 修复本地开发编译失败计划 (Fix Local Dev Build Failure)

**Goal:** 解决集成 `go:embed` 后导致 `make dev-server` 无法运行的问题。

---

### Task 1: 自动化目录占位 (Makefile Automation)

**Files:**
- Modify: `Makefile`

**Changes:**
- 在 `build-server` 和 `dev-server` 目标中增加 `mkdir -p internal/api/dist && touch internal/api/dist/.gitkeep` 逻辑。

### Task 2: 仓库清理与规范 (Git Cleanup)

**Files:**
- Modify: `.gitignore`

**Changes:**
- 确保 `internal/api/dist/` 下的所有真实内容被忽略，但保留占位文件（可选）或直接忽略整个目录。

### Task 3: 验证

- [ ] **Step 1: 运行 `make clean`** 清理环境。
- [ ] **Step 2: 运行 `make dev-server`** 观察是否能成功启动。
