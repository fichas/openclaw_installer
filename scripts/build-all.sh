#!/bin/bash
#
# OpenClaw Cross-Platform Build Script
# One-click build for all platforms: Windows, macOS, Linux
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
OUTPUT_DIR="${PROJECT_DIR}/dist"

# Print helpers
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_section() { echo -e "${CYAN}========================================${NC}"; echo -e "${CYAN}$1${NC}"; echo -e "${CYAN}========================================${NC}"; }

# Show help
show_help() {
    cat << EOF
OpenClaw Cross-Platform Build Script

Usage: $0 [OPTIONS] [PLATFORMS...]

Platforms:
    all         Build for all platforms (default)
    windows     Build for Windows (amd64, arm64)
    macos       Build for macOS (amd64, arm64)
    linux       Build for Linux (amd64, arm64)

Options:
    -v, --version x.x.x   Set build version (default: 1.0.0)
    -c, --clean           Clean before build
    -p, --package         Create distribution packages
    -u, --usb             Create USB deployment structure
    -h, --help            Show this help message

Examples:
    $0                              # Build all platforms
    $0 windows linux                # Build Windows and Linux only
    $0 -v 2.0.0 all                 # Build all with version 2.0.0
    $0 -c -p windows                # Clean, build Windows, create packages
    $0 --usb                        # Build all and create USB structure

EOF
}

# Clean build artifacts
clean_build() {
    log_info "Cleaning build artifacts..."
    rm -rf "${OUTPUT_DIR}"
    mkdir -p "${OUTPUT_DIR}"
    log_success "Clean complete"
}

# Build Windows
build_windows() {
    log_section "Building Windows (amd64, arm64)"
    if [ -f "${SCRIPT_DIR}/build-windows.sh" ]; then
        "${SCRIPT_DIR}/build-windows.sh" all
    else
        log_error "build-windows.sh not found"
        return 1
    fi
}

# Build macOS
build_macos() {
    log_section "Building macOS (amd64, arm64)"
    if [ -f "${SCRIPT_DIR}/build-macos.sh" ]; then
        "${SCRIPT_DIR}/build-macos.sh" all
    else
        log_error "build-macos.sh not found"
        return 1
    fi
}

# Build Linux
build_linux() {
    log_section "Building Linux (amd64, arm64)"
    if [ -f "${SCRIPT_DIR}/build-linux.sh" ]; then
        "${SCRIPT_DIR}/build-linux.sh" all
    else
        log_error "build-linux.sh not found"
        return 1
    fi
}

# Create distribution packages
create_packages() {
    log_section "Creating Distribution Packages"

    mkdir -p "${OUTPUT_DIR}/packages"

    # Package Windows builds
    for dir in "${OUTPUT_DIR}"/openclaw-*-windows-*; do
        if [ -d "$dir" ]; then
            name=$(basename "$dir")
            log_info "Packaging $name..."
            (cd "${OUTPUT_DIR}" && zip -r "packages/${name}.zip" "$name")
        fi
    done

    # Package macOS builds
    for dir in "${OUTPUT_DIR}"/openclaw-*-darwin-*; do
        if [ -d "$dir" ]; then
            name=$(basename "$dir")
            log_info "Packaging $name..."
            (cd "${OUTPUT_DIR}" && tar -czf "packages/${name}.tar.gz" "$name")
        fi
    done

    # Package Linux builds
    for dir in "${OUTPUT_DIR}"/openclaw-*-linux-*; do
        if [ -d "$dir" ]; then
            name=$(basename "$dir")
            log_info "Packaging $name..."
            (cd "${OUTPUT_DIR}" && tar -czf "packages/${name}.tar.gz" "$name")
        fi
    done

    log_success "Packages created in ${OUTPUT_DIR}/packages/"
}

# Create USB deployment structure
create_usb_deploy() {
    log_section "Creating USB Deployment Structure"

    local USB_DIR="${OUTPUT_DIR}/usb-deploy/OpenClaw"
    mkdir -p "${USB_DIR}/installers"
    mkdir -p "${USB_DIR}/packages"
    mkdir -p "${USB_DIR}/config"

    log_info "Copying installers..."

    # Copy all builds
    for dir in "${OUTPUT_DIR}"/openclaw-*; do
        if [ -d "$dir" ]; then
            cp -r "$dir" "${USB_DIR}/installers/"
        fi
    done

    # Copy packages
    if [ -d "${OUTPUT_DIR}/packages" ]; then
        cp "${OUTPUT_DIR}"/packages/* "${USB_DIR}/packages/" 2>/dev/null || true
    fi

    # Create README
    cat > "${USB_DIR}/README.txt" << EOF
OpenClaw USB Deployment Package
================================
Version: ${VERSION}
Build Time: ${BUILD_TIME}
Git Commit: ${GIT_COMMIT}

Quick Start:
1. Run the appropriate installer for your platform:
   - Windows: Double-click installers/openclaw-${VERSION}-windows-amd64/openclaw-installer.exe
   - macOS: Run installers/openclaw-${VERSION}-darwin-amd64/openclaw-installer
   - Linux: Run installers/openclaw-${VERSION}-linux-amd64/openclaw-installer

2. Follow the installation wizard in your browser

Platform Support:
- Windows 10/11 (x64, ARM64)
- macOS 10.14+ (Intel, Apple Silicon)
- Linux (x64, ARM64)

For more information: https://github.com/openclaw/openclaw
EOF

    # Create install scripts
    cat > "${USB_DIR}/install.sh" << 'EOF'
#!/bin/bash
# OpenClaw Auto-Install Script for Linux/macOS

PLATFORM=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
    x86_64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
INSTALLER_DIR="${SCRIPT_DIR}/installers/openclaw-EOF
    echo -n "${VERSION}-" >> "${USB_DIR}/install.sh"
    cat >> "${USB_DIR}/install.sh" << 'EOF'
${PLATFORM}-${ARCH}"

if [ ! -d "$INSTALLER_DIR" ]; then
    echo "Error: No installer found for ${PLATFORM}/${ARCH}"
    echo "Available installers:"
    ls -1 "${SCRIPT_DIR}/installers/"
    exit 1
fi

echo "Starting OpenClaw Installer for ${PLATFORM}/${ARCH}..."
cd "$INSTALLER_DIR"
./openclaw-installer
EOF
    chmod +x "${USB_DIR}/install.sh"

    log_success "USB deployment structure created: ${USB_DIR}"
}

# Print build summary
print_summary() {
    log_section "Build Summary"

    echo "Version: ${VERSION}"
    echo "Build Time: ${BUILD_TIME}"
    echo "Git Commit: ${GIT_COMMIT}"
    echo ""
    echo "Output Directory: ${OUTPUT_DIR}"
    echo ""

    if [ -d "${OUTPUT_DIR}" ]; then
        echo "Built Artifacts:"
        find "${OUTPUT_DIR}" -type f -name "openclaw-installer*" -o -name "*.exe" 2>/dev/null | while read -r f; do
            size=$(ls -lh "$f" | awk '{print $5}')
            echo "  $(basename "$f") ($size)"
        done
        echo ""

        echo "Directory Structure:"
        ls -1 "${OUTPUT_DIR}"
    fi
}

# Main function
main() {
    local PLATFORMS=""
    local DO_CLEAN=false
    local DO_PACKAGE=false
    local DO_USB=false

    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -v|--version)
                VERSION="$2"
                shift 2
                ;;
            -c|--clean)
                DO_CLEAN=true
                shift
                ;;
            -p|--package)
                DO_PACKAGE=true
                shift
                ;;
            -u|--usb)
                DO_USB=true
                shift
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            all|windows|macos|linux)
                PLATFORMS="${PLATFORMS} $1"
                shift
                ;;
            *)
                log_error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done

    # Default to all platforms if none specified
    if [ -z "$PLATFORMS" ]; then
        PLATFORMS="all"
    fi

    log_section "OpenClaw Cross-Platform Build"
    echo "Version: ${VERSION}"
    echo "Build Time: ${BUILD_TIME}"
    echo "Platforms: ${PLATFORMS}"
    echo ""

    # Clean if requested
    if [ "$DO_CLEAN" = true ]; then
        clean_build
    fi

    # Create output directory
    mkdir -p "${OUTPUT_DIR}"

    # Export version for sub-scripts
    export VERSION
    export BUILD_TIME
    export GIT_COMMIT

    # Build requested platforms
    for platform in $PLATFORMS; do
        case $platform in
            all)
                build_windows
                build_macos
                build_linux
                ;;
            windows)
                build_windows
                ;;
            macos)
                build_macos
                ;;
            linux)
                build_linux
                ;;
        esac
    done

    # Create packages if requested
    if [ "$DO_PACKAGE" = true ]; then
        create_packages
    fi

    # Create USB deployment if requested
    if [ "$DO_USB" = true ]; then
        create_usb_deploy
    fi

    # Print summary
    print_summary

    log_success "Build complete!"
}

# Run main function
main "$@"
