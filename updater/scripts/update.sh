#!/bin/bash
#
# OpenClaw Updater - Linux/macOS 更新脚本包装器
# 此脚本用于包装 Go 更新程序，提供额外的检查和便利功能
#

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 脚本目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# 检测平台
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

# 转换架构名称
case "$ARCH" in
    x86_64)
        ARCH="amd64"
        ;;
    arm64|aarch64)
        ARCH="arm64"
        ;;
esac

# 更新程序二进制文件名
UPDATER_BIN="openclaw-updater-${OS}-${ARCH}"

# 检查更新程序是否存在
if [[ ! -f "$SCRIPT_DIR/$UPDATER_BIN" ]]; then
    # 尝试使用通用名称
    if [[ -f "$SCRIPT_DIR/openclaw-updater" ]]; then
        UPDATER_BIN="openclaw-updater"
    else
        echo -e "${RED}Error: Updater binary not found: $UPDATER_BIN${NC}"
        exit 1
    fi
fi

# 检查权限
if [[ "$OS" == "linux" ]] && [[ "$EUID" -ne 0 ]]; then
    echo -e "${YELLOW}Warning: This script may require root privileges to update system files.${NC}"
    echo "Consider running with sudo."
fi

# 执行更新程序
echo -e "${GREEN}OpenClaw Updater${NC}"
echo "Platform: $OS/$ARCH"
echo ""

exec "$SCRIPT_DIR/$UPDATER_BIN" "$@"
