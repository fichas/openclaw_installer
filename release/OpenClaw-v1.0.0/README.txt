OpenClaw v1.0.0 Release Package
================================

Thank you for downloading OpenClaw!

QUICK START
-----------

Windows:
  1. Double-click install.bat (or run install.ps1 in PowerShell)
  2. The installer will auto-detect your architecture (x64/ARM64)
  3. Follow the web-based installation wizard

macOS:
  1. Double-click install-mac.command
  2. The installer will auto-detect your Mac type (Intel/Apple Silicon)
  3. Follow the web-based installation wizard

Linux:
  1. Run: ./install-linux.sh
  2. The installer will auto-detect your architecture (x64/ARM64)
  3. Follow the web-based installation wizard

PACKAGE CONTENTS
----------------

windows/
  - openclaw-installer-windows-amd64.exe  (for x64 systems)
  - openclaw-installer-windows-arm64.exe  (for ARM64 systems)
  - install.bat                           (Command Prompt installer)
  - install.ps1                           (PowerShell installer)

macos/
  - OpenClaw-Installer-darwin-amd64       (for Intel Macs)
  - OpenClaw-Installer-darwin-arm64       (for Apple Silicon Macs)
  - install-mac.command                   (Auto-detect installer)

linux/
  - openclaw-installer-linux-amd64        (for x64 systems)
  - openclaw-installer-linux-arm64        (for ARM64 systems)
  - install-linux.sh                      (Auto-detect installer)

shared/
  - adapters/                             (Platform adapters)
    - dingtalk/                           (DingTalk adapter)
    - feishu/                             (Feishu adapter)
    - wechat-work/                        (WeChat Work adapter)
  - config-templates/                     (Configuration templates)
    - dingtalk-adapter.yaml.template
    - feishu-adapter.yaml.template
    - openclaw.yaml.template
    - wecom-adapter.yaml.template

SYSTEM REQUIREMENTS
-------------------

Windows:
  - Windows 10 or Windows 11
  - x64 (AMD64) or ARM64 processor
  - 100 MB free disk space

macOS:
  - macOS 10.14 (Mojave) or later
  - Intel or Apple Silicon (M1/M2/M3) processor
  - 100 MB free disk space

Linux:
  - Most modern Linux distributions
  - x64 (AMD64) or ARM64 processor
  - 100 MB free disk space

INSTALLATION NOTES
------------------

1. The installer runs a local web server on port 18080
2. Your browser will automatically open to http://localhost:18080
3. If the browser doesn't open, manually navigate to the URL
4. Follow the on-screen instructions to complete installation

TROUBLESHOOTING
---------------

- Port 18080 in use: Make sure no other application is using port 18080
- Firewall blocking: Allow the installer through your firewall
- Permission denied: Make sure the installer has execute permissions (macOS/Linux)

SUPPORT
-------

For more information and support, visit:
https://github.com/openclaw/openclaw

Version: 1.0.0
Build Date: 2026-02-28
