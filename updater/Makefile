# OpenClaw Updater Makefile

VERSION ?= 1.0.0
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

# Build directories
BUILD_DIR := build
DIST_DIR := dist

# Platform targets
PLATFORMS := linux-amd64 linux-arm64 darwin-amd64 darwin-arm64 windows-amd64

.PHONY: all build build-all clean test install uninstall

all: build

# Build for current platform
build:
	@echo "Building OpenClaw Updater for current platform..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/openclaw-updater .

# Build for all platforms
build-all: $(PLATFORMS)

$(PLATFORMS):
	@echo "Building for $@..."
	@mkdir -p $(DIST_DIR)
	@GOOS=$(word 1,$(subst -, ,$@)) GOARCH=$(word 2,$(subst -, ,$@)) \
		go build $(LDFLAGS) -o $(DIST_DIR)/openclaw-updater-$@$(if $(findstring windows,$@),.exe,) .

linux-amd64:
linux-arm64:
darwin-amd64:
darwin-arm64:
windows-amd64:

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR) $(DIST_DIR) coverage.out coverage.html

# Install locally (requires sudo on Linux/macOS)
install: build
	@echo "Installing OpenClaw Updater..."
	@install -d /usr/local/bin
	@install -m 755 $(BUILD_DIR)/openclaw-updater /usr/local/bin/
	@install -d /etc/openclaw
	@echo '{"version":"1.0.0","auto_update":false}' > /etc/openclaw/updater.json
	@echo "Installation complete. Run 'openclaw-updater -help' for usage."

# Uninstall
uninstall:
	@echo "Uninstalling OpenClaw Updater..."
	@rm -f /usr/local/bin/openclaw-updater
	@echo "Uninstall complete."

# Development mode - run with dry-run
dev: build
	./$(BUILD_DIR)/openclaw-updater -dry-run -v -yes

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Run linter
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, skipping..."; \
	fi

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Generate documentation
docs:
	@echo "Generating documentation..."
	@mkdir -p docs/generated
	@go doc -all . > docs/generated/api.txt

# Package for distribution
package: build-all
	@echo "Creating distribution packages..."
	@mkdir -p $(DIST_DIR)/packages
	@for platform in $(PLATFORMS); do \
		if [ -f "$(DIST_DIR)/openclaw-updater-$$platform" ] || [ -f "$(DIST_DIR)/openclaw-updater-$$platform.exe" ]; then \
			echo "Packaging $$platform..."; \
			if echo "$$platform" | grep -q "windows"; then \
				cd $(DIST_DIR) && zip -q packages/openclaw-updater-$$platform.zip openclaw-updater-$$platform.exe ../scripts/openclaw-update.bat ../README.md; \
			else \
				cd $(DIST_DIR) && tar -czf packages/openclaw-updater-$$platform.tar.gz openclaw-updater-$$platform ../scripts/openclaw-update.sh ../README.md; \
			fi; \
		fi; \
	done
	@echo "Packages created in $(DIST_DIR)/packages/"

# Help
help:
	@echo "OpenClaw Updater Makefile"
	@echo ""
	@echo "Targets:"
	@echo "  build          Build for current platform"
	@echo "  build-all      Build for all platforms"
	@echo "  test           Run tests"
	@echo "  test-coverage  Run tests with coverage report"
	@echo "  clean          Clean build artifacts"
	@echo "  install        Install locally (requires sudo)"
	@echo "  uninstall      Uninstall from system"
	@echo "  dev            Run in development mode (dry-run)"
	@echo "  fmt            Format code"
	@echo "  lint           Run linter"
	@echo "  deps           Download dependencies"
	@echo "  docs           Generate documentation"
	@echo "  package        Create distribution packages"
	@echo "  help           Show this help"
