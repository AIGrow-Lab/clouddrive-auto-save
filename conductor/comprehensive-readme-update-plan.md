# README 全量更新计划 (Comprehensive README Update Plan)

**Goal:** 全面检查并更新 `README.md`，使其准确反映近期项目引入的各项重要能力（异构镜像支持、GoReleaser 自动发布等），并补全缺失的环境变量说明。

---

### Task 1: 更新特性描述与配置说明

**Files:**
- Modify: `README.md`

**Changes:**
1. **核心特性 - 跨平台兼容**: 补充说明现在支持通过 GitHub Releases 直接下载各平台（Windows, Mac, Linux）预编译好的独立二进制程序。
2. **核心特性 - 容器化优先**: 明确标注官方 Docker 镜像现已支持 **多架构 / 异构镜像 (amd64 / arm64)**，支持树莓派、Mac M系列等 ARM 设备部署。
3. **技术架构 - CI/CD**: 将“自动构建 amd64 镜像”的描述更新为包含“多架构(amd64/arm64)镜像”以及“GoReleaser 全自动跨平台编译发布”。
4. **配置说明 (Configuration)**: 补充缺失的 `LISTEN_ADDR`（API 监听地址与端口）以及常用的 `TZ`（时区）环境变量。

### Task 2: 补充“快速开始”下载方式

**Files:**
- Modify: `README.md`

**Changes:**
1. 在“快速开始”中增加“方式三：直接下载预编译程序”，引导用户前往 GitHub Releases 页面下载由 GoReleaser 自动打包的产物。