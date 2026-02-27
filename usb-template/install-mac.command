#!/bin/bash

# OpenClaw macOS 安装器

clear
echo "============================================"
echo "   OpenClaw 跨平台 AI 助手安装器"
echo "============================================"
echo ""

# 检测架构
ARCH=$(uname -m)
if [ "$ARCH" = "x86_64" ]; then
    ARCH="amd64"
    echo "[INFO] 检测到系统: macOS Intel (x64)"
elif [ "$ARCH" = "arm64" ]; then
    echo "[INFO] 检测到系统: macOS Apple Silicon (ARM64)"
else
    ARCH="amd64"
    echo "[WARNING] 无法检测架构，默认使用 x64 版本"
fi

# 获取脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

# 如果通过双击运行，SCRIPT_DIR 可能是 /Users/xxx 而不是 U 盘路径
# 尝试找到实际的 U 盘路径
if [[ "$SCRIPT_DIR" == *"/Users/"* ]] || [[ "$SCRIPT_DIR" == *"/home/"* ]]; then
    # 尝试从 Volumes 中找到 OpenClaw
    for vol in /Volumes/*/; do
        if [ -f "${vol}install-mac.command" ]; then
            SCRIPT_DIR="$vol"
            break
        fi
    done
fi

echo ""
echo "============================================"
echo "   请选择安装方式"
echo "============================================"
echo ""
echo " [1] 安装到系统目录 (推荐)"
echo "     路径: /usr/local/bin/"
echo "     需要管理员密码，所有用户可用"
echo ""
echo " [2] 安装到用户目录"
echo "     路径: ~/.local/bin/"
echo "     无需管理员权限，仅当前用户可用"
echo ""
echo " [3] 仅运行，不安装 (便携模式)"
echo "     直接从U盘运行，不复制到系统"
echo ""

read -p "请输入选项 (1/2/3): " choice

case $choice in
    1)
        echo ""
        echo "[INFO] 准备安装到系统目录..."
        INSTALL_DIR="/usr/local/bin"
        CONFIG_DIR="/usr/local/etc/openclaw"

        echo "[INFO] 需要管理员权限..."
        if ! sudo -n true 2>/dev/null; then
            echo "[INFO] 请输入管理员密码:"
        fi

        sudo mkdir -p "$INSTALL_DIR"
        sudo mkdir -p "$CONFIG_DIR"

        echo "[INFO] 复制安装文件..."
        sudo cp "${SCRIPT_DIR}/installers/openclaw-installer-darwin-${ARCH}" "$INSTALL_DIR/openclaw"
        sudo chmod +x "$INSTALL_DIR/openclaw"

        echo "[INFO] 复制适配器配置..."
        sudo cp -r "${SCRIPT_DIR}/packages/config-templates/"* "$CONFIG_DIR/"

        # 添加到 PATH（如果需要）
        if [[ ":$PATH:" != *":/usr/local/bin:"* ]]; then
            echo 'export PATH="/usr/local/bin:$PATH"' >> ~/.zshrc
            echo "[INFO] 已添加到 PATH，请运行: source ~/.zshrc"
        fi

        echo ""
        echo "============================================"
        echo "   安装完成！"
        echo "============================================"
        echo ""
        echo "[INFO] 正在启动配置向导..."
        echo "[INFO] 浏览器将自动打开 http://localhost:18080"
        echo ""

        sudo "$INSTALL_DIR/openclaw" &
        ;;

    2)
        echo ""
        echo "[INFO] 安装到用户目录..."
        INSTALL_DIR="$HOME/.local/bin"
        CONFIG_DIR="$HOME/.openclaw"

        mkdir -p "$INSTALL_DIR"
        mkdir -p "$CONFIG_DIR"

        echo "[INFO] 复制安装文件..."
        cp "${SCRIPT_DIR}/installers/openclaw-installer-darwin-${ARCH}" "$INSTALL_DIR/openclaw"
        chmod +x "$INSTALL_DIR/openclaw"

        echo "[INFO] 复制适配器配置..."
        cp -r "${SCRIPT_DIR}/packages/config-templates/"* "$CONFIG_DIR/"

        # 添加到 PATH（如果需要）
        if [[ ":$PATH:" != *":$HOME/.local/bin:"* ]]; then
            echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc
            echo "[INFO] 已添加到 PATH，请运行: source ~/.zshrc"
        fi

        echo ""
        echo "============================================"
        echo "   安装完成！"
        echo "============================================"
        echo ""
        echo "[INFO] 正在启动配置向导..."
        echo "[INFO] 浏览器将自动打开 http://localhost:18080"
        echo ""

        "$INSTALL_DIR/openclaw" &
        ;;

    3|*)
        echo ""
        echo "[INFO] 便携模式 - 直接从U盘运行"
        echo "[INFO] 启动安装器..."
        echo ""
        "${SCRIPT_DIR}/installers/openclaw-installer-darwin-${ARCH}" --portable &
        ;;
esac

echo ""
echo "安装程序已在后台运行。"
echo "如果浏览器没有自动打开，请手动访问: http://localhost:18080"
echo ""
read -p "按回车键关闭此窗口..."
