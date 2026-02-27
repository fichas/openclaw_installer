#!/bin/bash
#
# OpenClaw Wails Installer Build Script
# Builds for Windows, macOS, and Linux
#

set -e

echo "================================"
echo "OpenClaw Wails Installer Builder"
echo "================================"

cd "$(dirname "$0")"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Version info
VERSION=${VERSION:-"1.0.0"}
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

LDFLAGS="-X main.Version=$VERSION -X main.BuildTime=$BUILD_TIME -X main.GitCommit=$GIT_COMMIT -s -w"

# Create output directory
mkdir -p dist

# Function to build for a platform
build_platform() {
    local os=$1
    local arch=$2
    local output_name=$3
    local extra_flags=$4

    echo -e "${BLUE}Building for $os/$arch...${NC}"

    GOOS=$os GOARCH=$arch CGO_ENABLED=0 \
        go build -ldflags "$LDFLAGS $extra_flags" \
        -o "dist/$output_name" .

    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Built: dist/$output_name${NC}"
    else
        echo -e "${RED}✗ Failed to build for $os/$arch${NC}"
        return 1
    fi
}

# Download dependencies
echo -e "${YELLOW}Downloading dependencies...${NC}"
go mod tidy

# Parse arguments
BUILD_ALL=false
BUILD_WINDOWS=false
BUILD_MACOS=false
BUILD_LINUX=false

if [ $# -eq 0 ]; then
    BUILD_ALL=true
else
    for arg in "$@"; do
        case $arg in
            --windows) BUILD_WINDOWS=true ;;
            --macos|--darwin) BUILD_MACOS=true ;;
            --linux) BUILD_LINUX=true ;;
            --all) BUILD_ALL=true ;;
            *) echo "Unknown option: $arg"; exit 1 ;;
        esac
    done
fi

# Build all platforms
if [ "$BUILD_ALL" = true ]; then
    BUILD_WINDOWS=true
    BUILD_MACOS=true
    BUILD_LINUX=true
fi

# Build Windows
if [ "$BUILD_WINDOWS" = true ]; then
    echo ""
    echo -e "${YELLOW}=== Windows Builds ===${NC}"
    # Windows GUI (no console window)
    build_platform "windows" "amd64" "OpenClaw-Installer-windows-amd64.exe" "-H=windowsgui"
    build_platform "windows" "arm64" "OpenClaw-Installer-windows-arm64.exe" "-H=windowsgui"
fi

# Build macOS
if [ "$BUILD_MACOS" = true ]; then
    echo ""
    echo -e "${YELLOW}=== macOS Builds ===${NC}"
    build_platform "darwin" "amd64" "OpenClaw-Installer-darwin-amd64"
    build_platform "darwin" "arm64" "OpenClaw-Installer-darwin-arm64"
fi

# Build Linux
if [ "$BUILD_LINUX" = true ]; then
    echo ""
    echo -e "${YELLOW}=== Linux Builds ===${NC}"
    build_platform "linux" "amd64" "OpenClaw-Installer-linux-amd64"
    build_platform "linux" "arm64" "OpenClaw-Installer-linux-arm64"
fi

# Summary
echo ""
echo -e "${GREEN}================================${NC}"
echo -e "${GREEN}Build Complete!${NC}"
echo -e "${GREEN}================================${NC}"
echo ""
echo "Output directory: ./dist/"
echo ""
ls -lh dist/
