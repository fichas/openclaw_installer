#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"

cd "$ROOT_DIR"

case "${1:-all}" in
  installer)
    bash "$SCRIPT_DIR/build-installer.sh"
    ;;
  config)
    bash "$SCRIPT_DIR/build-config.sh"
    ;;
  all)
    bash "$SCRIPT_DIR/build-installer.sh"
    bash "$SCRIPT_DIR/build-config.sh"
    ;;
  *)
    echo "用法: $0 {installer|config|all}"
    exit 1
    ;;
esac
