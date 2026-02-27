#!/bin/bash
#
# OpenClaw Release Package Creator
# Creates complete release packages for all platforms
#

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# Version and metadata
VERSION="${VERSION:-1.0.0}"
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Directories
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
DIST_DIR="${PROJECT_DIR}/dist"
RELEASE_DIR="${PROJECT_DIR}/release"
RELEASE_NAME="OpenClaw-v${VERSION}"
RELEASE_PATH="${RELEASE_DIR}/${RELEASE_NAME}"

# Print helpers
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_section() { echo -e "${CYAN}========================================${NC}"; echo -e "${CYAN}$1${NC}"; echo -e "${CYAN}========================================${NC}"; }

# Show help
show_help() {
    cat << EOF
OpenClaw Release Package Creator

Usage: $0 [OPTIONS]

Options:
    -v, --version x.x.x   Set release version (default: 1.0.0)
    -b, --build           Build all platforms before packaging
    -c, --clean           Clean release directory before creating
    -h, --help            Show this help message

Examples:
    $0                              # Create release with existing builds
    $0 -v 2.0.0                     # Create release with version 2.0.0
    $0 -b                           # Build all platforms then create release
    $0 -b -v 2.0.0 -c               # Clean, build, and create release

EOF
}

# Clean release directory
clean_release() {
    log_info "Cleaning release directory..."
    rm -rf "${RELEASE_DIR}"
    mkdir -p "${RELEASE_DIR}"
    log_success "Release directory cleaned"
}

# Build all platforms
build_all() {
    log_section "Building All Platforms"
    if [ -f "${SCRIPT_DIR}/build-all.sh" ]; then
        "${SCRIPT_DIR}/build-all.sh" -v "${VERSION}" all
    else
        log_error "build-all.sh not found"
        exit 1
    fi
}

# Create directory structure
create_structure() {
    log_info "Creating release directory structure..."

    # Platform directories
    mkdir -p "${RELEASE_PATH}/Windows/amd64"
    mkdir -p "${RELEASE_PATH}/Windows/arm64"
    mkdir -p "${RELEASE_PATH}/macOS/amd64"
    mkdir -p "${RELEASE_PATH}/macOS/arm64"
    mkdir -p "${RELEASE_PATH}/Linux/amd64"
    mkdir -p "${RELEASE_PATH}/Linux/arm64"

    # Shared directories
    mkdir -p "${RELEASE_PATH}/packages/config-templates"
    mkdir -p "${RELEASE_PATH}/adapters/wechat-work"
    mkdir -p "${RELEASE_PATH}/adapters/dingtalk"
    mkdir -p "${RELEASE_PATH}/adapters/feishu"

    log_success "Directory structure created"
}

# Copy Windows builds
copy_windows() {
    log_info "Copying Windows builds..."

    # Windows amd64
    if [ -f "${DIST_DIR}/openclaw-${VERSION}-windows-amd64/openclaw-installer.exe" ]; then
        cp "${DIST_DIR}/openclaw-${VERSION}-windows-amd64/openclaw-installer.exe" \
           "${RELEASE_PATH}/Windows/amd64/"
        cp "${DIST_DIR}/openclaw-${VERSION}-windows-amd64/README.txt" \
           "${RELEASE_PATH}/Windows/amd64/" 2>/dev/null || true
    else
        log_warn "Windows amd64 build not found"
    fi

    # Windows arm64
    if [ -f "${DIST_DIR}/openclaw-${VERSION}-windows-arm64/openclaw-installer.exe" ]; then
        cp "${DIST_DIR}/openclaw-${VERSION}-windows-arm64/openclaw-installer.exe" \
           "${RELEASE_PATH}/Windows/arm64/"
    else
        log_warn "Windows arm64 build not found"
    fi

    # Create install.bat
    cat > "${RELEASE_PATH}/Windows/install.bat" << 'EOF'
@echo off
:: OpenClaw Installer for Windows
:: Auto-detects architecture and runs appropriate installer

echo ========================================
echo   OpenClaw Installer for Windows
echo ========================================
echo.

:: Detect architecture
if "%PROCESSOR_ARCHITECTURE%"=="AMD64" (
    set ARCH=amd64
    echo Detected: Windows x64 (amd64)
) else if "%PROCESSOR_ARCHITECTURE%"=="ARM64" (
    set ARCH=arm64
    echo Detected: Windows ARM64
) else (
    echo Warning: Unknown architecture, defaulting to amd64
    set ARCH=amd64
)

:: Check if installer exists
if not exist "%~dp0%ARCH%\openclaw-installer.exe" (
    echo Error: Installer not found for architecture %ARCH%
    echo.
    echo Available installers:
    dir /b "%~dp0"
    pause
    exit /b 1
)

:: Run installer
echo.
echo Starting OpenClaw Installer...
echo.
cd /d "%~dp0%ARCH%"
start "" "openclaw-installer.exe"

echo Installer launched!
echo If browser doesn't open automatically, visit: http://localhost:18080
echo.
timeout /t 3 /nobreak > nul
EOF

    # Create install.ps1
    cat > "${RELEASE_PATH}/Windows/install.ps1" << 'EOF'
# OpenClaw Installer for Windows (PowerShell)
# Auto-detects architecture and runs appropriate installer

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "   OpenClaw Installer for Windows" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Detect architecture
$arch = switch ($env:PROCESSOR_ARCHITECTURE) {
    "AMD64" { "amd64" }
    "ARM64" { "arm64" }
    default { "amd64" }
}

Write-Host "Detected: Windows $arch" -ForegroundColor Green
Write-Host ""

# Check if installer exists
$installerPath = Join-Path $PSScriptRoot "$arch\openclaw-installer.exe"
if (-not (Test-Path $installerPath)) {
    Write-Error "Installer not found: $installerPath"
    exit 1
}

# Run installer
Write-Host "Starting OpenClaw Installer..." -ForegroundColor Green
Start-Process -FilePath $installerPath -WorkingDirectory (Split-Path $installerPath)

Write-Host ""
Write-Host "Installer launched!" -ForegroundColor Green
Write-Host "If browser doesn't open automatically, visit: http://localhost:18080"
Write-Host ""
Start-Sleep -Seconds 3
EOF

    log_success "Windows files copied"
}

# Copy macOS builds
copy_macos() {
    log_info "Copying macOS builds..."

    # macOS amd64
    if [ -f "${DIST_DIR}/openclaw-${VERSION}-darwin-amd64/openclaw-installer" ]; then
        cp "${DIST_DIR}/openclaw-${VERSION}-darwin-amd64/openclaw-installer" \
           "${RELEASE_PATH}/macOS/amd64/OpenClaw-Installer"
        chmod +x "${RELEASE_PATH}/macOS/amd64/OpenClaw-Installer"
    else
        log_warn "macOS amd64 build not found"
    fi

    # macOS arm64
    if [ -f "${DIST_DIR}/openclaw-${VERSION}-darwin-arm64/openclaw-installer" ]; then
        cp "${DIST_DIR}/openclaw-${VERSION}-darwin-arm64/openclaw-installer" \
           "${RELEASE_PATH}/macOS/arm64/OpenClaw-Installer"
        chmod +x "${RELEASE_PATH}/macOS/arm64/OpenClaw-Installer"
    else
        log_warn "macOS arm64 build not found"
    fi

    # Create install-mac.command
    cat > "${RELEASE_PATH}/macOS/install-mac.command" << 'EOF'
#!/bin/bash
# OpenClaw Installer for macOS
# Auto-detects architecture and runs appropriate installer

cd "$(dirname "$0")"

echo "========================================"
echo "   OpenClaw Installer for macOS"
echo "========================================"
echo ""

# Detect architecture
ARCH=$(uname -m)
case "$ARCH" in
    x86_64)
        ARCH="amd64"
        echo "Detected: Intel Mac (x86_64)"
        ;;
    arm64)
        ARCH="arm64"
        echo "Detected: Apple Silicon (arm64)"
        ;;
    *)
        echo "Warning: Unknown architecture ($ARCH), trying amd64"
        ARCH="amd64"
        ;;
esac

# Check if installer exists
INSTALLER="$ARCH/OpenClaw-Installer"
if [ ! -f "$INSTALLER" ]; then
    echo "Error: Installer not found for architecture $ARCH"
    echo ""
    echo "Available installers:"
    ls -1 */OpenClaw-Installer 2>/dev/null || echo "  (none found)"
    exit 1
fi

echo ""
echo "Starting OpenClaw Installer..."
echo ""

# Run installer
chmod +x "$INSTALLER"
"$INSTALLER" &

echo "Installer launched!"
echo "If browser doesn't open automatically, visit: http://localhost:18080"
echo ""
sleep 3
EOF
    chmod +x "${RELEASE_PATH}/macOS/install-mac.command"

    log_success "macOS files copied"
}

# Copy Linux builds
copy_linux() {
    log_info "Copying Linux builds..."

    # Linux amd64
    if [ -f "${DIST_DIR}/openclaw-${VERSION}-linux-amd64/openclaw-installer" ]; then
        cp "${DIST_DIR}/openclaw-${VERSION}-linux-amd64/openclaw-installer" \
           "${RELEASE_PATH}/Linux/amd64/"
        cp "${DIST_DIR}/openclaw-${VERSION}-linux-amd64/openclaw.service" \
           "${RELEASE_PATH}/Linux/amd64/" 2>/dev/null || true
        chmod +x "${RELEASE_PATH}/Linux/amd64/openclaw-installer"
    else
        log_warn "Linux amd64 build not found"
    fi

    # Linux arm64
    if [ -f "${DIST_DIR}/openclaw-${VERSION}-linux-arm64/openclaw-installer" ]; then
        cp "${DIST_DIR}/openclaw-${VERSION}-linux-arm64/openclaw-installer" \
           "${RELEASE_PATH}/Linux/arm64/"
        chmod +x "${RELEASE_PATH}/Linux/arm64/openclaw-installer"
    else
        log_warn "Linux arm64 build not found"
    fi

    # Create install-linux.sh
    cat > "${RELEASE_PATH}/Linux/install-linux.sh" << 'EOF'
#!/bin/bash
# OpenClaw Installer for Linux
# Auto-detects architecture and runs appropriate installer

cd "$(dirname "$0")"

echo "========================================"
echo "   OpenClaw Installer for Linux"
echo "========================================"
echo ""

# Detect architecture
ARCH=$(uname -m)
case "$ARCH" in
    x86_64)
        ARCH="amd64"
        echo "Detected: Linux x64 (amd64)"
        ;;
    aarch64|arm64)
        ARCH="arm64"
        echo "Detected: Linux ARM64"
        ;;
    *)
        echo "Warning: Unknown architecture ($ARCH), trying amd64"
        ARCH="amd64"
        ;;
esac

# Check if installer exists
INSTALLER="$ARCH/openclaw-installer"
if [ ! -f "$INSTALLER" ]; then
    echo "Error: Installer not found for architecture $ARCH"
    echo ""
    echo "Available installers:"
    ls -1 */openclaw-installer 2>/dev/null || echo "  (none found)"
    exit 1
fi

echo ""
echo "Starting OpenClaw Installer..."
echo ""

# Run installer
chmod +x "$INSTALLER"
"$INSTALLER" &

echo "Installer launched!"
echo "If browser doesn't open automatically, visit: http://localhost:18080"
echo ""
sleep 3
EOF
    chmod +x "${RELEASE_PATH}/Linux/install-linux.sh"

    log_success "Linux files copied"
}

# Create shared files
create_shared_files() {
    log_info "Creating shared files..."

    # Create VERSION.txt
    cat > "${RELEASE_PATH}/VERSION.txt" << EOF
OpenClaw Release v${VERSION}
========================

Version: ${VERSION}
Build Time: ${BUILD_TIME}
Git Commit: ${GIT_COMMIT}

Platforms:
- Windows 10/11 (x64, ARM64)
- macOS 10.14+ (Intel, Apple Silicon)
- Linux (x64, ARM64)

For installation instructions, see README.txt
EOF

    # Create config templates
    cat > "${RELEASE_PATH}/packages/config-templates/openclaw.yaml" << 'EOF'
# OpenClaw Configuration Template
# Copy this file to your installation directory and customize

server:
  port: 8080
  host: 0.0.0.0
  tls: false
  # cert_file: /path/to/cert.pem
  # key_file: /path/to/key.pem

log:
  level: info
  file: /var/log/openclaw/openclaw.log

adapters:
  - name: "企业微信"
    type: wechat-work
    enabled: false
    config:
      corp_id: ""
      corp_secret: ""
      agent_id: ""

  - name: "钉钉"
    type: dingtalk
    enabled: false
    config:
      app_key: ""
      app_secret: ""

  - name: "飞书"
    type: feishu
    enabled: false
    config:
      app_id: ""
      app_secret: ""
EOF

    # Create adapter configs
    echo '{"name":"wechat-work","type":"messaging","version":"1.0.0"}' > \
        "${RELEASE_PATH}/adapters/wechat-work/adapter.json"
    echo '{"name":"dingtalk","type":"messaging","version":"1.0.0"}' > \
        "${RELEASE_PATH}/adapters/dingtalk/adapter.json"
    echo '{"name":"feishu","type":"messaging","version":"1.0.0"}' > \
        "${RELEASE_PATH}/adapters/feishu/adapter.json"

    log_success "Shared files created"
}

# Create main README.txt
create_readme() {
    log_info "Creating README.txt..."

    cat > "${RELEASE_PATH}/README.txt" << EOF
OpenClaw v${VERSION} - Quick Start Guide
========================================

Thank you for downloading OpenClaw!

System Requirements
-------------------
- Windows 10/11 (x64 or ARM64)
- macOS 10.14+ (Intel or Apple Silicon)
- Linux (x64 or ARM64)

Quick Installation
------------------

Windows:
1. Double-click "install.bat" (or run "install.ps1" in PowerShell)
2. Or navigate to Windows/amd64/ and run "openclaw-installer.exe"
3. Wait for browser to open (http://localhost:18080)

macOS:
1. Double-click "macOS/install-mac.command"
2. Or open Terminal and run: ./macOS/install-mac.command
3. Wait for browser to open (http://localhost:18080)

Linux:
1. Open terminal in this directory
2. Run: ./Linux/install-linux.sh
3. Or navigate to Linux/amd64/ and run: ./openclaw-installer
4. Wait for browser to open (http://localhost:18080)

Manual Installation
-------------------
If the auto-installer doesn't work, manually run the appropriate binary:
- Windows: Windows/amd64/openclaw-installer.exe
- macOS Intel: macOS/amd64/OpenClaw-Installer
- macOS Apple Silicon: macOS/arm64/OpenClaw-Installer
- Linux x64: Linux/amd64/openclaw-installer
- Linux ARM64: Linux/arm64/openclaw-installer

Configuration
-------------
Configuration templates are in packages/config-templates/
Copy and customize openclaw.yaml for your needs.

Logs
----
- Windows: %APPDATA%\OpenClaw\logs\
- macOS: ~/Library/Logs/OpenClaw/
- Linux: ~/.local/share/OpenClaw/logs/

Support
-------
For more information, visit: https://github.com/openclaw/openclaw
For issues and bug reports: https://github.com/openclaw/openclaw/issues

Version: ${VERSION}
Build Time: ${BUILD_TIME}
EOF

    log_success "README.txt created"
}

# Generate SHA256 checksums
generate_checksums() {
    log_info "Generating SHA256 checksums..."

    cd "${RELEASE_PATH}"

    # Find all binary files and generate checksums
    find . -type f \( -name "*.exe" -o -name "openclaw-installer" -o -name "OpenClaw-Installer" \) -exec sha256sum {} \; > checksums.txt

    # Add other important files
    sha256sum README.txt VERSION.txt >> checksums.txt 2>/dev/null || true

    cd "${PROJECT_DIR}"

    log_success "Checksums generated: ${RELEASE_PATH}/checksums.txt"
}

# Create platform-specific packages
create_platform_packages() {
    log_section "Creating Platform-Specific Packages"

    cd "${RELEASE_DIR}"

    # Check for zip command
    local HAS_ZIP=false
    if command -v zip > /dev/null 2>&1; then
        HAS_ZIP=true
    else
        log_warn "zip command not found, using tar.gz instead"
    fi

    # Windows package
    if [ -d "${RELEASE_NAME}/Windows" ]; then
        log_info "Creating Windows package..."
        if [ "$HAS_ZIP" = true ]; then
            zip -r "${RELEASE_NAME}-Windows.zip" "${RELEASE_NAME}/Windows" "${RELEASE_NAME}/README.txt" "${RELEASE_NAME}/VERSION.txt"
        else
            tar -czf "${RELEASE_NAME}-Windows.tar.gz" "${RELEASE_NAME}/Windows" "${RELEASE_NAME}/README.txt" "${RELEASE_NAME}/VERSION.txt"
        fi
    fi

    # macOS package
    if [ -d "${RELEASE_NAME}/macOS" ]; then
        log_info "Creating macOS package..."
        if [ "$HAS_ZIP" = true ]; then
            zip -r "${RELEASE_NAME}-macOS.zip" "${RELEASE_NAME}/macOS" "${RELEASE_NAME}/README.txt" "${RELEASE_NAME}/VERSION.txt"
        else
            tar -czf "${RELEASE_NAME}-macOS.tar.gz" "${RELEASE_NAME}/macOS" "${RELEASE_NAME}/README.txt" "${RELEASE_NAME}/VERSION.txt"
        fi
    fi

    # Linux package
    if [ -d "${RELEASE_NAME}/Linux" ]; then
        log_info "Creating Linux package..."
        tar -czf "${RELEASE_NAME}-Linux.tar.gz" "${RELEASE_NAME}/Linux" "${RELEASE_NAME}/README.txt" "${RELEASE_NAME}/VERSION.txt"
    fi

    # Full package
    log_info "Creating full release package..."
    if [ "$HAS_ZIP" = true ]; then
        zip -r "${RELEASE_NAME}.zip" "${RELEASE_NAME}"
    else
        tar -czf "${RELEASE_NAME}.tar.gz" "${RELEASE_NAME}"
    fi

    cd "${PROJECT_DIR}"

    log_success "Platform packages created"
}

# Print release summary
print_summary() {
    log_section "Release Summary"

    echo "Release: ${RELEASE_NAME}"
    echo "Version: ${VERSION}"
    echo "Build Time: ${BUILD_TIME}"
    echo "Git Commit: ${GIT_COMMIT}"
    echo ""
    echo "Release Directory: ${RELEASE_PATH}"
    echo ""

    echo "Directory Structure:"
    find "${RELEASE_PATH}" -type f | sort | head -30

    echo ""
    echo "Package Files:"
    ls -lh "${RELEASE_DIR}"/*.zip "${RELEASE_DIR}"/*.tar.gz 2>/dev/null || true
}

# Main function
main() {
    local DO_BUILD=false
    local DO_CLEAN=false

    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -v|--version)
                VERSION="$2"
                RELEASE_NAME="OpenClaw-v${VERSION}"
                RELEASE_PATH="${RELEASE_DIR}/${RELEASE_NAME}"
                shift 2
                ;;
            -b|--build)
                DO_BUILD=true
                shift
                ;;
            -c|--clean)
                DO_CLEAN=true
                shift
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done

    log_section "OpenClaw Release Package Creator"
    echo "Version: ${VERSION}"
    echo "Build Time: ${BUILD_TIME}"
    echo ""

    # Clean if requested
    if [ "$DO_CLEAN" = true ]; then
        clean_release
    fi

    # Build if requested
    if [ "$DO_BUILD" = true ]; then
        build_all
    fi

    # Create release structure
    create_structure

    # Copy platform builds
    copy_windows
    copy_macos
    copy_linux

    # Create shared files
    create_shared_files
    create_readme

    # Generate checksums
    generate_checksums

    # Create packages
    create_platform_packages

    # Print summary
    print_summary

    log_success "Release creation complete!"
    log_info "Release location: ${RELEASE_PATH}"
}

# Run main function
main "$@"
