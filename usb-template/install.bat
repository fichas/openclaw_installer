@echo off
chcp 65001 >nul 2>&1

:: Check if running from UNC path
if "%~dp0"=="\\" (
    echo [ERROR] Cannot run from network path.
    echo Please copy OpenClaw folder to local drive (e.g., D:\OpenClaw)
    pause
    exit /b 1
)

cd /d "%~dp0"

:: Check if PowerShell is available
powershell -Command "Get-Host" >nul 2>&1
if %errorlevel% neq 0 (
    echo [ERROR] PowerShell not found. Please install PowerShell.
    pause
    exit /b 1
)

:: Run PowerShell script with proper encoding
powershell -NoProfile -ExecutionPolicy Bypass -Command "[Console]::OutputEncoding = [System.Text.Encoding]::UTF8; & '%~dp0install.ps1'"
