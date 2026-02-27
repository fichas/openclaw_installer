# OpenClaw Installer Build Makefile
# Supports cross-compilation for multiple platforms

# Project metadata
PROJECT_NAME := openclaw-installer
VERSION := 1.0.0
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Go settings
GO := go
GO_VERSION := 1.19
INSTALLER_DIR := ./installer
UPDATER_DIR := ./updater

# Build flags
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT) -s -w"
GCFLAGS := -gcflags="-trimpath"
ASMFLAGS := -asmflags="-trimpath"

# Output directories
DIST_DIR := ./dist
BUILD_DIR := ./build
RELEASE_DIR := ./release
USB_DIR := ./usb-deploy/OpenClaw

# Target platforms
PLATFORMS := darwin-amd64 darwin-arm64 windows-amd64 windows-arm64 linux-amd64 linux-arm64

# Source files
INSTALLER_SRC := $(wildcard $(INSTALLER_DIR)/*.go)

# Default target
.PHONY: all
all: clean build-all package

# Help target
.PHONY: help
help:
	@echo "OpenClaw Installer Build System"
	@echo ""
	@echo "Available targets:"
	@echo "  make build-all       - Build all platform binaries"
	@echo "  make build-darwin    - Build macOS binaries (amd64, arm64)"
	@echo "  make build-windows   - Build Windows binaries (amd64, arm64)"
	@echo "  make build-windows-gui - Build Windows GUI binaries (no console)"
	@echo "  make build-linux     - Build Linux binaries (amd64, arm64)"
	@echo "  make package         - Create distribution packages"
	@echo "  make usb-deploy      - Create U盘 deployment structure"
	@echo "  make clean           - Clean build artifacts"
	@echo "  make test            - Run all tests"
	@echo "  make test-installer  - Run installer tests only"
	@echo "  make lint            - Run linters"
	@echo "  make version         - Display version information"

# Version information
.PHONY: version
version:
	@echo "Project: $(PROJECT_NAME)"
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "Go Version: $(GO_VERSION)"

# Create directories
$(DIST_DIR):
	@mkdir -p $(DIST_DIR)

$(BUILD_DIR):
	@mkdir -p $(BUILD_DIR)

$(RELEASE_DIR):
	@mkdir -p $(RELEASE_DIR)

$(USB_DIR):
	@mkdir -p $(USB_DIR)/installers
	@mkdir -p $(USB_DIR)/packages/openclaw-core
	@mkdir -p $(USB_DIR)/packages/adapters/wecom-adapter
	@mkdir -p $(USB_DIR)/packages/adapters/dingtalk-adapter
	@mkdir -p $(USB_DIR)/packages/adapters/feishu-adapter
	@mkdir -p $(USB_DIR)/packages/config-templates
	@mkdir -p $(USB_DIR)/resources/icons
	@mkdir -p $(USB_DIR)/resources/licenses
	@mkdir -p $(USB_DIR)/resources/docs
	@mkdir -p $(USB_DIR)/autorun

# Build all platforms
.PHONY: build-all
build-all: build-darwin build-windows build-linux
	@echo "All platforms built successfully!"
	@echo "Binaries location: $(DIST_DIR)/"
	@ls -lh $(DIST_DIR)/

# macOS builds
.PHONY: build-darwin
build-darwin: $(DIST_DIR)
	@echo "Building macOS binaries..."
	@echo "  -> Building darwin-amd64..."
	@cd $(INSTALLER_DIR) && GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 \
		$(GO) build $(LDFLAGS) $(GCFLAGS) $(ASMFLAGS) \
		-o ../$(DIST_DIR)/$(PROJECT_NAME)-darwin-amd64 .
	@echo "  -> Building darwin-arm64..."
	@cd $(INSTALLER_DIR) && GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 \
		$(GO) build $(LDFLAGS) $(GCFLAGS) $(ASMFLAGS) \
		-o ../$(DIST_DIR)/$(PROJECT_NAME)-darwin-arm64 .
	@echo "macOS builds complete!"

# Windows builds
.PHONY: build-windows
build-windows: $(DIST_DIR)
	@echo "Building Windows binaries..."
	@echo "  -> Building windows-amd64..."
	@cd $(INSTALLER_DIR) && GOOS=windows GOARCH=amd64 CGO_ENABLED=0 \
		$(GO) build $(LDFLAGS) $(GCFLAGS) $(ASMFLAGS) \
		-o ../$(DIST_DIR)/$(PROJECT_NAME)-windows-amd64.exe .
	@echo "  -> Building windows-arm64..."
	@cd $(INSTALLER_DIR) && GOOS=windows GOARCH=arm64 CGO_ENABLED=0 \
		$(GO) build $(LDFLAGS) $(GCFLAGS) $(ASMFLAGS) \
		-o ../$(DIST_DIR)/$(PROJECT_NAME)-windows-arm64.exe .
	@echo "Windows builds complete!"

# Windows GUI builds (no console window)
.PHONY: build-windows-gui
build-windows-gui: $(DIST_DIR)
	@echo "Building Windows GUI binaries (no console window)..."
	@echo "  -> Building windows-amd64-gui..."
	@cd $(INSTALLER_DIR) && GOOS=windows GOARCH=amd64 CGO_ENABLED=0 \
		$(GO) build -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT) -s -w -H=windowsgui" $(GCFLAGS) $(ASMFLAGS) \
		-o ../$(DIST_DIR)/$(PROJECT_NAME)-windows-amd64-gui.exe .
	@echo "  -> Building windows-arm64-gui..."
	@cd $(INSTALLER_DIR) && GOOS=windows GOARCH=arm64 CGO_ENABLED=0 \
		$(GO) build -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT) -s -w -H=windowsgui" $(GCFLAGS) $(ASMFLAGS) \
		-o ../$(DIST_DIR)/$(PROJECT_NAME)-windows-arm64-gui.exe .
	@echo "Windows GUI builds complete!"
	@echo "Note: GUI binaries log to %LOCALAPPDATA%\OpenClaw\Logs\"

# Linux builds
.PHONY: build-linux
build-linux: $(DIST_DIR)
	@echo "Building Linux binaries..."
	@echo "  -> Building linux-amd64..."
	@cd $(INSTALLER_DIR) && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
		$(GO) build $(LDFLAGS) $(GCFLAGS) $(ASMFLAGS) \
		-o ../$(DIST_DIR)/$(PROJECT_NAME)-linux-amd64 .
	@echo "  -> Building linux-arm64..."
	@cd $(INSTALLER_DIR) && GOOS=linux GOARCH=arm64 CGO_ENABLED=0 \
		$(GO) build $(LDFLAGS) $(GCFLAGS) $(ASMFLAGS) \
		-o ../$(DIST_DIR)/$(PROJECT_NAME)-linux-arm64 .
	@echo "Linux builds complete!"

# Build specific platform (usage: make build-single PLATFORM=linux-amd64)
.PHONY: build-single
build-single: $(DIST_DIR)
	@if [ -z "$(PLATFORM)" ]; then \
		echo "Error: PLATFORM not specified. Usage: make build-single PLATFORM=linux-amd64"; \
		echo "Available platforms: $(PLATFORMS)"; \
		exit 1; \
	fi
	@echo "Building for $(PLATFORM)..."
	@cd $(INSTALLER_DIR) && \
		GOOS=$(shell echo $(PLATFORM) | cut -d- -f1) \
		GOARCH=$(shell echo $(PLATFORM) | cut -d- -f2) \
		CGO_ENABLED=0 \
		$(GO) build $(LDFLAGS) $(GCFLAGS) $(ASMFLAGS) \
		-o ../$(DIST_DIR)/$(PROJECT_NAME)-$(PLATFORM)$(shell echo $(PLATFORM) | grep -q windows && echo '.exe') .
	@echo "Build complete: $(DIST_DIR)/$(PROJECT_NAME)-$(PLATFORM)"

# Build updater binary
.PHONY: build-updater
build-updater: $(DIST_DIR)
	@echo "Building updater binaries..."
	@for platform in $(PLATFORMS); do \
		os=$$(echo $$platform | cut -d- -f1); \
		arch=$$(echo $$platform | cut -d- -f2); \
		ext=$$(echo $$platform | grep -q windows && echo '.exe' || echo ''); \
		echo "  -> Building updater-$$platform..."; \
		cd $(UPDATER_DIR) && GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 \
			$(GO) build $(LDFLAGS) -o ../$(DIST_DIR)/openclaw-updater-$$platform$$ext . 2>/dev/null || \
			echo "    (skipped - no updater source)"; \
		done
	@echo "Updater builds complete!"

# Run tests
.PHONY: test
test: test-installer

.PHONY: test-installer
test-installer:
	@echo "Running installer tests..."
	@cd $(INSTALLER_DIR) && $(GO) test -v -race -coverprofile=coverage.out ./...
	@cd $(INSTALLER_DIR) && $(GO) tool cover -func=coverage.out

# Run tests for all platforms (cross-compilation check)
.PHONY: test-compile
test-compile:
	@echo "Testing cross-compilation for all platforms..."
	@for platform in $(PLATFORMS); do \
		os=$$(echo $$platform | cut -d- -f1); \
		arch=$$(echo $$platform | cut -d- -f2); \
		echo "  -> Testing compile for $$platform..."; \
		(cd $(INSTALLER_DIR) && GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 \
			$(GO) build -o /dev/null .) || exit 1; \
	done
	@echo "All platforms compile successfully!"

# Lint code
.PHONY: lint
lint:
	@echo "Running linters..."
	@which golangci-lint > /dev/null 2>&1 || (echo "golangci-lint not installed, installing..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	@cd $(INSTALLER_DIR) && golangci-lint run ./...

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	@cd $(INSTALLER_DIR) && $(GO) fmt ./...

# Vet code
.PHONY: vet
vet:
	@echo "Running go vet..."
	@cd $(INSTALLER_DIR) && $(GO) vet ./...

# Create distribution packages
.PHONY: package
package: build-all $(RELEASE_DIR)
	@echo "Creating distribution packages..."
	@for platform in $(PLATFORMS); do \
		os=$$(echo $$platform | cut -d- -f1); \
		arch=$$(echo $$platform | cut -d- -f2); \
		bin_name=$(PROJECT_NAME)-$$platform; \
		if echo $$platform | grep -q windows; then \
			bin_name=$${bin_name}.exe; \
			archive_name=$(PROJECT_NAME)-$$platform.zip; \
			echo "  -> Creating $$archive_name..."; \
			cd $(DIST_DIR) && zip -q $$archive_name $$bin_name; \
		else \
			archive_name=$(PROJECT_NAME)-$$platform.tar.gz; \
			echo "  -> Creating $$archive_name..."; \
			cd $(DIST_DIR) && tar -czf $$archive_name $$bin_name; \
		fi; \
	done
	@echo "Distribution packages created in $(RELEASE_DIR)/"
	@ls -lh $(DIST_DIR)/*.zip $(DIST_DIR)/*.tar.gz 2>/dev/null || true

# Create U盘 deployment structure
.PHONY: usb-deploy
usb-deploy: build-all $(USB_DIR)
	@echo "Creating U盘 deployment structure..."
	@echo "  -> Copying installer binaries..."
	@cp $(DIST_DIR)/$(PROJECT_NAME)-darwin-amd64 $(USB_DIR)/installers/ 2>/dev/null || true
	@cp $(DIST_DIR)/$(PROJECT_NAME)-darwin-arm64 $(USB_DIR)/installers/ 2>/dev/null || true
	@cp $(DIST_DIR)/$(PROJECT_NAME)-windows-amd64.exe $(USB_DIR)/installers/ 2>/dev/null || true
	@cp $(DIST_DIR)/$(PROJECT_NAME)-windows-arm64.exe $(USB_DIR)/installers/ 2>/dev/null || true
	@cp $(DIST_DIR)/$(PROJECT_NAME)-linux-amd64 $(USB_DIR)/installers/ 2>/dev/null || true
	@cp $(DIST_DIR)/$(PROJECT_NAME)-linux-arm64 $(USB_DIR)/installers/ 2>/dev/null || true

	@echo "  -> Creating configuration templates..."
	@echo '# OpenClaw Configuration Template' > $(USB_DIR)/packages/config-templates/openclaw.yaml.template
	@echo 'server:' >> $(USB_DIR)/packages/config-templates/openclaw.yaml.template
	@echo '  port: 8080' >> $(USB_DIR)/packages/config-templates/openclaw.yaml.template
	@echo '  host: 0.0.0.0' >> $(USB_DIR)/packages/config-templates/openclaw.yaml.template
	@echo '' >> $(USB_DIR)/packages/config-templates/openclaw.yaml.template
	@echo 'log:' >> $(USB_DIR)/packages/config-templates/openclaw.yaml.template
	@echo '  level: info' >> $(USB_DIR)/packages/config-templates/openclaw.yaml.template
	@echo '  file: /var/log/openclaw/openclaw.log' >> $(USB_DIR)/packages/config-templates/openclaw.yaml.template
	@echo '' >> $(USB_DIR)/packages/config-templates/openclaw.yaml.template
	@echo 'adapters:' >> $(USB_DIR)/packages/config-templates/openclaw.yaml.template
	@echo '  - wecom' >> $(USB_DIR)/packages/config-templates/openclaw.yaml.template
	@echo '  - dingtalk' >> $(USB_DIR)/packages/config-templates/openclaw.yaml.template
	@echo '  - feishu' >> $(USB_DIR)/packages/config-templates/openclaw.yaml.template

	@echo '# WeCom Adapter Configuration' > $(USB_DIR)/packages/config-templates/wecom-adapter.yaml.template
	@echo 'webhook_url: "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=YOUR_KEY"' >> $(USB_DIR)/packages/config-templates/wecom-adapter.yaml.template
	@echo 'corp_id: "YOUR_CORP_ID"' >> $(USB_DIR)/packages/config-templates/wecom-adapter.yaml.template
	@echo 'corp_secret: "YOUR_CORP_SECRET"' >> $(USB_DIR)/packages/config-templates/wecom-adapter.yaml.template

	@echo '# DingTalk Adapter Configuration' > $(USB_DIR)/packages/config-templates/dingtalk-adapter.yaml.template
	@echo 'webhook_url: "https://oapi.dingtalk.com/robot/send?access_token=YOUR_TOKEN"' >> $(USB_DIR)/packages/config-templates/dingtalk-adapter.yaml.template
	@echo 'secret: "YOUR_SECRET"' >> $(USB_DIR)/packages/config-templates/dingtalk-adapter.yaml.template

	@echo '# Feishu Adapter Configuration' > $(USB_DIR)/packages/config-templates/feishu-adapter.yaml.template
	@echo 'webhook_url: "https://open.feishu.cn/open-apis/bot/v2/hook/YOUR_TOKEN"' >> $(USB_DIR)/packages/config-templates/feishu-adapter.yaml.template
	@echo 'app_id: "YOUR_APP_ID"' >> $(USB_DIR)/packages/config-templates/feishu-adapter.yaml.template
	@echo 'app_secret: "YOUR_APP_SECRET"' >> $(USB_DIR)/packages/config-templates/feishu-adapter.yaml.template

	@echo "  -> Copying installer scripts from template..."
	@cp ./usb-template/install.bat $(USB_DIR)/install.bat
	@cp ./usb-template/install.ps1 $(USB_DIR)/install.ps1
	@cp ./usb-template/install-mac.command $(USB_DIR)/install-mac.command
	@cp ./usb-template/install-linux.sh $(USB_DIR)/install-linux.sh
	@chmod +x $(USB_DIR)/install-mac.command
	@chmod +x $(USB_DIR)/install-linux.sh

	@echo "  -> Copying improved autorun scripts..."
	@cp ./usb-template/install.bat $(USB_DIR)/autorun/install.bat 2>/dev/null || true
	@cp ./usb-template/install-mac.command $(USB_DIR)/autorun/install-mac.command 2>/dev/null || true
	@cp ./usb-template/install-linux.sh $(USB_DIR)/autorun/install-linux.sh 2>/dev/null || true
	@chmod +x $(USB_DIR)/autorun/install-mac.command 2>/dev/null || true
	@chmod +x $(USB_DIR)/autorun/install-linux.sh 2>/dev/null || true

	@echo "  -> Creating Windows autorun.inf..."
	@echo '[autorun]' > $(USB_DIR)/autorun/autorun.inf
	@echo 'open=install.bat' >> $(USB_DIR)/autorun/autorun.inf
	@echo 'label=OpenClaw Installer' >> $(USB_DIR)/autorun/autorun.inf
	@echo 'icon=resources\icons\openclaw.ico' >> $(USB_DIR)/autorun/autorun.inf
	@echo 'action=Install OpenClaw' >> $(USB_DIR)/autorun/autorun.inf

	@echo "  -> Copying README..."
	@cp ./usb-template/README.txt $(USB_DIR)/README.txt

	@echo "U盘 deployment structure created at $(USB_DIR)/"
	@echo "Directory structure:"
	@find $(USB_DIR) -type f | head -30

# Download adapter packages (placeholder - requires actual download URLs)
.PHONY: download-adapters
download-adapters: $(USB_DIR)
	@echo "Downloading adapter packages..."
	@echo "Note: This requires actual adapter package URLs"
	@echo "Creating placeholder adapter package structure..."
	@for adapter in wecom dingtalk feishu; do \
		for platform in $(PLATFORMS); do \
			os=$$(echo $$platform | cut -d- -f1); \
			arch=$$(echo $$platform | cut -d- -f2); \
			if echo $$platform | grep -q windows; then \
				touch $(USB_DIR)/packages/adapters/$$adapter-adapter/$$adapter-adapter-$$platform.zip; \
			else \
				touch $(USB_DIR)/packages/adapters/$$adapter-adapter/$$adapter-adapter-$$platform.tar.gz; \
			fi; \
		done; \
	done
	@echo "Adapter package structure created (placeholder files)"

# Create release archive
.PHONY: release-all
release-all: clean build-all package usb-deploy
	@echo "Creating release archive..."
	@mkdir -p $(RELEASE_DIR)
	@cp -r $(USB_DIR) $(RELEASE_DIR)/
	@cd $(RELEASE_DIR) && tar -czf openclaw-installer-$(VERSION)-usb-deploy.tar.gz OpenClaw/
	@echo "Release archive created: $(RELEASE_DIR)/openclaw-installer-$(VERSION)-usb-deploy.tar.gz"

# Verify builds
.PHONY: verify
verify: build-all
	@echo "Verifying builds..."
	@echo "Checking binary sizes and architectures:"
	@for platform in $(PLATFORMS); do \
		bin=$(DIST_DIR)/$(PROJECT_NAME)-$$platform; \
		if echo $$platform | grep -q windows; then bin=$${bin}.exe; fi; \
		if [ -f "$$bin" ]; then \
			size=$$(ls -lh "$$bin" | awk '{print $$5}'); \
			echo "  $$platform: $$size"; \
		else \
			echo "  $$platform: MISSING!"; \
		fi; \
	done

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(DIST_DIR) $(BUILD_DIR) $(RELEASE_DIR) $(USB_DIR)
	@cd $(INSTALLER_DIR) && rm -f coverage.out
	@echo "Clean complete!"

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	@cd $(INSTALLER_DIR) && $(GO) mod download
	@cd $(INSTALLER_DIR) && $(GO) mod verify

# Update dependencies
.PHONY: deps-update
deps-update:
	@echo "Updating dependencies..."
	@cd $(INSTALLER_DIR) && $(GO) get -u ./...
	@cd $(INSTALLER_DIR) && $(GO) mod tidy

# Development build (current platform only)
.PHONY: dev
dev:
	@echo "Building for current platform (development)..."
	@cd $(INSTALLER_DIR) && $(GO) build -o ../$(DIST_DIR)/$(PROJECT_NAME)-dev .
	@echo "Development build complete: $(DIST_DIR)/$(PROJECT_NAME)-dev"

# Run development server
.PHONY: run
run: dev
	@echo "Running development server..."
	@cd $(INSTALLER_DIR) && $(GO) run .

# Generate checksums
.PHONY: checksums
checksums: build-all
	@echo "Generating checksums..."
	@cd $(DIST_DIR) && sha256sum $(PROJECT_NAME)-* > checksums.sha256
	@echo "Checksums saved to $(DIST_DIR)/checksums.sha256"

# Docker build (optional)
.PHONY: docker-build
docker-build:
	@echo "Building with Docker..."
	@docker run --rm -v "$(PWD)":/src -w /src/installer \
		-e CGO_ENABLED=0 \
		golang:$(GO_VERSION) \
		go build -o ../$(DIST_DIR)/$(PROJECT_NAME)-linux-amd64 .

# CI/CD targets
.PHONY: ci-build
ci-build: deps test-compile build-all verify checksums

.PHONY: ci-test
ci-test: test

# Platform-specific shortcuts
.PHONY: mac
mac: build-darwin

.PHONY: win
win: build-windows

.PHONY: linux-only
linux-only: build-linux
