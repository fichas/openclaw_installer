@echo off
setlocal enabledelayedexpansion

:: OpenClaw Windows Installer
:: Version: 1.0.0

echo ========================================
echo OpenClaw Installer for Windows
echo Version: 1.0.0
echo ========================================
echo.

:: Auto-detect architecture
if "%PROCESSOR_ARCHITECTURE%"=="AMD64" (
    set ARCH=amd64
    set INSTALLER=openclaw-installer-windows-amd64.exe
) else if "%PROCESSOR_ARCHITECTURE%"=="ARM64" (
    set ARCH=arm64
    set INSTALLER=openclaw-installer-windows-arm64.exe
) else (
    echo Unsupported architecture: %PROCESSOR_ARCHITECTURE%
    echo Supported architectures: AMD64, ARM64
    pause
    exit /b 1
)

echo Detected architecture: %ARCH%
echo.

:: Check if installer exists
if not exist "%INSTALLER%" (
    echo Error: Installer not found: %INSTALLER%
    echo Available files:
    dir /b *.exe
    pause
    exit /b 1
)

echo Starting OpenClaw Installer...
echo.

:: Run installer
start "" "%INSTALLER%"

echo.
echo Installer started. A browser window should open shortly.
echo If it doesn't open automatically, visit: http://localhost:18080
echo.
pause
