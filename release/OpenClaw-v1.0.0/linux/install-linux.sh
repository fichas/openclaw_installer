#!/bin/bash
#
# OpenClaw Installer for Linux
# Version: 1.0.0
#

set -e

echo "========================================"
echo "OpenClaw Installer for Linux"
echo "Version: 1.0.0"
echo "========================================"
echo ""

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"

# Auto-detect architecture
ARCH=$(uname -m)
INSTALLER=""

case "$ARCH" in
    x86_64)
        INSTALLER="openclaw-installer-linux-amd64"
        echo "Detected architecture: x86_64 (AMD64)"
        ;;
    aarch64|arm64)
        INSTALLER="openclaw-installer-linux-arm64"
        echo "Detected architecture: ARM64"
        ;;
    *)
        echo "Unsupported architecture: $ARCH"
        echo "Supported architectures: x86_64, aarch64/arm64"
        read -p "Press Enter to exit..."
        exit 1
        ;;
esac

echo ""

# Check if installer exists
if [ ! -f "$INSTALLER" ]; then
    echo "Error: Installer not found: $INSTALLER"
    echo "Available files:"
    ls -1 "$SCRIPT_DIR"/
    read -p "Press Enter to exit..."
    exit 1
fi

# Make sure installer is executable
chmod +x "$INSTALLER"

echo "Starting OpenClaw Installer..."
echo ""

# Run installer
"./$INSTALLER" &

echo ""
echo "Installer started. A browser window should open shortly."
echo "If it doesn't open automatically, visit: http://localhost:18080"
echo ""
read -p "Press Enter to exit..."
