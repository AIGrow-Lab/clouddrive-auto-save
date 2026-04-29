# 修复 GoReleaser 构建中 fs_embed 找不到前端文件的计划

## 目标 (Objective)
修复在 GitHub Actions 使用 GoReleaser 自动构建发版产物时出现的编译错误：
`build failed: exit status 1: internal/api/fs_embed.go:11:12: pattern all:dist: no matching files found`

## 问题分析 (Context & Motivation)
Go 语言的 `//go:embed` 指令要求目标文件必须相对于声明此指令的源代码文件存在。
目前项目前端构建在 `web/dist` 目录下，而后端声明嵌入的文件是 `internal/api/fs_embed.go`，里面写的是 `//go:embed all:dist`。
在单独的 Docker 构建中，有一步操作 `COPY --from=web-builder /app/web/dist ./internal/api/dist` 把产物挪过去了，所以能成功。
但在使用 GoReleaser 以及直接调用 `make build-server` 时，由于前端文件并没有复制到 `internal/api/dist` 中，加上 `.goreleaser.yaml` 中配置了编译选项 `-tags=embed`，导致 Go 编译器直接报错找不到 `dist` 目录。

## 解决方案 (Proposed Solution)
修改项目的 `Makefile`，在 `build-web` 任务的末尾增加拷贝操作。
前端编译完成后，自动将 `web/dist` 目录复制一份至 `internal/api/dist`，这样无论是本地开发编译带 embed 标签的版本，还是在流水线中使用 GoReleaser 都能保证 `dist` 目录的存在并被成功嵌入。

## 具体改动 (Changes)
**修改 `Makefile`**
定位到 `build-web` 目标：
```makefile
## build-web: 编译 Vue 3 前端代码到 web/dist 目录
build-web:
	@echo "=> Building frontend..."
	cd $(WEB_DIR) && npm install && npm run build
```
修改为：
```makefile
## build-web: 编译 Vue 3 前端代码到 web/dist 目录
build-web:
	@echo "=> Building frontend..."
	cd $(WEB_DIR) && npm install && npm run build
	@echo "=> Copying frontend to internal/api/dist for embedding..."
	rm -rf internal/api/dist
	cp -r $(WEB_DIR)/dist internal/api/dist
```

## 验证 (Verification)
手动运行 `make build-web` 观察 `internal/api/dist` 是否生成，接着运行 `go build -tags=embed ./cmd/server/main.go` 检查是否能够成功编译。
在正式解决后可再次推送更新后的代码，GitHub Actions 会重试 GoReleaser 构建。