# Docker 镜像自动发布流水线计划 (Docker Publish Pipeline Plan)

**Goal:** 移除 `Dockerfile` 中的国内代理配置以适应国际环境网络；优化 `Makefile` 增加镜像推送能力；并创建 GitHub Action，实现在推送 `main` 分支时发布 `latest` 标签，打 `tag` 时发布对应的版本标签。

---

### Task 1: 清理 Dockerfile 代理配置

**Files:**
- Modify: `Dockerfile`

**Changes:**
1. **移除 npm 代理**: 删除 `RUN npm config set registry https://registry.npmmirror.com`。
2. **移除 Alpine 代理**: 删除 `RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.tuna.tsinghua.edu.cn/g' /etc/apk/repositories` 语句（两处）。
3. **移除 Go 代理**: 删除 `ENV GOPROXY=https://goproxy.cn,direct`。

### Task 2: 优化 Makefile 以支持 CI 推送

**Files:**
- Modify: `Makefile`

**Changes:**
1. 修改 `docker-build` 以免默认打上无意义的 `:latest`，并在 CI 环境中通过环境变量接收指定的 `DOCKER_TAG`。
2. 新增 `docker-push` 目标：
   ```makefile
   docker-push:
   	docker push $(DOCKER_IMAGE):$(DOCKER_TAG)
   ```

### Task 3: 编写 GitHub Action 流水线

**Files:**
- Create: `.github/workflows/docker-publish.yml`

**Changes:**
- **触发器 (Triggers)**: `push` 到 `main` 分支，或匹配 `v*` 的 `tags` (如 `v1.0.0`)。
- **流程定义 (Job)**:
  1. `actions/checkout@v4` 拉取代码。
  2. `docker/login-action@v3` 使用 GitHub Secrets 登录 Docker Hub (username: `zcq98`, password: `secrets.DOCKER_SECRET`)。
  3. 通过 Bash 计算当前需构建的 `DOCKER_TAG`，如果是 `main` 分支则为 `latest`，否则为 Tag 名称 (`${GITHUB_REF#refs/tags/}`)。
  4. 运行 `make docker-build DOCKER_TAG=$TAG_NAME`。
  5. 运行 `make docker-push DOCKER_TAG=$TAG_NAME`。

### Task 4: 验证测试
- 确保本地运行 `make docker-build` 可以成功（无国内镜像源）。
- 推送代码到 GitHub 后，观察 Actions 控制台日志。