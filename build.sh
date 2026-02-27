#!/bin/bash
#
# OpenClaw Installer Build Script
# Supports cross-compilation for Windows, macOS, and Linux (amd64/arm64)
#

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Version and metadata
VERSION="${VERSION:-1.0.0}"
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Directories
INSTALLER_DIR="installer"
OUTPUT_DIR="dist"
USB_DIR="${OUTPUT_DIR}/usb-deploy"
ADAPTERS_DIR="adapters"

# Go build flags
LDFLAGS="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}"
GCFLAGS="-trimpath"

# Print helpers
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Show help
show_help() {
    cat << EOF
OpenClaw Installer Build Script

Usage: $0 [COMMAND] [OPTIONS]

Commands:
    all             Build for all platforms
    installer       Build installer for current platform only
    adapters        Download and package adapters
    usb             Create USB deployment package
    clean           Clean build artifacts
    test            Run tests
    release         Create release archives
    help            Show this help message

Platform-specific builds:
    windows-amd64   Build for Windows (amd64)
    windows-arm64   Build for Windows (arm64)
    darwin-amd64    Build for macOS (amd64)
    darwin-arm64    Build for macOS (arm64)
    linux-amd64     Build for Linux (amd64)
    linux-arm64     Build for Linux (arm64)

Options:
    VERSION=x.x.x   Set build version (default: 1.0.0)

Examples:
    $0 all                              # Build everything
    $0 installer                        # Build for current platform
    VERSION=2.0.0 $0 all                # Build with specific version
    $0 linux-amd64                      # Build for Linux amd64 only
    $0 usb                              # Create USB deployment package

EOF
}

# Clean build artifacts
clean() {
    log_info "Cleaning build artifacts..."
    rm -rf "${OUTPUT_DIR}"
    rm -rf "${ADAPTERS_DIR}"
    log_success "Clean complete"
}

# Run tests
run_tests() {
    log_info "Running tests..."
    cd "${INSTALLER_DIR}"
    go test -v ./...
    cd ..
    log_success "Tests complete"
}

# Setup adapters
build_adapters() {
    log_info "Setting up adapters..."
    mkdir -p "${ADAPTERS_DIR}"

    # WeChat Work adapter
    log_info "Creating WeChat Work adapter..."
    mkdir -p "${ADAPTERS_DIR}/wechat-work"
    cat > "${ADAPTERS_DIR}/wechat-work/adapter.json" << 'EOF'
{
    "name": "wechat-work",
    "type": "messaging",
    "display_name": "企业微信",
    "version": "1.0.0",
    "description": "企业微信消息适配器",
    "supported_platforms": ["windows", "darwin", "linux"],
    "config_schema": {
        "corp_id": {"type": "string", "required": true},
        "corp_secret": {"type": "string", "required": true},
        "agent_id": {"type": "string", "required": true}
    }
}
EOF

    # DingTalk adapter
    log_info "Creating DingTalk adapter..."
    mkdir -p "${ADAPTERS_DIR}/dingtalk"
    cat > "${ADAPTERS_DIR}/dingtalk/adapter.json" << 'EOF'
{
    "name": "dingtalk",
    "type": "messaging",
    "display_name": "钉钉",
    "version": "1.0.0",
    "description": "钉钉消息适配器",
    "supported_platforms": ["windows", "darwin", "linux"],
    "config_schema": {
        "app_key": {"type": "string", "required": true},
        "app_secret": {"type": "string", "required": true},
        "robot_code": {"type": "string", "required": false}
    }
}
EOF

    # Feishu adapter
    log_info "Creating Feishu adapter..."
    mkdir -p "${ADAPTERS_DIR}/feishu"
    cat > "${ADAPTERS_DIR}/feishu/adapter.json" << 'EOF'
{
    "name": "feishu",
    "type": "messaging",
    "display_name": "飞书",
    "version": "1.0.0",
    "description": "飞书消息适配器",
    "supported_platforms": ["windows", "darwin", "linux"],
    "config_schema": {
        "app_id": {"type": "string", "required": true},
        "app_secret": {"type": "string", "required": true},
        "encrypt_key": {"type": "string", "required": false}
    }
}
EOF

    log_success "Adapters ready"
}

# Build for specific platform
build_platform() {
    local goos=$1
    local goarch=$2
    local output_name="openclaw-${VERSION}-${goos}-${goarch}"
    local binary_name="openclaw-installer"

    if [ "${goos}" = "windows" ]; then
        binary_name="${binary_name}.exe"
    fi

    log_info "Building for ${goos}/${goarch}..."

    mkdir -p "${OUTPUT_DIR}/${output_name}"

    cd "${INSTALLER_DIR}"
    GOOS="${goos}" GOARCH="${goarch}" go build \
        -ldflags "${LDFLAGS}" \
        -gcflags "${GCFLAGS}" \
        -o "../${OUTPUT_DIR}/${output_name}/${binary_name}" \
        .
    cd ..

    # Copy adapters
    if [ -d "${ADAPTERS_DIR}" ]; then
        cp -r "${ADAPTERS_DIR}"/* "${OUTPUT_DIR}/${output_name}/" 2>/dev/null || true
    fi

    # Create platform-specific README
    cat > "${OUTPUT_DIR}/${output_name}/README.txt" << EOF
OpenClaw Installer ${VERSION}
Platform: ${goos}/${goarch}
Build Time: ${BUILD_TIME}
Git Commit: ${GIT_COMMIT}

Installation:
1. Run the openclaw-installer binary
2. Open http://localhost:18080 in your browser
3. Follow the installation wizard

For more information, visit: https://github.com/openclaw/openclaw
EOF

    log_success "Built: ${output_name}"
}

# Build for all platforms
build_all() {
    log_info "Building for all platforms (version: ${VERSION})..."

    build_adapters

    build_platform "windows" "amd64"
    build_platform "windows" "arm64"
    build_platform "darwin" "amd64"
    build_platform "darwin" "arm64"
    build_platform "linux" "amd64"
    build_platform "linux" "arm64"

    log_success "All builds complete!"
    echo ""
    ls -lh "${OUTPUT_DIR}"/*/
}

# Build for current platform
build_current() {
    log_info "Building for current platform..."

    local goos=$(go env GOOS)
    local goarch=$(go env GOARCH)

    mkdir -p "${OUTPUT_DIR}"

    cd "${INSTALLER_DIR}"
    go build \
        -ldflags "${LDFLAGS}" \
        -gcflags "${GCFLAGS}" \
        -o "../${OUTPUT_DIR}/openclaw-installer" \
        .
    cd ..

    log_success "Build complete: ${OUTPUT_DIR}/openclaw-installer"
}

# Create USB deployment package
create_usb_package() {
    log_info "Creating USB deployment package..."

    # Ensure all platforms are built
    if [ ! -d "${OUTPUT_DIR}/openclaw-${VERSION}-linux-amd64" ]; then
        log_warn "Platform builds not found. Building all platforms first..."
        build_all
    fi

    mkdir -p "${USB_DIR}/installers"
    mkdir -p "${USB_DIR}/adapters"
    mkdir -p "${USB_DIR}/config"

    log_info "Copying installers..."
    cp -r "${OUTPUT_DIR}"/openclaw-${VERSION}-* "${USB_DIR}/installers/"

    log_info "Copying adapters..."
    cp -r "${ADAPTERS_DIR}"/* "${USB_DIR}/adapters/"

    log_info "Creating README..."
    cat > "${USB_DIR}/README.txt" << EOF
OpenClaw USB Deployment Package
================================
Version: ${VERSION}
Build Time: ${BUILD_TIME}
Git Commit: ${GIT_COMMIT}

Quick Start:
1. Run ./install.sh (Linux/macOS) or install.bat (Windows)
2. Or manually navigate to your platform directory in installers/
3. Run the openclaw-installer binary
4. Open http://localhost:18080 in your browser

Platform Directories:
- installers/openclaw-${VERSION}-windows-amd64/
- installers/openclaw-${VERSION}-windows-arm64/
- installers/openclaw-${VERSION}-darwin-amd64/
- installers/openclaw-${VERSION}-darwin-arm64/
- installers/openclaw-${VERSION}-linux-amd64/
- installers/openclaw-${VERSION}-linux-arm64/

Adapters:
- adapters/wechat-work/    (企业微信)
- adapters/dingtalk/       (钉钉)
- adapters/feishu/         (飞书)
EOF

    log_info "Creating install script..."
    cat > "${USB_DIR}/install.sh" << 'EOF'
#!/bin/bash
# Auto-detect platform and run installer

set -e

PLATFORM=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
    x86_64)
        ARCH="amd64"
        ;;
    arm64|aarch64)
        ARCH="arm64"
        ;;
    *)
        echo "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
INSTALLER_DIR="${SCRIPT_DIR}/installers/openclaw-${VERSION}-${PLATFORM}-${ARCH}"

if [ ! -d "$INSTALLER_DIR" ]; then
    echo "Error: No installer found for platform: ${PLATFORM}/${ARCH}"
    echo "Available platforms:"
    ls -1 "${SCRIPT_DIR}/installers/"
    exit 1
fi

echo "Starting OpenClaw Installer for ${PLATFORM}/${ARCH}..."
cd "$INSTALLER_DIR"

if [ -f "./openclaw-installer.exe" ]; then
    ./openclaw-installer.exe
else
    ./openclaw-installer
fi
EOF
    chmod +x "${USB_DIR}/install.sh"

    log_info "Creating Windows install script..."
    cat > "${USB_DIR}/install.bat" << 'EOF'
@echo off
setlocal enabledelayedexpansion

:: Auto-detect architecture
if "%PROCESSOR_ARCHITECTURE%"=="AMD64" (
    set ARCH=amd64
) else if "%PROCESSOR_ARCHITECTURE%"=="ARM64" (
    set ARCH=arm64
) else (
    echo Unsupported architecture: %PROCESSOR_ARCHITECTURE%
    exit /b 1
)

set PLATFORM=windows
set INSTALLER_DIR=installers\openclaw-VERSION-placeholder-%PLATFORM%-!ARCH!

if not exist "!INSTALLER_DIR!" (
    echo Error: No installer found for platform: %PLATFORM%/!ARCH!
    echo Available platforms:
    dir /b installers\
    exit /b 1
)

echo Starting OpenClaw Installer for %PLATFORM%/!ARCH!...
cd "!INSTALLER_DIR!"
start openclaw-installer.exe
EOF
    # Replace VERSION placeholder
    sed -i "s/VERSION-placeholder/${VERSION}/g" "${USB_DIR}/install.bat"

    log_success "USB deployment package created: ${USB_DIR}/"
    echo ""
    tree "${USB_DIR}" 2>/dev/null || find "${USB_DIR}" -type f | head -20
}

# Create release archives
create_releases() {
    log_info "Creating release archives..."

    if [ ! -d "${OUTPUT_DIR}/openclaw-${VERSION}-linux-amd64" ]; then
        log_warn "Platform builds not found. Building all platforms first..."
        build_all
    fi

    mkdir -p "${OUTPUT_DIR}/releases"

    for dir in "${OUTPUT_DIR}"/openclaw-${VERSION}-*/; do
        if [ -d "$dir" ]; then
            name=$(basename "$dir")
            log_info "Creating ${name}.tar.gz..."
            tar -czf "${OUTPUT_DIR}/releases/${name}.tar.gz" -C "${OUTPUT_DIR}" "$name"
        fi
    done

    # Create checksums
    cd "${OUTPUT_DIR}/releases"
    sha256sum *.tar.gz > checksums.txt 2>/dev/null || shasum -a 256 *.tar.gz > checksums.txt
    cd ../..

    log_success "Release archives created in ${OUTPUT_DIR}/releases/"
    ls -lh "${OUTPUT_DIR}/releases/"
}

# Main command handler
main() {
    case "${1:-}" in
        all)
            build_all
            ;;
        installer|current)
            build_current
            ;;
        adapters)
            build_adapters
            ;;
        usb|usb-deploy)
            create_usb_package
            ;;
        clean)
            clean
            ;;
        test)
            run_tests
            ;;
        release)
            create_releases
            ;;
        windows-amd64)
            build_adapters
            build_platform "windows" "amd64"
            ;;
        windows-arm64)
            build_adapters
            build_platform "windows" "arm64"
            ;;
        darwin-amd64|macos-amd64)
            build_adapters
            build_platform "darwin" "amd64"
            ;;
        darwin-arm64|macos-arm64)
            build_adapters
            build_platform "darwin" "arm64"
            ;;
        linux-amd64)
            build_adapters
            build_platform "linux" "amd64"
            ;;
        linux-arm64)
            build_adapters
            build_platform "linux" "arm64"
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            log_error "Unknown command: ${1:-}"
            echo ""
            show_help
            exit 1
            ;;
    esac
}

# Run main function
main "$@"
