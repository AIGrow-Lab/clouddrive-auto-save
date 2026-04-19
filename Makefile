# ==========================================
# 统一云盘自动转存系统 (UCAS) Makefile
# ==========================================

BIN_DIR = bin
APP_NAME = $(BIN_DIR)/ucas
WEB_DIR = web
GO_BUILD_FLAGS = -v

.PHONY: all help dev-web dev-server build-web build-server build test clean

# 默认执行 help
all: help

# ------------------------------------------
# 开发环境 (Development)
# ------------------------------------------

## dev-web: 启动 Vue 3 前端开发服务器 (运行在 5173 端口, 开启 DEBUG 模式)
dev-web:
	@echo "=> Starting Vue 3 dev server (DEBUG mode)..."
	cd $(WEB_DIR) && LOG_LEVEL=DEBUG npm run dev

## dev-server: 启动 Go 后端开发服务器 (运行在 8080 端口, 开启 DEBUG 日志)
dev-server:
	@echo "=> Starting Go backend server (DEBUG mode)..."
	go mod tidy
	LOG_LEVEL=DEBUG go run cmd/server/main.go

# ------------------------------------------
# 构建打包 (Build)
# ------------------------------------------

## build-web: 编译 Vue 3 前端代码到 web/dist 目录
build-web:
	@echo "=> Building frontend..."
	cd $(WEB_DIR) && npm install && npm run build

## build-server: 编译 Go 后端，并将前端资源内嵌 (依赖 build-web)
build-server: build-web
	@echo "=> Building backend binary..."
	go mod tidy
	mkdir -p $(BIN_DIR)
	go build $(GO_BUILD_FLAGS) -o $(APP_NAME) ./cmd/server/main.go
	@echo "=> Build successful! Binary generated: $(APP_NAME)"

## build: 完整构建流程的快捷别名 (等同于 build-server)
build: build-server

# ------------------------------------------
# 测试与质量检查 (Test & Quality Check)
# ------------------------------------------

## lint: 检查代码格式 (go fmt)
lint:
	@echo "=> Checking code format..."
	@if [ -n "$$(gofmt -l .)" ]; then \
		echo "The following files are not formatted:"; \
		gofmt -l .; \
		exit 1; \
	fi
	@echo "=> Code format check passed."

## vet: 静态分析检查 (go vet)
vet:
	@echo "=> Running go vet..."
	go vet ./...

## test: 运行 Go 单元测试 (含竞态检测和覆盖率)
test:
	@echo "=> Running tests with race detection..."
	go test -v -race ./...
	@echo ""
	@echo "=> Collecting coverage info for packages with tests..."
	@PACKAGES=$$(go list -f '{{if .TestGoFiles}}{{.ImportPath}}{{end}}' ./...); \
	if [ -n "$$PACKAGES" ]; then \
		go test -coverprofile=coverage.out $$PACKAGES > /dev/null; \
		echo "=> Coverage Summary:"; \
		go tool cover -func=coverage.out | tail -n 1; \
	else \
		echo "=> No test files found, skipping coverage collection."; \
	fi

## test-html: 在浏览器中查看覆盖率报告
test-html: test
	@echo "=> Generating and opening coverage report..."
	go tool cover -html=coverage.out

## check: 运行 lint, vet 和 test (完整验证流程)
check: lint vet test
	@echo "=> All checks passed successfully!"

## clean: 清理构建产物 (二进制文件、覆盖率报告和前端 dist 目录)
clean:
	@echo "=> Cleaning build artifacts..."
	rm -rf $(BIN_DIR)
	rm -rf $(WEB_DIR)/dist
	rm -f coverage.out
	@echo "=> Clean finished."

# ------------------------------------------
# 帮助信息 (Help)
# ------------------------------------------

## help: 显示本帮助信息
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
