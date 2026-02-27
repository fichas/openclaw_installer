#!/bin/bash
#
# OpenClaw macOS Build Script
# Builds macOS application bundle for amd64 and arm64
#

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Version and metadata
VERSION="${VERSION:-1.0.0}"
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Directories
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
INSTALLER_DIR="${PROJECT_DIR}/installer"
OUTPUT_DIR="${PROJECT_DIR}/dist"

# Go build flags
LDFLAGS="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}"
GCFLAGS="-trimpath"

# Print helpers
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Show help
show_help() {
    cat << EOF
OpenClaw macOS Build Script

Usage: $0 [COMMAND] [OPTIONS]

Commands:
    all             Build for all macOS platforms (amd64 + arm64)
    amd64           Build for Intel Mac (x86_64)
    arm64           Build for Apple Silicon (M1/M2/M3)
    universal       Build universal binary (amd64 + arm64)
    clean           Clean build artifacts
    help            Show this help message

Options:
    VERSION=x.x.x   Set build version (default: 1.0.0)

Features:
    - Builds native macOS binary
    - Supports Intel and Apple Silicon Macs
    - Creates .app bundle structure
    - Logs to ~/Library/Logs/OpenClaw/

Examples:
    $0 all                              # Build all macOS versions
    $0 amd64                            # Build Intel version
    $0 arm64                            # Build Apple Silicon version
    $0 universal                        # Build universal binary
    VERSION=2.0.0 $0 all                # Build with specific version

EOF
}

# Clean build artifacts
clean() {
    log_info "Cleaning macOS build artifacts..."
    rm -rf "${OUTPUT_DIR}"/openclaw-*-darwin-*
    rm -rf "${OUTPUT_DIR}"/*.app
    log_success "Clean complete"
}

# Create .app bundle structure
create_app_bundle() {
    local arch=$1
    local app_name="OpenClaw-Installer-${VERSION}-${arch}.app"
    local app_dir="${OUTPUT_DIR}/${app_name}"
    local contents_dir="${app_dir}/Contents"
    local macos_dir="${contents_dir}/MacOS"
    local resources_dir="${contents_dir}/Resources"

    log_info "Creating .app bundle for ${arch}..."

    mkdir -p "${macos_dir}"
    mkdir -p "${resources_dir}"

    # Build the binary
    cd "${INSTALLER_DIR}"
    GOOS=darwin GOARCH="${arch}" CGO_ENABLED=0 \
        go build -ldflags "${LDFLAGS}" -gcflags "${GCFLAGS}" \
        -o "${macos_dir}/openclaw-installer" \
        .
    cd "${PROJECT_DIR}"

    # Create Info.plist
    cat > "${contents_dir}/Info.plist" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleDevelopmentRegion</key>
    <string>en</string>
    <key>CFBundleExecutable</key>
    <string>openclaw-installer</string>
    <key>CFBundleIdentifier</key>
    <string>org.openclaw.installer</string>
    <key>CFBundleInfoDictionaryVersion</key>
    <string>6.0</string>
    <key>CFBundleName</key>
    <string>OpenClaw Installer</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>CFBundleShortVersionString</key>
    <string>${VERSION}</string>
    <key>CFBundleVersion</key>
    <string>${VERSION}</string>
    <key>LSMinimumSystemVersion</key>
    <string>10.14</string>
    <key>NSHighResolutionCapable</key>
    <true/>
</dict>
</plist>
EOF

    # Create PkgInfo
    echo "APPL????" > "${contents_dir}/PkgInfo"

    log_success "Created: ${app_name}"
}

# Build macOS binary (without .app bundle)
build_macos() {
    local arch=$1
    local output_name="openclaw-${VERSION}-darwin-${arch}"
    local binary_name="openclaw-installer"

    log_info "Building macOS binary for ${arch}..."

    mkdir -p "${OUTPUT_DIR}/${output_name}"

    cd "${INSTALLER_DIR}"
    GOOS=darwin GOARCH="${arch}" CGO_ENABLED=0 \
        go build -ldflags "${LDFLAGS}" -gcflags "${GCFLAGS}" \
        -o "${OUTPUT_DIR}/${output_name}/${binary_name}" \
        .
    cd "${PROJECT_DIR}"

    # Create README
    cat > "${OUTPUT_DIR}/${output_name}/README.txt" << EOF
OpenClaw Installer for macOS
============================
Version: ${VERSION}
Architecture: ${arch}
Build Time: ${BUILD_TIME}
Git Commit: ${GIT_COMMIT}

Features:
- Native macOS application
- Auto-opens browser on startup
- Logs saved to: ~/Library/Logs/OpenClaw/

Installation:
1. Run: ./openclaw-installer
2. Wait for browser to open (http://localhost:18080)
3. Follow the web-based installation wizard

Supported Platforms:
- macOS 10.14+ (Intel)
- macOS 11.0+ (Apple Silicon)

Troubleshooting:
- If browser doesn't open, manually visit http://localhost:18080
- Check logs in ~/Library/Logs/OpenClaw/ for details
- Ensure port 18080 is not in use by another application

For more information, visit: https://github.com/openclaw/openclaw
EOF

    # Make binary executable
    chmod +x "${OUTPUT_DIR}/${output_name}/${binary_name}"

    log_success "Built: ${output_name}"
}

# Build universal binary (amd64 + arm64)
build_universal() {
    log_info "Building universal binary (amd64 + arm64)..."

    local output_name="openclaw-${VERSION}-darwin-universal"
    local app_name="OpenClaw-Installer-${VERSION}-universal.app"

    # Build both architectures
    build_macos "amd64"
    build_macos "arm64"

    # Create universal binary using lipo
    mkdir -p "${OUTPUT_DIR}/${output_name}"

    lipo -create \
        "${OUTPUT_DIR}/openclaw-${VERSION}-darwin-amd64/openclaw-installer" \
        "${OUTPUT_DIR}/openclaw-${VERSION}-darwin-arm64/openclaw-installer" \
        -output "${OUTPUT_DIR}/${output_name}/openclaw-installer"

    # Copy README
    cp "${OUTPUT_DIR}/openclaw-${VERSION}-darwin-amd64/README.txt" \
        "${OUTPUT_DIR}/${output_name}/README.txt"

    # Make binary executable
    chmod +x "${OUTPUT_DIR}/${output_name}/openclaw-installer"

    log_success "Built universal binary: ${output_name}"
}

# Build all macOS versions
build_all() {
    log_info "Building all macOS versions (version: ${VERSION})..."

    build_macos "amd64"
    echo ""
    build_macos "arm64"

    log_success "All macOS builds complete!"
    echo ""
    ls -lh "${OUTPUT_DIR}"/*/openclaw-installer
}

# Main command handler
main() {
    case "${1:-}" in
        all)
            build_all
            ;;
        amd64|x64|intel)
            build_macos "amd64"
            ;;
        arm64|m1|m2|m3|apple-silicon)
            build_macos "arm64"
            ;;
        universal|fat)
            build_universal
            ;;
        app-bundle)
            create_app_bundle "amd64"
            create_app_bundle "arm64"
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
