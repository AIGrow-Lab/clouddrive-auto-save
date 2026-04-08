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

## dev-web: 启动 Vue 3 前端开发服务器 (运行在 5173 端口)
dev-web:
	@echo "=> Starting Vue 3 dev server..."
	cd $(WEB_DIR) && npm run dev

## dev-server: 启动 Go 后端开发服务器 (运行在 8080 端口)
dev-server:
	@echo "=> Starting Go backend server..."
	go mod tidy
	go run cmd/server/main.go

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
# 测试与清理 (Test & Clean)
# ------------------------------------------

## test: 运行 Go 单元测试
test:
	@echo "=> Running tests..."
	go test -v ./...

## clean: 清理构建产物 (二进制文件和前端 dist 目录)
clean:
	@echo "=> Cleaning build artifacts..."
	rm -rf $(BIN_DIR)
	rm -rf $(WEB_DIR)/dist
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
