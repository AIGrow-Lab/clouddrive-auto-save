# 切换至文件级版本号维护计划 (Switch to File-Based Versioning Plan)

**Goal:** 摒弃复杂的 `git describe` 动态获取方式，改由根目录下的 `VERSION` 文件集中维护项目版本号，确保打包镜像和发布的名称更加干净可控。

---

### Task 1: 创建全局版本文件

**Files:**
- Create: `VERSION`

**Changes:**
- 在项目根目录创建 `VERSION` 文件，初始内容为 `1.0.0`。

### Task 2: 简化 Makefile 版本获取逻辑

**Files:**
- Modify: `Makefile`

**Changes:**
- 将 `VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "v1.0.0")` 替换为读取 `VERSION` 文件的逻辑：
  `VERSION ?= $(shell cat VERSION 2>/dev/null || echo "1.0.0")`。

### Task 3: 验证

- [ ] **Step 1: 运行 `make build-server`** 确认编译输出的版本号为干净的 `1.0.0`。
- [ ] **Step 2: 运行 `make docker-build`** 确认镜像标签为 `zcq98/clouddrive-auto-save:1.0.0`。
