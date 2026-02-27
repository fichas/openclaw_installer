#!/bin/bash

# OpenClaw Linux 安装器

clear
echo "============================================"
echo "   OpenClaw 跨平台 AI 助手安装器"
echo "============================================"
echo ""

# 检测架构
ARCH=$(uname -m)
if [ "$ARCH" = "x86_64" ]; then
    ARCH="amd64"
    echo "[INFO] 检测到系统: Linux 64位 (x64)"
elif [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then
    ARCH="arm64"
    echo "[INFO] 检测到系统: Linux ARM64"
else
    ARCH="amd64"
    echo "[WARNING] 无法检测架构，默认使用 x64 版本"
fi

# 获取脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

echo ""
echo "============================================"
echo "   请选择安装方式"
echo "============================================"
echo ""
echo " [1] 安装到系统目录 (推荐，需要sudo)"
echo "     路径: /usr/local/bin/"
echo "     所有用户可用，可创建systemd服务"
echo ""
echo " [2] 安装到用户目录"
echo "     路径: ~/.local/bin/"
echo "     无需sudo，仅当前用户可用"
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
        CONFIG_DIR="/etc/openclaw"

        echo "[INFO] 需要管理员权限..."
        if ! sudo -n true 2>/dev/null; then
            echo "[INFO] 请输入sudo密码:"
        fi

        sudo mkdir -p "$INSTALL_DIR"
        sudo mkdir -p "$CONFIG_DIR"
        sudo mkdir -p "$HOME/.openclaw"

        echo "[INFO] 复制安装文件..."
        sudo cp "${SCRIPT_DIR}/installers/openclaw-installer-linux-${ARCH}" "$INSTALL_DIR/openclaw"
        sudo chmod +x "$INSTALL_DIR/openclaw"

        echo "[INFO] 复制适配器配置..."
        sudo cp -r "${SCRIPT_DIR}/packages/config-templates/"* "$CONFIG_DIR/" 2>/dev/null || true
        cp -r "${SCRIPT_DIR}/packages/config-templates/"* "$HOME/.openclaw/" 2>/dev/null || true

        # 添加到 PATH（如果需要）
        if [[ ":$PATH:" != *":/usr/local/bin:"* ]]; then
            if [ -f "$HOME/.bashrc" ]; then
                echo 'export PATH="/usr/local/bin:$PATH"' >> "$HOME/.bashrc"
            fi
            if [ -f "$HOME/.zshrc" ]; then
                echo 'export PATH="/usr/local/bin:$PATH"' >> "$HOME/.zshrc"
            fi
            echo "[INFO] 已添加到 PATH，请重新登录或运行: source ~/.bashrc"
        fi

        # 可选：创建 systemd 服务
        read -p "是否创建 systemd 服务？(y/N): " create_service
        if [ "$create_service" = "y" ] || [ "$create_service" = "Y" ]; then
            echo "[INFO] 创建 systemd 服务..."
            sudo tee /etc/systemd/system/openclaw.service > /dev/null <<EOF
[Unit]
Description=OpenClaw AI Assistant
After=network.target

[Service]
Type=simple
User=$USER
ExecStart=/usr/local/bin/openclaw gateway
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF
            sudo systemctl daemon-reload
            echo "[INFO] 服务已创建，可使用以下命令管理："
            echo "  sudo systemctl start openclaw   # 启动"
            echo "  sudo systemctl enable openclaw  # 开机自启"
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
        cp "${SCRIPT_DIR}/installers/openclaw-installer-linux-${ARCH}" "$INSTALL_DIR/openclaw"
        chmod +x "$INSTALL_DIR/openclaw"

        echo "[INFO] 复制适配器配置..."
        cp -r "${SCRIPT_DIR}/packages/config-templates/"* "$CONFIG_DIR/"

        # 添加到 PATH（如果需要）
        if [[ ":$PATH:" != *":$HOME/.local/bin:"* ]]; then
            if [ -f "$HOME/.bashrc" ]; then
                echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$HOME/.bashrc"
            fi
            if [ -f "$HOME/.zshrc" ]; then
                echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$HOME/.zshrc"
            fi
            echo "[INFO] 已添加到 PATH，请重新登录或运行: source ~/.bashrc"
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
        "${SCRIPT_DIR}/installers/openclaw-installer-linux-${ARCH}" --portable &
        ;;
esac

echo ""
echo "安装程序已在后台运行。"
echo "如果浏览器没有自动打开，请手动访问: http://localhost:18080"
echo ""
read -p "按回车键关闭此窗口..."
