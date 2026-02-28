#!/bin/bash
#
# GitHub Actions Release 检查脚本
# 使用方法: ./scripts/check-release.sh [TAG]
#

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 获取仓库信息
REPO=$(git remote get-url origin 2>/dev/null | sed 's/.*github.com[:/]//;s/\.git$//' || echo "")
if [ -z "$REPO" ]; then
    echo -e "${RED}错误: 无法获取远程仓库信息${NC}"
    exit 1
fi

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  OpenClaw Release 检查工具${NC}"
echo -e "${BLUE}  仓库: $REPO${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# 检查 gh CLI
if ! command -v gh &> /dev/null; then
    echo -e "${YELLOW}提示: 安装 GitHub CLI 可获得更好的体验${NC}"
    echo "  安装: https://cli.github.com/"
    echo ""
fi

# 获取最近的标签
LATEST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "")
if [ -z "$LATEST_TAG" ]; then
    echo -e "${YELLOW}警告: 没有找到标签${NC}"
    echo "创建标签: git tag -a v1.0.0 -m 'Release v1.0.0'"
    echo "推送标签: git push origin v1.0.0"
    exit 1
fi

TARGET_TAG=${1:-$LATEST_TAG}
echo -e "检查标签: ${GREEN}$TARGET_TAG${NC}"
echo ""

# 方法1: 使用 gh CLI (推荐)
if command -v gh &> /dev/null; then
    echo -e "${BLUE}使用 GitHub CLI 检查...${NC}"
    echo ""

    # 检查 workflow 运行状态
    echo -e "${YELLOW}最近的 Workflow 运行:${NC}"
    gh run list --repo "$REPO" --limit 5 --json databaseId,headBranch,status,conclusion,createdAt,event 2>/dev/null | \
        jq -r '.[] | "  \(.status) | \(.event) | \(.createdAt)"' 2>/dev/null || \
        gh run list --limit 5
    echo ""

    # 检查 release
    echo -e "${YELLOW}Release 信息:${NC}"
    if gh release view "$TARGET_TAG" --repo "$REPO" 2>/dev/null; then
        echo ""
        echo -e "${GREEN}✓ Release $TARGET_TAG 已发布!${NC}"
        echo ""
        echo -e "${YELLOW}下载地址:${NC}"
        gh release view "$TARGET_TAG" --repo "$REPO" --json assets 2>/dev/null | \
            jq -r '.assets[] | "  \(.name): \(.url)"' 2>/dev/null || \
            echo "  查看: https://github.com/$REPO/releases/tag/$TARGET_TAG"
    else
        echo -e "${YELLOW}Release 尚未创建${NC}"
        echo "  正在等待 GitHub Actions 完成..."
    fi

else
    # 方法2: 使用 curl
    echo -e "${BLUE}使用 HTTP API 检查...${NC}"
    echo ""

    # 检查 workflow 运行
    echo -e "${YELLOW}Workflow 状态:${NC}"
    curl -s "https://api.github.com/repos/$REPO/actions/runs?per_page=5" | \
        jq -r '.workflow_runs[] | "  \(.status): \(.name) (\(.head_branch))"' 2>/dev/null || \
        echo "  请手动查看: https://github.com/$REPO/actions"
    echo ""

    # 检查 release
    echo -e "${YELLOW}Release 状态:${NC}"
    RELEASE_INFO=$(curl -s "https://api.github.com/repos/$REPO/releases/tags/$TARGET_TAG")

    if echo "$RELEASE_INFO" | jq -e '.tag_name' >/dev/null 2>&1; then
        echo -e "${GREEN}✓ Release $TARGET_TAG 已发布!${NC}"
        echo ""
        echo -e "${YELLOW}发布资源:${NC}"
        echo "$RELEASE_INFO" | jq -r '.assets[] | "  - \(.name) (\(.size) bytes)"' 2>/dev/null || true
        echo ""
        echo -e "${YELLOW}下载页面:${NC}"
        echo "  https://github.com/$REPO/releases/tag/$TARGET_TAG"
    else
        echo -e "${YELLOW}Release 尚未创建${NC}"
        echo "  请检查 Actions 状态: https://github.com/$REPO/actions"
    fi
fi

echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}快捷链接:${NC}"
echo "  Actions: https://github.com/$REPO/actions"
echo "  Releases: https://github.com/$REPO/releases"
echo -e "${BLUE}========================================${NC}"
