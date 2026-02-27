# OpenClaw Installer - Cross-Platform Build Makefile
# 支持 Windows, macOS, Linux 构建

# 变量定义
VERSION ?= 1.0.0
APP_NAME = OpenClaw-Installer
BUILD_DIR = build
SCRIPTS_DIR = scripts

# 默认目标
.PHONY: all help clean version

all: help

# 帮助信息
help:
	@echo "OpenClaw Installer - Cross-Platform Build System"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  all              - 显示此帮助信息"
	@echo "  build            - 构建当前平台"
	@echo "  build-all        - 构建所有平台 (Windows, macOS, Linux)"
	@echo "  windows          - 构建 Windows 版本 (.exe)"
	@echo "  macos            - 构建 macOS 版本 (.app)"
	@echo "  linux            - 构建 Linux 版本"
	@echo "  clean            - 清理构建目录"
	@echo "  version          - 显示版本信息"
	@echo ""
	@echo "环境变量:"
	@echo "  VERSION          - 设置版本号 (默认: 1.0.0)"
	@echo "  APPLE_DEVELOPER_ID - macOS 签名证书 ID"
	@echo ""
	@echo "示例:"
	@echo "  make build-all VERSION=1.2.0"
	@echo "  make windows"
	@echo "  make clean"

# 版本信息
version:
	@echo "OpenClaw Installer v$(VERSION)"
	@go version

# 清理构建目录
clean:
	@echo "Cleaning build directory..."
	@rm -rf $(BUILD_DIR)
	@echo "Done."

# 构建当前平台
build:
	@echo "Building for current platform..."
	@go mod tidy
	@go build -ldflags "-s -w -X main.version=$(VERSION)" -o $(APP_NAME) .
	@echo "Build complete: $(APP_NAME)"

# 构建所有平台
build-all: windows macos linux
	@echo ""
	@echo "=========================================="
	@echo "All platforms built successfully!"
	@echo "=========================================="
	@echo ""
	@echo "Output directories:"
	@echo "  Windows: $(BUILD_DIR)/windows/"
	@echo "  macOS:   $(BUILD_DIR)/macos/"
	@echo "  Linux:   $(BUILD_DIR)/linux/"
	@echo ""

# Windows 构建
windows:
	@echo "=========================================="
	@echo "Building Windows..."
	@echo "=========================================="
	@chmod +x $(SCRIPTS_DIR)/build-windows.sh
	@VERSION=$(VERSION) $(SCRIPTS_DIR)/build-windows.sh

# macOS 构建
macos:
	@echo "=========================================="
	@echo "Building macOS..."
	@echo "=========================================="
	@chmod +x $(SCRIPTS_DIR)/build-macos.sh
	@VERSION=$(VERSION) $(SCRIPTS_DIR)/build-macos.sh

# Linux 构建
linux:
	@echo "=========================================="
	@echo "Building Linux..."
	@echo "=========================================="
	@chmod +x $(SCRIPTS_DIR)/build-linux.sh
	@VERSION=$(VERSION) $(SCRIPTS_DIR)/build-linux.sh

# 快速构建 (仅二进制，不打包)
quick-windows:
	@echo "Quick build Windows x64..."
	@mkdir -p $(BUILD_DIR)/windows
	@GOOS=windows GOARCH=amd64 CGO_ENABLED=1 \
		CC=x86_64-w64-mingw32-gcc \
		go build -ldflags "-s -w -H=windowsgui -X main.version=$(VERSION)" \
		-o $(BUILD_DIR)/windows/$(APP_NAME)-x64.exe .
	@echo "Created: $(BUILD_DIR)/windows/$(APP_NAME)-x64.exe"

quick-macos:
	@echo "Quick build macOS..."
	@mkdir -p $(BUILD_DIR)/macos
	@GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w -X main.version=$(VERSION)" \
		-o $(BUILD_DIR)/macos/$(APP_NAME)-x64 .
	@echo "Created: $(BUILD_DIR)/macos/$(APP_NAME)-x64"

quick-linux:
	@echo "Quick build Linux x64..."
	@mkdir -p $(BUILD_DIR)/linux
	@go build -ldflags "-s -w -X main.version=$(VERSION)" \
		-o $(BUILD_DIR)/linux/$(APP_NAME)-x64 .
	@echo "Created: $(BUILD_DIR)/linux/$(APP_NAME)-x64"

# 开发模式
dev:
	@echo "Running in development mode..."
	@wails dev

# 运行测试
test:
	@echo "Running tests..."
	@go test -v ./...

# 代码格式化
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Done."

# 代码检查
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, skipping..."; \
	fi

# 依赖更新
deps:
	@echo "Updating dependencies..."
	@go mod tidy
	@go mod download
	@echo "Done."

# 安装到系统 (Linux/macOS)
install: build
	@echo "Installing to /usr/local/bin..."
	@sudo cp $(APP_NAME) /usr/local/bin/openclaw-installer
	@echo "Installed as: openclaw-installer"

# 卸载
uninstall:
	@echo "Uninstalling..."
	@sudo rm -f /usr/local/bin/openclaw-installer
	@echo "Done."
