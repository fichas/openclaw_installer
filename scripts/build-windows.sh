#!/bin/bash
#
# OpenClaw Windows Build Script
# Builds Windows GUI version (no console window) for amd64 and arm64
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
ADAPTERS_DIR="adapters"

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
OpenClaw Windows Build Script

Usage: $0 [COMMAND] [OPTIONS]

Commands:
    all             Build for all Windows platforms (amd64 + arm64)
    amd64           Build for Windows x64 only
    arm64           Build for Windows ARM64 only
    clean           Clean build artifacts
    help            Show this help message

Options:
    VERSION=x.x.x   Set build version (default: 1.0.0)

Features:
    - Builds Windows GUI application (no console window)
    - Logs redirected to %LOCALAPPDATA%\OpenClaw\Logs\
    - Supports Windows 10/11 x64 and ARM64

Examples:
    $0 all                              # Build all Windows versions
    $0 amd64                            # Build Windows x64 only
    VERSION=2.0.0 $0 all                # Build with specific version

EOF
}

# Clean build artifacts
clean() {
    log_info "Cleaning Windows build artifacts..."
    rm -rf "${OUTPUT_DIR}"/openclaw-*-windows-*
    log_success "Clean complete"
}

# Setup adapters
setup_adapters() {
    log_info "Setting up adapters..."
    mkdir -p "${ADAPTERS_DIR}"

    # WeChat Work adapter
    if [ ! -d "${ADAPTERS_DIR}/wechat-work" ]; then
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
    fi

    # DingTalk adapter
    if [ ! -d "${ADAPTERS_DIR}/dingtalk" ]; then
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
    fi

    # Feishu adapter
    if [ ! -d "${ADAPTERS_DIR}/feishu" ]; then
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
    fi

    log_success "Adapters ready"
}

# Build Windows GUI version
build_windows() {
    local arch=$1
    local output_name="openclaw-${VERSION}-windows-${arch}"
    local binary_name="openclaw-installer.exe"

    log_info "Building Windows GUI version for ${arch}..."

    # Create output directory
    mkdir -p "${OUTPUT_DIR}/${output_name}"

    # Build with Windows GUI subsystem (no console window)
    # -H=windowsgui is the key flag to hide console window
    cd "${INSTALLER_DIR}"
    GOOS=windows GOARCH="${arch}" go build \
        -ldflags "-s -w -H=windowsgui -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
        -gcflags "-trimpath" \
        -o "../${OUTPUT_DIR}/${output_name}/${binary_name}" \
        .
    cd ..

    # Copy adapters
    if [ -d "${ADAPTERS_DIR}" ]; then
        cp -r "${ADAPTERS_DIR}"/* "${OUTPUT_DIR}/${output_name}/" 2>/dev/null || true
    fi

    # Create Windows-specific README
    cat > "${OUTPUT_DIR}/${output_name}/README.txt" << EOF
OpenClaw Installer for Windows
==============================
Version: ${VERSION}
Architecture: ${arch}
Build Time: ${BUILD_TIME}
Git Commit: ${GIT_COMMIT}

Features:
- GUI application (no console window)
- Auto-opens browser on startup
- Logs saved to: %LOCALAPPDATA%\OpenClaw\Logs\

Installation:
1. Double-click openclaw-installer.exe
2. Wait for browser to open (http://localhost:18080)
3. Follow the web-based installation wizard

Supported Platforms:
- Windows 10 (x64/ARM64)
- Windows 11 (x64/ARM64)

Troubleshooting:
- If browser doesn't open, manually visit http://localhost:18080
- Check logs in %LOCALAPPDATA%\OpenClaw\Logs\ for details
- Ensure port 18080 is not in use by another application

For more information, visit: https://github.com/openclaw/openclaw
EOF

    # Create a batch file for easy launching
    cat > "${OUTPUT_DIR}/${output_name}/Start-Installer.bat" << 'EOF'
@echo off
:: OpenClaw Installer Launcher
:: This batch file launches the installer and handles common issues

echo Starting OpenClaw Installer...
echo.

:: Check if running as administrator (optional)
net session > nul 2>&1
if %errorlevel% == 0 (
    echo Running with administrator privileges
) else (
    echo Running without administrator privileges
    echo Some features may require elevation
)
echo.

:: Start the installer
start "" "%~dp0openclaw-installer.exe"

:: Wait a moment
timeout /t 2 /nobreak > nul

:: Try to open browser
echo Opening browser...
start http://localhost:18080

echo.
echo Installer is running. Press any key to close this window.
pause > nul
EOF

    log_success "Built: ${output_name}"
    log_info "Output: ${OUTPUT_DIR}/${output_name}/${binary_name}"
}

# Build all Windows versions
build_all() {
    log_info "Building all Windows versions (version: ${VERSION})..."

    setup_adapters

    build_windows "amd64"
    echo ""
    build_windows "arm64"

    log_success "All Windows builds complete!"
    echo ""
    ls -lh "${OUTPUT_DIR}"/*/openclaw-installer.exe
}

# Main command handler
main() {
    case "${1:-}" in
        all)
            build_all
            ;;
        amd64|x64)
            setup_adapters
            build_windows "amd64"
            ;;
        arm64)
            setup_adapters
            build_windows "arm64"
            ;;
        clean)
            clean
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
