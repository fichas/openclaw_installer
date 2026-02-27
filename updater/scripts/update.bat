@echo off
REM OpenClaw Updater - Windows 更新脚本包装器
REM 此脚本用于包装 Go 更新程序，提供额外的检查和便利功能

setlocal enabledelayedexpansion

REM 颜色定义（仅支持 Windows 10+)
set "RED="
set "GREEN="
set "YELLOW="
set "NC="
for /F "tokens=1,2 delims=#" %%a in ('"prompt #$H#$E# & echo on & for %%b in (1) do rem"') do (
    set "ESC=%%b"
)

REM 检测架构
if "%PROCESSOR_ARCHITECTURE%"=="AMD64" (
    set "ARCH=amd64"
) else if "%PROCESSOR_ARCHITEW6432%"=="AMD64" (
    set "ARCH=amd64"
) else if "%PROCESSOR_ARCHITECTURE%"=="ARM64" (
    set "ARCH=arm64"
) else (
    echo Error: Unsupported architecture: %PROCESSOR_ARCHITECTURE%
    exit /b 1
)

REM 更新程序二进制文件名
set "UPDATER_BIN=openclaw-updater-windows-%ARCH%.exe"

REM 脚本目录
set "SCRIPT_DIR=%~dp0"

REM 检查更新程序是否存在
if not exist "%SCRIPT_DIR%\%UPDATER_BIN%" (
    REM 尝试使用通用名称
    if exist "%SCRIPT_DIR%\openclaw-updater.exe" (
        set "UPDATER_BIN=openclaw-updater.exe"
    ) else (
        echo Error: Updater binary not found: %UPDATER_BIN%
        exit /b 1
    )
)

REM 检查管理员权限
net session > nul 2>&1
if %errorlevel% neq 0 (
    echo Warning: This script may require administrator privileges to update system files.
    echo Consider running as Administrator.
    echo.
)

REM 执行更新程序
echo OpenClaw Updater
echo Platform: windows/%ARCH%
echo.

"%SCRIPT_DIR%\%UPDATER_BIN%" %*

endlocal
