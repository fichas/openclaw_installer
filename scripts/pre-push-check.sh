#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"

echo "[1/5] Node workspace tests"
cd "$ROOT_DIR"
pnpm test

echo "[2/5] Installer type check"
pnpm --filter @openclaw/installer test

echo "[3/5] Config server build"
pnpm build:config

echo "[4/5] Go updater tests"
cd "$ROOT_DIR/updater"
go test ./...

echo "[5/5] Go updater race tests"
go test -race ./...

echo "All pre-push checks passed."
