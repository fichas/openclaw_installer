#!/bin/bash

# Linux 构建脚本
# 支持 x64 和 ARM64 架构

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BUILD_DIR="$PROJECT_ROOT/build/linux"

# 版本信息
VERSION="${VERSION:-1.0.0}"
APP_NAME="openclaw-installer"

echo "=========================================="
echo "Building Linux installer"
echo "Version: $VERSION"
echo "=========================================="

cd "$PROJECT_ROOT"

# 创建构建目录
mkdir -p "$BUILD_DIR"

# 下载依赖
echo "[1/6] Downloading dependencies..."
go mod tidy

# 构建 Linux x64
echo "[2/6] Building Linux x64..."
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
    go build -ldflags "-s -w -X main.version=$VERSION" \
    -o "$BUILD_DIR/${APP_NAME}-x64" .

echo "    Created: $BUILD_DIR/${APP_NAME}-x64"

# 构建 Linux ARM64
echo "[3/6] Building Linux ARM64..."
CGO_ENABLED=1 GOOS=linux GOARCH=arm64 \
    CC=aarch64-linux-gnu-gcc \
    go build -ldflags "-s -w -X main.version=$VERSION" \
    -o "$BUILD_DIR/${APP_NAME}-arm64" . 2>/dev/null || {
    echo "    Skipped: aarch64-linux-gnu-gcc not found"
    echo "    To build ARM64 on x64, install: gcc-aarch64-linux-gnu"
}

# 创建 AppImage (如果 appimagetool 可用)
echo "[4/6] Creating AppImage..."
if command -v appimagetool &> /dev/null || [ -f "/usr/local/bin/appimagetool" ]; then
    APPIMAGE_DIR=$(mktemp -d)

    # 创建 AppDir 结构
    mkdir -p "$APPIMAGE_DIR/AppDir/usr/bin"
    mkdir -p "$APPIMAGE_DIR/AppDir/usr/share/applications"
    mkdir -p "$APPIMAGE_DIR/AppDir/usr/share/icons/hicolor/256x256/apps"

    # 复制二进制文件
    cp "$BUILD_DIR/${APP_NAME}-x64" "$APPIMAGE_DIR/AppDir/usr/bin/$APP_NAME"

    # 创建 .desktop 文件
    cat > "$APPIMAGE_DIR/AppDir/usr/share/applications/openclaw-installer.desktop" << EOF
[Desktop Entry]
Name=OpenClaw Installer
Exec=$APP_NAME
Icon=openclaw-installer
Type=Application
Categories=System;Settings;
Comment=Cross-platform AI assistant installer
EOF

    cp "$APPIMAGE_DIR/AppDir/usr/share/applications/openclaw-installer.desktop" \
       "$APPIMAGE_DIR/AppDir/"

    # 创建图标 (使用默认图标或从项目复制)
    touch "$APPIMAGE_DIR/AppDir/usr/share/icons/hicolor/256x256/apps/openclaw-installer.png"

    # 创建 AppRun
    cat > "$APPIMAGE_DIR/AppDir/AppRun" << 'EOF'
#!/bin/bash
SELF=$(readlink -f "$0")
HERE=${SELF%/*}
export PATH="${HERE}/usr/bin:${PATH}"
exec "${HERE}/usr/bin/openclaw-installer" "$@"
EOF
    chmod +x "$APPIMAGE_DIR/AppDir/AppRun"

    # 构建 AppImage
    if command -v appimagetool &> /dev/null; then
        appimagetool "$APPIMAGE_DIR/AppDir" "$BUILD_DIR/${APP_NAME}-x86_64.AppImage" 2>/dev/null || {
            echo "    Warning: AppImage creation failed"
        }
    fi

    rm -rf "$APPIMAGE_DIR"

    if [ -f "$BUILD_DIR/${APP_NAME}-x86_64.AppImage" ]; then
        echo "    Created: $BUILD_DIR/${APP_NAME}-x86_64.AppImage"
    fi
else
    echo "    Skipped: appimagetool not found"
fi

# 创建 DEB 包
echo "[5/6] Creating DEB package..."
if command -v dpkg-deb &> /dev/null; then
    DEB_DIR=$(mktemp -d)

    # 创建 deb 包结构
    mkdir -p "$DEB_DIR/DEBIAN"
    mkdir -p "$DEB_DIR/usr/bin"
    mkdir -p "$DEB_DIR/usr/share/applications"
    mkdir -p "$DEB_DIR/usr/share/doc/openclaw-installer"

    # 控制文件
    cat > "$DEB_DIR/DEBIAN/control" << EOF
Package: openclaw-installer
Version: $VERSION
Section: utils
Priority: optional
Architecture: amd64
Depends: libgtk-3-0, libwebkit2gtk-4.0-37
Maintainer: OpenClaw Team <team@openclaw.org>
Description: OpenClaw Installer
 Cross-platform AI assistant installer for OpenClaw.
EOF

    # 复制二进制文件
    cp "$BUILD_DIR/${APP_NAME}-x64" "$DEB_DIR/usr/bin/$APP_NAME"
    chmod 755 "$DEB_DIR/usr/bin/$APP_NAME"

    # 创建 .desktop 文件
    cat > "$DEB_DIR/usr/share/applications/openclaw-installer.desktop" << EOF
[Desktop Entry]
Name=OpenClaw Installer
Exec=/usr/bin/$APP_NAME
Type=Application
Categories=System;Settings;
Comment=Cross-platform AI assistant installer
Terminal=false
EOF
    chmod 644 "$DEB_DIR/usr/share/applications/openclaw-installer.desktop"

    # 创建 changelog
    cat > "$DEB_DIR/usr/share/doc/openclaw-installer/changelog" << EOF
openclaw-installer ($VERSION) stable; urgency=medium

  * Initial release

 -- OpenClaw Team <team@openclaw.org>  $(date -R)
EOF
    gzip -9 "$DEB_DIR/usr/share/doc/openclaw-installer/changelog"

    # 构建 deb 包
    dpkg-deb --build "$DEB_DIR" "$BUILD_DIR/${APP_NAME}_${VERSION}_amd64.deb" 2>/dev/null || {
        echo "    Warning: DEB package creation failed"
    }

    rm -rf "$DEB_DIR"

    if [ -f "$BUILD_DIR/${APP_NAME}_${VERSION}_amd64.deb" ]; then
        echo "    Created: $BUILD_DIR/${APP_NAME}_${VERSION}_amd64.deb"
    fi
else
    echo "    Skipped: dpkg-deb not found"
fi

# 创建 RPM 包 (如果 rpmbuild 可用)
echo "[6/6] Creating RPM package..."
if command -v rpmbuild &> /dev/null; then
    RPM_TOPDIR=$(mktemp -d)
    mkdir -p "$RPM_TOPDIR"/{BUILD,RPMS,SOURCES,SPECS,SRPMS}

    # 创建 spec 文件
    cat > "$RPM_TOPDIR/SPECS/openclaw-installer.spec" << EOF
Name:           openclaw-installer
Version:        $VERSION
Release:        1%{?dist}
Summary:        Cross-platform AI assistant installer
License:        MIT
BuildArch:      x86_64

%description
Cross-platform AI assistant installer for OpenClaw.

%install
mkdir -p %{buildroot}/usr/bin
mkdir -p %{buildroot}/usr/share/applications
cp $BUILD_DIR/${APP_NAME}-x64 %{buildroot}/usr/bin/$APP_NAME
chmod 755 %{buildroot}/usr/bin/$APP_NAME

cat > %{buildroot}/usr/share/applications/openclaw-installer.desktop << 'DESKTOP'
[Desktop Entry]
Name=OpenClaw Installer
Exec=/usr/bin/$APP_NAME
Type=Application
Categories=System;Settings;
Comment=Cross-platform AI assistant installer
Terminal=false
DESKTOP
chmod 644 %{buildroot}/usr/share/applications/openclaw-installer.desktop

%files
/usr/bin/$APP_NAME
/usr/share/applications/openclaw-installer.desktop

%changelog
* $(date "+%a %b %d %Y") OpenClaw Team <team@openclaw.org> - $VERSION-1
- Initial release
EOF

    # 构建 RPM
    rpmbuild --define "_topdir $RPM_TOPDIR" -bb "$RPM_TOPDIR/SPECS/openclaw-installer.spec" 2>/dev/null || {
        echo "    Warning: RPM package creation failed"
    }

    # 复制生成的 RPM
    find "$RPM_TOPDIR/RPMS" -name "*.rpm" -exec cp {} "$BUILD_DIR/" \;

    rm -rf "$RPM_TOPDIR"

    if ls "$BUILD_DIR/"/*.rpm 1> /dev/null 2>&1; then
        echo "    Created: $(ls "$BUILD_DIR/"/*.rpm | head -1)"
    fi
else
    echo "    Skipped: rpmbuild not found"
fi

# 创建 tar.gz 分发包
echo "[Extra] Creating tar.gz distribution..."
cd "$BUILD_DIR"
tar -czf "${APP_NAME}-linux-x64.tar.gz" "${APP_NAME}-x64" 2>/dev/null || true
if [ -f "${APP_NAME}-arm64" ]; then
    tar -czf "${APP_NAME}-linux-arm64.tar.gz" "${APP_NAME}-arm64" 2>/dev/null || true
fi

echo ""
echo "=========================================="
echo "Linux build complete!"
echo "Output: $BUILD_DIR"
echo "=========================================="
ls -lh "$BUILD_DIR/"
