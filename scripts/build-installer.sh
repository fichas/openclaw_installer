#!/bin/bash
set -e

echo "=== 构建 OpenClaw 安装器 ==="
cd "$(dirname "$0")/.."

export ELECTRON_MIRROR=https://npmmirror.com/mirrors/electron/
export ELECTRON_BUILDER_BINARIES_MIRROR=https://npmmirror.com/mirrors/electron-builder-binaries/

pnpm --filter @openclaw/installer build
echo "=== 安装器构建完成 ==="
