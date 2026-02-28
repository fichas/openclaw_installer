#!/bin/bash
#
# GitHub Actions 构建监控脚本
# 使用方法: ./scripts/watch-build.sh
#

REPO="fichas/openclaw_installer"

echo "========================================"
echo "  OpenClaw 构建监控"
echo "========================================"
echo ""
echo "监控仓库: $REPO"
echo ""

# 获取最新的运行
LATEST=$(curl -s "https://api.github.com/repos/$REPO/actions/runs?per_page=1" 2>/dev/null)

if [ -z "$LATEST" ]; then
    echo "无法获取构建状态，请检查网络连接"
    exit 1
fi

RUN_NUMBER=$(echo "$LATEST" | grep -o '"run_number": [0-9]*' | head -1 | cut -d' ' -f2)
STATUS=$(echo "$LATEST" | grep -o '"status": "[^"]*"' | head -1 | cut -d'"' -f4)
CONCLUSION=$(echo "$LATEST" | grep -o '"conclusion": "[^"]*"' | head -1 | cut -d'"' -f4)
MESSAGE=$(echo "$LATEST" | grep -o '"message": "[^"]*"' | head -1 | cut -d'"' -f4)

echo "最新构建: #$RUN_NUMBER"
echo "提交信息: $MESSAGE"
echo ""

if [ "$STATUS" = "completed" ]; then
    if [ "$CONCLUSION" = "success" ]; then
        echo "✅ 构建成功!"
        echo ""
        echo "下载地址:"
        echo "  https://github.com/$REPO/releases"
    else
        echo "❌ 构建失败: $CONCLUSION"
        echo ""
        echo "查看日志:"
        echo "  https://github.com/$REPO/actions"
    fi
else
    echo "⏳ 构建进行中..."
    echo ""
    echo "实时查看:"
    echo "  https://github.com/$REPO/actions"
    echo ""
    echo "按 Ctrl+C 停止监控"
    echo ""

    # 持续监控
    while true; do
        sleep 10
        LATEST=$(curl -s "https://api.github.com/repos/$REPO/actions/runs?per_page=1" 2>/dev/null)
        NEW_STATUS=$(echo "$LATEST" | grep -o '"status": "[^"]*"' | head -1 | cut -d'"' -f4)
        NEW_CONCLUSION=$(echo "$LATEST" | grep -o '"conclusion": "[^"]*"' | head -1 | cut -d'"' -f4)

        if [ "$NEW_STATUS" = "completed" ]; then
            echo ""
            if [ "$NEW_CONCLUSION" = "success" ]; then
                echo "✅ 构建成功!"
                echo ""
                echo "下载地址:"
                echo "  https://github.com/$REPO/releases"
            else
                echo "❌ 构建失败: $NEW_CONCLUSION"
                echo ""
                echo "查看日志:"
                echo "  https://github.com/$REPO/actions"
            fi
            break
        fi
        echo -n "."
    done
fi

echo ""
echo "========================================"
