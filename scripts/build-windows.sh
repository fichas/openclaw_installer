#!/bin/bash

# OpenClaw Windows 无窗口构建脚本
# 此脚本专门用于构建 Windows 版本，确保运行时不会显示 CMD 窗口

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  OpenClaw Windows 无窗口构建脚本${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# 检查依赖
check_dependency() {
    if ! command -v "$1" &> /dev/null; then
        echo -e "${RED}错误: 未找到 $1${NC}"
        return 1
    fi
    return 0
}

# 检查 Go
echo -e "${BLUE}检查依赖...${NC}"
if ! check_dependency "go"; then
    echo -e "${RED}请先安装 Go: https://golang.org/dl/${NC}"
    exit 1
fi

# 检查是否安装了 mingw-w64（用于 Windows 交叉编译）
if command -v "x86_64-w64-mingw32-gcc" &> /dev/null; then
    echo -e "${GREEN}找到 mingw-w64 交叉编译器${NC}"
    HAS_MINGW=true
else
    echo -e "${YELLOW}警告: 未找到 mingw-w64 交叉编译器${NC}"
    echo -e "${YELLOW}Windows 交叉编译需要 mingw-w64${NC}"
    HAS_MINGW=false
fi

# 解析参数
BUILD_ARCH="amd64"
BUILD_DEBUG=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --arch)
            BUILD_ARCH="$2"
            shift 2
            ;;
        --debug)
            BUILD_DEBUG=true
            shift
            ;;
        --help|-h)
            echo "用法: $0 [选项]"
            echo ""
            echo "选项:"
            echo "  --arch <arch>    目标架构: amd64 或 arm64 (默认: amd64)"
            echo "  --debug          启用调试模式（保留控制台窗口用于调试）"
            echo "  --help, -h       显示此帮助信息"
            echo ""
            echo "示例:"
            echo "  $0                           # 构建 Windows amd64 版本"
            echo "  $0 --arch arm64              # 构建 Windows arm64 版本"
            echo "  $0 --debug                   # 构建调试版本（带控制台）"
            exit 0
            ;;
        *)
            echo -e "${RED}未知选项: $1${NC}"
            exit 1
            ;;
    esac
done

# 验证架构
if [[ "$BUILD_ARCH" != "amd64" && "$BUILD_ARCH" != "arm64" ]]; then
    echo -e "${RED}错误: 不支持的架构: $BUILD_ARCH${NC}"
    echo -e "${RED}支持的架构: amd64, arm64${NC}"
    exit 1
fi

echo ""
echo -e "${BLUE}构建配置:${NC}"
echo "  目标平台: Windows"
echo "  目标架构: $BUILD_ARCH"
echo "  调试模式: $BUILD_DEBUG"
echo ""

# 进入 wails-installer 目录
cd "$PROJECT_ROOT/wails-installer"

# 下载依赖
echo -e "${BLUE}下载 Go 依赖...${NC}"
go mod tidy

# 构建输出目录
OUTPUT_DIR="$PROJECT_ROOT/dist/windows-$BUILD_ARCH"
mkdir -p "$OUTPUT_DIR"

# 构建参数
BINARY_NAME="OpenClaw-Installer.exe"
OUTPUT_PATH="$OUTPUT_DIR/$BINARY_NAME"

# 设置构建标志
# -s: 省略符号表
# -w: 省略 DWARF 调试信息
# -H=windowsgui: 关键标志！使程序作为 Windows GUI 应用运行，不显示 CMD 窗口
if [ "$BUILD_DEBUG" = true ]; then
    # 调试模式：不隐藏控制台窗口
    LDFLAGS="-s -w"
    echo -e "${YELLOW}调试模式: 保留控制台窗口${NC}"
else
    # 发布模式：隐藏控制台窗口
    LDFLAGS="-s -w -H=windowsgui"
    echo -e "${GREEN}发布模式: 隐藏 CMD 窗口 (-H=windowsgui)${NC}"
fi

# 执行构建
echo ""
echo -e "${BLUE}开始构建...${NC}"
echo "  输出: $OUTPUT_PATH"
echo ""

export GOOS=windows
export GOARCH=$BUILD_ARCH
export CGO_ENABLED=1

# 根据架构设置编译器
if [ "$BUILD_ARCH" = "amd64" ]; then
    if [ "$HAS_MINGW" = true ]; then
        export CC=x86_64-w64-mingw32-gcc
        export CXX=x86_64-w64-mingw32-g++
    fi
elif [ "$BUILD_ARCH" = "arm64" ]; then
    if [ "$HAS_MINGW" = true ]; then
        export CC=aarch64-w64-mingw32-gcc
        export CXX=aarch64-w64-mingw32-g++
    fi
fi

# 构建
if go build -ldflags "$LDFLAGS" -o "$OUTPUT_PATH" .; then
    echo ""
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}  构建成功!${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo ""
    echo "输出文件: $OUTPUT_PATH"
    echo ""

    # 显示文件大小
    if command -v "ls" &> /dev/null; then
        ls -lh "$OUTPUT_PATH" 2>/dev/null || true
    fi

    echo ""
    echo -e "${BLUE}Windows 无窗口特性:${NC}"
    echo "  - 双击运行时不会显示 CMD 窗口"
    echo "  - 所有日志输出到文件: %APPDATA%\\OpenClaw\\logs\\installer.log"
    echo "  - 适用于 Windows 10/11 x64/ARM64"
    echo ""

    if [ "$BUILD_DEBUG" = false ]; then
        echo -e "${YELLOW}注意: 如果程序崩溃，使用 --debug 标志构建以查看错误信息${NC}"
    fi

    exit 0
else
    echo ""
    echo -e "${RED}========================================${NC}"
    echo -e "${RED}  构建失败!${NC}"
    echo -e "${RED}========================================${NC}"
    echo ""
    echo -e "${YELLOW}提示: 如果在 Linux/macOS 上交叉编译 Windows 版本，${NC}"
    echo -e "${YELLOW}      需要安装 mingw-w64:${NC}"
    echo ""
    echo "  Ubuntu/Debian: sudo apt-get install mingw-w64"
    echo "  macOS:         brew install mingw-w64"
    echo "  Fedora:        sudo dnf install mingw64-gcc-c++"
    echo ""
    exit 1
fi
