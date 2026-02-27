#!/bin/bash
#
# OpenClaw Linux Build Script
# Builds Linux binary for amd64 and arm64
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
OpenClaw Linux Build Script

Usage: $0 [COMMAND] [OPTIONS]

Commands:
    all             Build for all Linux platforms (amd64 + arm64)
    amd64           Build for Linux x64 only
    arm64           Build for Linux ARM64 only
    deb             Build Debian package (.deb)
    rpm             Build RPM package (.rpm)
    appimage        Build AppImage
    clean           Clean build artifacts
    help            Show this help message

Options:
    VERSION=x.x.x   Set build version (default: 1.0.0)

Features:
    - Builds native Linux binary (statically linked)
    - Supports x86_64 and ARM64 architectures
    - Creates systemd service file
    - Logs to ~/.local/share/OpenClaw/logs/

Examples:
    $0 all                              # Build all Linux versions
    $0 amd64                            # Build x64 version
    $0 arm64                            # Build ARM64 version
    $0 deb                              # Build Debian package
    VERSION=2.0.0 $0 all                # Build with specific version

EOF
}

# Clean build artifacts
clean() {
    log_info "Cleaning Linux build artifacts..."
    rm -rf "${OUTPUT_DIR}"/openclaw-*-linux-*
    rm -rf "${OUTPUT_DIR}"/*.deb
    rm -rf "${OUTPUT_DIR}"/*.rpm
    rm -rf "${OUTPUT_DIR}"/*.AppImage
    log_success "Clean complete"
}

# Build Linux binary
build_linux() {
    local arch=$1
    local output_name="openclaw-${VERSION}-linux-${arch}"
    local binary_name="openclaw-installer"

    log_info "Building Linux binary for ${arch}..."

    mkdir -p "${OUTPUT_DIR}/${output_name}"

    cd "${INSTALLER_DIR}"
    GOOS=linux GOARCH="${arch}" CGO_ENABLED=0 \
        go build -ldflags "${LDFLAGS}" -gcflags "${GCFLAGS}" \
        -o "${OUTPUT_DIR}/${output_name}/${binary_name}" \
        .
    cd "${PROJECT_DIR}"

    # Create README
    cat > "${OUTPUT_DIR}/${output_name}/README.txt" << EOF
OpenClaw Installer for Linux
============================
Version: ${VERSION}
Architecture: ${arch}
Build Time: ${BUILD_TIME}
Git Commit: ${GIT_COMMIT}

Features:
- Native Linux application (statically linked)
- No dependencies required
- Auto-opens browser on startup
- Logs saved to: ~/.local/share/OpenClaw/logs/

Installation:
1. Run: ./openclaw-installer
2. Wait for browser to open (http://localhost:18080)
3. Follow the web-based installation wizard

Systemd Service (Optional):
To run as a system service:
1. sudo cp openclaw-installer /usr/local/bin/
2. sudo cp openclaw.service /etc/systemd/system/
3. sudo systemctl enable --now openclaw

Supported Platforms:
- Linux x86_64 (amd64)
- Linux ARM64 (aarch64)
- Most modern Linux distributions

Troubleshooting:
- If browser doesn't open, manually visit http://localhost:18080
- Check logs in ~/.local/share/OpenClaw/logs/ for details
- Ensure port 18080 is not in use: sudo lsof -i :18080

For more information, visit: https://github.com/openclaw/openclaw
EOF

    # Create systemd service file
    cat > "${OUTPUT_DIR}/${output_name}/openclaw.service" << EOF
[Unit]
Description=OpenClaw AI Assistant
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/openclaw-installer
Restart=always
RestartSec=5
User=%I

[Install]
WantedBy=multi-user.target
EOF

    # Create install script
    cat > "${OUTPUT_DIR}/${output_name}/install.sh" << 'EOF'
#!/bin/bash
# OpenClaw Installer for Linux

set -e

echo "OpenClaw Installer"
echo "=================="
echo ""

# Check if running as root
if [ "$EUID" -eq 0 ]; then
    INSTALL_DIR="/usr/local/bin"
    SYSTEM_INSTALL=true
else
    INSTALL_DIR="$HOME/.local/bin"
    SYSTEM_INSTALL=false
    mkdir -p "$INSTALL_DIR"
fi

echo "Installing to: $INSTALL_DIR"

# Copy binary
cp "$(dirname "$0")/openclaw-installer" "$INSTALL_DIR/"
chmod +x "$INSTALL_DIR/openclaw-installer"

# Create config directory
mkdir -p "$HOME/.config/openclaw"

if [ "$SYSTEM_INSTALL" = true ]; then
    # Install systemd service
    cp "$(dirname "$0")/openclaw.service" /etc/systemd/system/
    systemctl daemon-reload
    echo "Systemd service installed. Run: sudo systemctl enable --now openclaw"
fi

echo ""
echo "Installation complete!"
echo "Run: openclaw-installer"
EOF

    chmod +x "${OUTPUT_DIR}/${output_name}/install.sh"

    # Make binary executable
    chmod +x "${OUTPUT_DIR}/${output_name}/${binary_name}"

    log_success "Built: ${output_name}"
}

# Build Debian package
build_deb() {
    log_info "Building Debian package..."

    local arch=$1
    local deb_arch="${arch}"
    [ "$arch" = "amd64" ] && deb_arch="amd64"
    [ "$arch" = "arm64" ] && deb_arch="arm64"

    local pkg_name="openclaw-installer_${VERSION}_${deb_arch}"
    local pkg_dir="${OUTPUT_DIR}/${pkg_name}"

    # Create package structure
    mkdir -p "${pkg_dir}/DEBIAN"
    mkdir -p "${pkg_dir}/usr/bin"
    mkdir -p "${pkg_dir}/usr/share/doc/openclaw-installer"
    mkdir -p "${pkg_dir}/etc/openclaw"

    # Build binary
    cd "${INSTALLER_DIR}"
    GOOS=linux GOARCH="${arch}" CGO_ENABLED=0 \
        go build -ldflags "${LDFLAGS}" -gcflags "${GCFLAGS}" \
        -o "${pkg_dir}/usr/bin/openclaw-installer" \
        .
    cd "${PROJECT_DIR}"

    chmod +x "${pkg_dir}/usr/bin/openclaw-installer"

    # Create control file
    cat > "${pkg_dir}/DEBIAN/control" << EOF
Package: openclaw-installer
Version: ${VERSION}
Section: utils
Priority: optional
Architecture: ${deb_arch}
Maintainer: OpenClaw Team <team@openclaw.org>
Description: OpenClaw AI Assistant Installer
 Cross-platform AI assistant installer for OpenClaw.
 Supports automated installation and configuration.
EOF

    # Create postinst script
    cat > "${pkg_dir}/DEBIAN/postinst" << 'EOF'
#!/bin/bash
set -e

# Create config directory
mkdir -p /etc/openclaw

# Create log directory
mkdir -p /var/log/openclaw
chmod 755 /var/log/openclaw

echo "OpenClaw Installer installed successfully!"
echo "Run: openclaw-installer"
EOF
    chmod 755 "${pkg_dir}/DEBIAN/postinst"

    # Build package
    dpkg-deb --build "${pkg_dir}" "${OUTPUT_DIR}/${pkg_name}.deb"

    # Cleanup
    rm -rf "${pkg_dir}"

    log_success "Built: ${pkg_name}.deb"
}

# Build RPM package
build_rpm() {
    log_info "Building RPM package..."
    log_warn "RPM build requires rpmbuild tool"

    local arch=$1
    local rpm_arch="${arch}"
    [ "$arch" = "amd64" ] && rpm_arch="x86_64"
    [ "$arch" = "arm64" ] && rpm_arch="aarch64"

    # Check for rpmbuild
    if ! command -v rpmbuild &> /dev/null; then
        log_error "rpmbuild not found. Please install rpm-build package."
        return 1
    fi

    local rpm_build_dir="${OUTPUT_DIR}/rpmbuild"
    mkdir -p "${rpm_build_dir}"/{BUILD,RPMS,SOURCES,SPECS,SRPMS}

    # Create spec file
    cat > "${rpm_build_dir}/SPECS/openclaw-installer.spec" << EOF
Name:           openclaw-installer
Version:        ${VERSION}
Release:        1%{?dist}
Summary:        OpenClaw AI Assistant Installer

License:        MIT
URL:            https://github.com/openclaw/openclaw
Source0:        openclaw-installer-%{version}.tar.gz

BuildArch:      ${rpm_arch}

%description
Cross-platform AI assistant installer for OpenClaw.
Supports automated installation and configuration.

%prep
%setup -q

%build
# Binary already built

%install
mkdir -p %{buildroot}/usr/bin
install -m 755 openclaw-installer %{buildroot}/usr/bin/
mkdir -p %{buildroot}/etc/openclaw

%files
/usr/bin/openclaw-installer
%dir /etc/openclaw

%changelog
* $(date +"%a %b %d %Y") OpenClaw Team <team@openclaw.org> - ${VERSION}-1
- Initial release
EOF

    # Build binary
    mkdir -p "${rpm_build_dir}/SOURCES/openclaw-installer-${VERSION}"
    cd "${INSTALLER_DIR}"
    GOOS=linux GOARCH="${arch}" CGO_ENABLED=0 \
        go build -ldflags "${LDFLAGS}" -gcflags "${GCFLAGS}" \
        -o "${rpm_build_dir}/SOURCES/openclaw-installer-${VERSION}/openclaw-installer" \
        .
    cd "${PROJECT_DIR}"

    # Create tarball
    cd "${rpm_build_dir}/SOURCES"
    tar -czf "openclaw-installer-${VERSION}.tar.gz" "openclaw-installer-${VERSION}"
    cd "${PROJECT_DIR}"

    # Build RPM
    rpmbuild --define "_topdir ${rpm_build_dir}" -ba "${rpm_build_dir}/SPECS/openclaw-installer.spec"

    # Copy RPM to output
    cp "${rpm_build_dir}/RPMS/${rpm_arch}/openclaw-installer-${VERSION}-1.*.${rpm_arch}.rpm" \
        "${OUTPUT_DIR}/"

    # Cleanup
    rm -rf "${rpm_build_dir}"

    log_success "Built: openclaw-installer-${VERSION}-1.${rpm_arch}.rpm"
}

# Build all Linux versions
build_all() {
    log_info "Building all Linux versions (version: ${VERSION})..."

    build_linux "amd64"
    echo ""
    build_linux "arm64"

    log_success "All Linux builds complete!"
    echo ""
    ls -lh "${OUTPUT_DIR}"/*/openclaw-installer
}

# Main command handler
main() {
    case "${1:-}" in
        all)
            build_all
            ;;
        amd64|x64)
            build_linux "amd64"
            ;;
        arm64|aarch64)
            build_linux "arm64"
            ;;
        deb|debian)
            build_linux "amd64"
            build_deb "amd64"
            ;;
        rpm)
            build_linux "amd64"
            build_rpm "amd64"
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
