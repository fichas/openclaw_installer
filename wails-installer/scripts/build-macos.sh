#!/bin/bash

# macOS 构建脚本
# 支持 x64 (amd64) 和 ARM64 架构
# 创建 .app 应用包

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BUILD_DIR="$PROJECT_ROOT/build/macos"

# 版本信息
VERSION="${VERSION:-1.0.0}"
APP_NAME="OpenClaw-Installer"
BUNDLE_ID="org.openclaw.installer"

echo "=========================================="
echo "Building macOS installer"
echo "Version: $VERSION"
echo "=========================================="

cd "$PROJECT_ROOT"

# 创建构建目录
mkdir -p "$BUILD_DIR"

# 下载依赖
echo "[1/6] Downloading dependencies..."
go mod tidy

# 构建函数
build_macos() {
    local ARCH=$1
    local OUTPUT_DIR=$2
    local GOARCH

    case $ARCH in
        x64) GOARCH="amd64" ;;
        arm64) GOARCH="arm64" ;;
        *) echo "Unknown architecture: $ARCH"; exit 1 ;;
    esac

    echo "[2/6] Building macOS $ARCH..."

    # 创建 .app 目录结构
    APP_BUNDLE="$OUTPUT_DIR/${APP_NAME}.app"
    mkdir -p "$APP_BUNDLE/Contents/MacOS"
    mkdir -p "$APP_BUNDLE/Contents/Resources"

    # 构建二进制文件
    CGO_ENABLED=1 GOOS=darwin GOARCH=$GOARCH \
        go build -ldflags "-s -w -X main.version=$VERSION" \
        -o "$APP_BUNDLE/Contents/MacOS/$APP_NAME" .

    # 创建 Info.plist
    cat > "$APP_BUNDLE/Contents/Info.plist" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleDevelopmentRegion</key>
    <string>en</string>
    <key>CFBundleExecutable</key>
    <string>${APP_NAME}</string>
    <key>CFBundleIdentifier</key>
    <string>${BUNDLE_ID}</string>
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
    <key>LSUIElement</key>
    <false/>
</dict>
</plist>
EOF

    # 创建 PkgInfo
    echo "APPL????" > "$APP_BUNDLE/Contents/PkgInfo"

    echo "    Created: $APP_BUNDLE"
}

# 构建 x64 版本
build_macos "x64" "$BUILD_DIR"
mv "$BUILD_DIR/${APP_NAME}.app" "$BUILD_DIR/${APP_NAME}-x64.app"

# 构建 ARM64 版本
build_macos "arm64" "$BUILD_DIR"
mv "$BUILD_DIR/${APP_NAME}.app" "$BUILD_DIR/${APP_NAME}-arm64.app"

# 创建通用二进制 (Universal Binary)
echo "[3/6] Creating Universal Binary..."
if command -v lipo &> /dev/null; then
    UNIVERSAL_APP="$BUILD_DIR/${APP_NAME}.app"
    mkdir -p "$UNIVERSAL_APP/Contents/MacOS"
    mkdir -p "$UNIVERSAL_APP/Contents/Resources"

    # 使用 lipo 合并两个架构
    lipo -create \
        "$BUILD_DIR/${APP_NAME}-x64.app/Contents/MacOS/$APP_NAME" \
        "$BUILD_DIR/${APP_NAME}-arm64.app/Contents/MacOS/$APP_NAME" \
        -output "$UNIVERSAL_APP/Contents/MacOS/$APP_NAME"

    # 复制 Info.plist
    cp "$BUILD_DIR/${APP_NAME}-x64.app/Contents/Info.plist" "$UNIVERSAL_APP/Contents/Info.plist"
    cp "$BUILD_DIR/${APP_NAME}-x64.app/Contents/PkgInfo" "$UNIVERSAL_APP/Contents/PkgInfo"

    # 设置可执行权限
    chmod +x "$UNIVERSAL_APP/Contents/MacOS/$APP_NAME"

    echo "    Created: $UNIVERSAL_APP"
else
    echo "    Skipped: lipo not found (requires macOS)"
fi

# 签名应用 (如果在 macOS 上且有证书)
echo "[4/6] Signing application..."
if [ "$(uname)" = "Darwin" ]; then
    if [ -n "$APPLE_DEVELOPER_ID" ]; then
        codesign --force --deep --sign "$APPLE_DEVELOPER_ID" \
            "$BUILD_DIR/${APP_NAME}.app" 2>/dev/null || echo "    Warning: codesign failed"
        echo "    Signed: $BUILD_DIR/${APP_NAME}.app"
    else
        # 临时签名
        codesign --force --deep --sign - "$BUILD_DIR/${APP_NAME}.app" 2>/dev/null || true
        echo "    Ad-hoc signed"
    fi
else
    echo "    Skipped: not running on macOS"
fi

# 创建 DMG (如果在 macOS 上)
echo "[5/6] Creating DMG..."
if [ "$(uname)" = "Darwin" ]; then
    if command -v hdiutil &> /dev/null; then
        # 创建临时目录
        DMG_TEMP=$(mktemp -d)
        cp -R "$BUILD_DIR/${APP_NAME}.app" "$DMG_TEMP/"
        ln -s /Applications "$DMG_TEMP/Applications"

        # 创建 DMG
        hdiutil create -volname "OpenClaw Installer" \
            -srcfolder "$DMG_TEMP" \
            -ov -format UDZO \
            "$BUILD_DIR/${APP_NAME}-macOS.dmg"

        rm -rf "$DMG_TEMP"
        echo "    Created: $BUILD_DIR/${APP_NAME}-macOS.dmg"
    else
        echo "    Skipped: hdiutil not found"
    fi
else
    echo "    Skipped: DMG creation requires macOS"
fi

# 创建 ZIP 分发包
echo "[6/6] Creating ZIP distribution..."
cd "$BUILD_DIR"
zip -9 -r "${APP_NAME}-macOS-x64.zip" "${APP_NAME}-x64.app" 2>/dev/null || true
zip -9 -r "${APP_NAME}-macOS-arm64.zip" "${APP_NAME}-arm64.app" 2>/dev/null || true
if [ -d "${APP_NAME}.app" ]; then
    zip -9 -r "${APP_NAME}-macOS-universal.zip" "${APP_NAME}.app" 2>/dev/null || true
fi

echo ""
echo "=========================================="
echo "macOS build complete!"
echo "Output: $BUILD_DIR"
echo "=========================================="
ls -lh "$BUILD_DIR/"
