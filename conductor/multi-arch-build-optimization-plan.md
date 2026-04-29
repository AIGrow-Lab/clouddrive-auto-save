# 多架构镜像构建优化计划 (Multi-Arch Build Optimization Plan)

**Goal:** 优化 `Dockerfile`，利用 Docker Buildx 的原生交叉编译能力取代 QEMU 指令级模拟，将跨架构镜像（特别是 ARM64）的构建时间从数十分钟缩短至几分钟。

---

## Task 1: 优化前端构建阶段

**Files:**

- Modify: `Dockerfile`

**Changes:**

1. 在 `web-builder` 阶段，通过 `FROM --platform=$BUILDPLATFORM node:20-alpine AS web-builder` 固定使用构建机原生架构（通常是 amd64）。
2. 原因：前端静态资源的构建结果与架构无关，强制使用原生架构可以利用 CPU 的最大算力全速运行 `npm install` 和 `npm run build`。

## Task 2: 优化后端编译阶段

**Files:**

- Modify: `Dockerfile`

**Changes:**

1. 在 `server-builder` 阶段，通过 `FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS server-builder` 固定使用原生架构。
2. 引入 Buildx 自动注入的环境变量 `ARG TARGETOS` 和 `ARG TARGETARCH`。
3. 在最终编译命令中加入交叉编译环境变量：`CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build ...`。
4. 原因：Go 语言拥有极其强悍的原生交叉编译能力（且本项目完全去除了 CGO 依赖），因此在 x86 环境下直接编译 ARM64 的二进制文件速度极快，无需使用低效的 QEMU 模拟器。

## Task 3: 开启 GitHub Actions Docker 缓存

**Files:**

- Modify: `.github/workflows/docker-publish-main.yml`
- Modify: `.github/workflows/docker-publish-tags.yml`

**Changes:**

1. 在两个工作流的 `docker/build-push-action@v5` 步骤中增加 GHA (GitHub Actions) 缓存配置：

   ```yaml
   cache-from: type=gha
   cache-to: type=gha,mode=max
   ```

2. 原因：利用 GitHub 官方提供的缓存空间持久化镜像层（例如 `npm install` 下载的依赖包），使得后续构建可以直接复用缓存，极大地缩短流水线总耗时。

## Task 4: 自动提交与验证

- 执行提交，推送到远程仓库。
- 触发 GitHub Actions 观察构建耗时及缓存命中情况。
