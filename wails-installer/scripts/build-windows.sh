#!/bin/bash

# Windows 构建脚本
# 支持 x64 和 ARM64 架构

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BUILD_DIR="$PROJECT_ROOT/build/windows"

# 版本信息
VERSION="${VERSION:-1.0.0}"
APP_NAME="OpenClaw-Installer"

echo "=========================================="
echo "Building Windows installer"
echo "Version: $VERSION"
echo "=========================================="

cd "$PROJECT_ROOT"

# 创建构建目录
mkdir -p "$BUILD_DIR"

# 下载依赖
echo "[1/5] Downloading dependencies..."
go mod tidy

# 构建 Windows x64
echo "[2/5] Building Windows x64..."
GOOS=windows GOARCH=amd64 CGO_ENABLED=1 \
    CC=x86_64-w64-mingw32-gcc \
    go build -ldflags "-s -w -H=windowsgui -X main.version=$VERSION" \
    -o "$BUILD_DIR/${APP_NAME}-x64.exe" .

echo "    Created: $BUILD_DIR/${APP_NAME}-x64.exe"

# 构建 Windows ARM64 (如果工具链可用)
echo "[3/5] Building Windows ARM64..."
if command -v aarch64-w64-mingw32-gcc &> /dev/null; then
    GOOS=windows GOARCH=arm64 CGO_ENABLED=1 \
        CC=aarch64-w64-mingw32-gcc \
        go build -ldflags "-s -w -H=windowsgui -X main.version=$VERSION" \
        -o "$BUILD_DIR/${APP_NAME}-arm64.exe" .
    echo "    Created: $BUILD_DIR/${APP_NAME}-arm64.exe"
else
    echo "    Skipped: aarch64-w64-mingw32-gcc not found"
fi

# 创建 NSIS 安装程序 (如果 nsis 可用)
echo "[4/5] Creating NSIS installer..."
if command -v makensis &> /dev/null; then
    cat > "$BUILD_DIR/installer.nsi" << 'NSIS_SCRIPT'
!include "MUI2.nsh"
!include "FileFunc.nsh"

Name "OpenClaw Installer"
OutFile "OpenClaw-Installer-Windows.exe"
InstallDir "$PROGRAMFILES64\OpenClaw"
RequestExecutionLevel admin

!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH

!insertmacro MUI_UNPAGE_WELCOME
!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES
!insertmacro MUI_UNPAGE_FINISH

!insertmacro MUI_LANGUAGE "English"

Section "Install"
    SetOutPath "$INSTDIR"

    ; 检测架构并安装对应版本
    ${If} ${RunningX64}
        File "OpenClaw-Installer-x64.exe"
        Rename "$INSTDIR\OpenClaw-Installer-x64.exe" "$INSTDIR\OpenClaw-Installer.exe"
    ${Else}
        File "OpenClaw-Installer-arm64.exe"
        Rename "$INSTDIR\OpenClaw-Installer-arm64.exe" "$INSTDIR\OpenClaw-Installer.exe"
    ${EndIf}

    ; 创建开始菜单快捷方式
    CreateDirectory "$SMPROGRAMS\OpenClaw"
    CreateShortcut "$SMPROGRAMS\OpenClaw\OpenClaw Installer.lnk" "$INSTDIR\OpenClaw-Installer.exe"

    ; 创建卸载程序
    WriteUninstaller "$INSTDIR\Uninstall.exe"

    ; 注册卸载信息
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\OpenClaw" \
        "DisplayName" "OpenClaw Installer"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\OpenClaw" \
        "UninstallString" "$INSTDIR\Uninstall.exe"
SectionEnd

Section "Uninstall"
    Delete "$INSTDIR\OpenClaw-Installer.exe"
    Delete "$INSTDIR\Uninstall.exe"
    Delete "$SMPROGRAMS\OpenClaw\OpenClaw Installer.lnk"
    RMDir "$SMPROGRAMS\OpenClaw"
    RMDir "$INSTDIR"
    DeleteRegKey HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\OpenClaw"
SectionEnd
NSIS_SCRIPT

    cd "$BUILD_DIR"
    makensis installer.nsi
    echo "    Created: $BUILD_DIR/OpenClaw-Installer-Windows.exe"
else
    echo "    Skipped: makensis not found"
fi

# 创建 ZIP 分发包
echo "[5/5] Creating ZIP distribution..."
cd "$BUILD_DIR"
zip -9 "${APP_NAME}-Windows-x64.zip" "${APP_NAME}-x64.exe" 2>/dev/null || true
if [ -f "${APP_NAME}-arm64.exe" ]; then
    zip -9 "${APP_NAME}-Windows-arm64.zip" "${APP_NAME}-arm64.exe" 2>/dev/null || true
fi

echo ""
echo "=========================================="
echo "Windows build complete!"
echo "Output: $BUILD_DIR"
echo "=========================================="
ls -lh "$BUILD_DIR/"
