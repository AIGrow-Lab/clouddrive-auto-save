# 修复跨平台编译隐患计划 (Fix Cross-Compilation Issue)

**Goal:** 将底层 SQLite 驱动替换为纯 Go 版本，消除对 CGO 的依赖，从而保证 GoReleaser 能够顺利进行跨平台自动发布。

---

### Task 1: 替换 GORM SQLite 驱动

**Files:**
- Modify: `internal/db/db.go`

**Changes:**
- 将 `import "gorm.io/driver/sqlite"` 改为 `import "github.com/glebarez/sqlite"`。

### Task 2: 整理依赖并清理环境

**Files:**
- Modify: `go.mod`, `go.sum`
- Modify: `Dockerfile`

**Changes:**
- 执行 `go get github.com/glebarez/sqlite` 并运行 `go mod tidy`。
- (可选优化) 在 `Dockerfile` 中移除 `apk add --no-cache gcc musl-dev`。

### Task 3: 验证

- [ ] **Step 1: 运行 `CGO_ENABLED=0 go build -v ./cmd/server/main.go`** 确保在无 CGO 环境下编译成功。
