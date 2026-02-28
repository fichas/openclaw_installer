#!/bin/bash
set -e

echo "=== 构建 OpenClaw 配置服务 ==="
cd "$(dirname "$0")/.."

pnpm --filter @openclaw/config-server build
echo "=== 配置服务构建完成 ==="
