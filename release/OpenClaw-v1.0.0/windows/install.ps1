# OpenClaw Windows Installer PowerShell Script
# Version: 1.0.0

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "OpenClaw Installer for Windows" -ForegroundColor Cyan
Write-Host "Version: 1.0.0" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Auto-detect architecture
$arch = $env:PROCESSOR_ARCHITECTURE
$installer = ""

switch ($arch) {
    "AMD64" {
        $installer = "openclaw-installer-windows-amd64.exe"
        Write-Host "Detected architecture: x64 (AMD64)" -ForegroundColor Green
    }
    "ARM64" {
        $installer = "openclaw-installer-windows-arm64.exe"
        Write-Host "Detected architecture: ARM64" -ForegroundColor Green
    }
    default {
        Write-Host "Unsupported architecture: $arch" -ForegroundColor Red
        Write-Host "Supported architectures: AMD64, ARM64" -ForegroundColor Yellow
        Read-Host "Press Enter to exit"
        exit 1
    }
}

Write-Host ""

# Check if installer exists
if (-not (Test-Path $installer)) {
    Write-Host "Error: Installer not found: $installer" -ForegroundColor Red
    Write-Host "Available files:" -ForegroundColor Yellow
    Get-ChildItem -Filter "*.exe" | ForEach-Object { Write-Host "  $($_.Name)" }
    Read-Host "Press Enter to exit"
    exit 1
}

Write-Host "Starting OpenClaw Installer..." -ForegroundColor Blue
Write-Host ""

# Run installer
Start-Process -FilePath $installer

Write-Host ""
Write-Host "Installer started. A browser window should open shortly." -ForegroundColor Green
Write-Host "If it doesn't open automatically, visit: http://localhost:18080" -ForegroundColor Yellow
Write-Host ""
Read-Host "Press Enter to exit"
