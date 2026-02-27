# OpenClaw Windows Installer
# PowerShell Script with proper UTF-8 support

[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$OutputEncoding = [System.Text.Encoding]::UTF8

# Check if running from UNC path
if ($PSScriptRoot -match "^\\\\") {
    Write-Host "错误: 不能从网络路径运行此脚本" -ForegroundColor Red
    Write-Host "请将 OpenClaw 文件夹复制到本地磁盘（如 D:\OpenClaw），然后再运行" -ForegroundColor Yellow
    Read-Host "按回车键退出"
    exit 1
}

# Get script directory
$ScriptDir = $PSScriptRoot
if (-not $ScriptDir) {
    $ScriptDir = (Get-Location).Path
}

Clear-Host
Write-Host "============================================" -ForegroundColor Cyan
Write-Host "   OpenClaw 跨平台 AI 助手安装器" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""

# Detect architecture
$Arch = "amd64"
if ($env:PROCESSOR_ARCHITECTURE -eq "AMD64") {
    $Arch = "amd64"
    Write-Host "[INFO] 检测到系统: Windows 64位 (x64)" -ForegroundColor Green
} elseif ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") {
    $Arch = "arm64"
    Write-Host "[INFO] 检测到系统: Windows ARM64" -ForegroundColor Green
} else {
    $Arch = "amd64"
    Write-Host "[WARNING] 无法检测架构，默认使用 x64 版本" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "============================================" -ForegroundColor Cyan
Write-Host "   请选择安装方式" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""
Write-Host " [1] 安装到系统目录 (推荐)" -ForegroundColor White
Write-Host "     路径: C:\Program Files\OpenClaw\" -ForegroundColor Gray
Write-Host "     需要管理员权限，所有用户可用" -ForegroundColor Gray
Write-Host ""
Write-Host " [2] 安装到用户目录" -ForegroundColor White
Write-Host "     路径: %LOCALAPPDATA%\OpenClaw\" -ForegroundColor Gray
Write-Host "     无需管理员权限，仅当前用户可用" -ForegroundColor Gray
Write-Host ""
Write-Host " [3] 仅运行，不安装 (便携模式)" -ForegroundColor White
Write-Host "     直接从U盘运行，不复制到系统" -ForegroundColor Gray
Write-Host ""

$choice = Read-Host "请输入选项 (1/2/3)"

switch ($choice) {
    "1" {
        Write-Host ""
        Write-Host "[INFO] 准备安装到系统目录..." -ForegroundColor Green

        # Check admin rights
        $currentPrincipal = New-Object Security.Principal.WindowsPrincipal([Security.Principal.WindowsIdentity]::GetCurrent())
        if (-not $currentPrincipal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) {
            Write-Host "[INFO] 需要管理员权限，正在请求提升..." -ForegroundColor Yellow
            Start-Process powershell.exe -ArgumentList "-NoProfile -ExecutionPolicy Bypass -File `"$PSCommandPath`"" -Verb RunAs
            exit
        }

        $InstallDir = "C:\Program Files\OpenClaw"
        $ConfigDir = "$env:PROGRAMDATA\OpenClaw"

        Write-Host "[INFO] 安装目录: $InstallDir" -ForegroundColor Gray
        Write-Host "[INFO] 配置目录: $ConfigDir" -ForegroundColor Gray
        Write-Host ""

        # Create directories
        New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
        New-Item -ItemType Directory -Force -Path $ConfigDir | Out-Null

        # Copy files
        Write-Host "[INFO] 复制安装文件..." -ForegroundColor Green
        Copy-Item -Path "$ScriptDir\installers\openclaw-installer-windows-$Arch.exe" -Destination "$InstallDir\openclaw.exe" -Force

        # Copy adapter configs
        Write-Host "[INFO] 复制适配器配置..." -ForegroundColor Green
        if (Test-Path "$ScriptDir\packages\config-templates") {
            Copy-Item -Path "$ScriptDir\packages\config-templates\*" -Destination $ConfigDir -Recurse -Force
        }

        # Add to PATH
        Write-Host "[INFO] 添加到系统环境变量..." -ForegroundColor Green
        $currentPath = [Environment]::GetEnvironmentVariable("PATH", "Machine")
        if ($currentPath -notlike "*$InstallDir*") {
            [Environment]::SetEnvironmentVariable("PATH", "$currentPath;$InstallDir", "Machine")
        }

        # Create Start Menu shortcut
        Write-Host "[INFO] 创建开始菜单快捷方式..." -ForegroundColor Green
        $WshShell = New-Object -comObject WScript.Shell
        $Shortcut = $WshShell.CreateShortcut("$env:APPDATA\Microsoft\Windows\Start Menu\Programs\OpenClaw.lnk")
        $Shortcut.TargetPath = "$InstallDir\openclaw.exe"
        $Shortcut.Save()

        Write-Host ""
        Write-Host "============================================" -ForegroundColor Green
        Write-Host "   安装完成！" -ForegroundColor Green
        Write-Host "============================================" -ForegroundColor Green
        Write-Host ""
        Write-Host "[INFO] 正在启动配置向导..." -ForegroundColor Green
        Write-Host "[INFO] 浏览器将自动打开 http://localhost:18080" -ForegroundColor Green
        Write-Host ""

        # Launch installer
        Start-Process -FilePath "$InstallDir\openclaw.exe" -WorkingDirectory $InstallDir
    }

    "2" {
        Write-Host ""
        Write-Host "[INFO] 安装到用户目录..." -ForegroundColor Green

        $InstallDir = "$env:LOCALAPPDATA\OpenClaw"
        $ConfigDir = "$env:APPDATA\OpenClaw"

        Write-Host "[INFO] 安装目录: $InstallDir" -ForegroundColor Gray
        Write-Host "[INFO] 配置目录: $ConfigDir" -ForegroundColor Gray
        Write-Host ""

        # Create directories
        New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
        New-Item -ItemType Directory -Force -Path $ConfigDir | Out-Null

        # Copy files
        Write-Host "[INFO] 复制安装文件..." -ForegroundColor Green
        Copy-Item -Path "$ScriptDir\installers\openclaw-installer-windows-$Arch.exe" -Destination "$InstallDir\openclaw.exe" -Force

        # Copy adapter configs
        Write-Host "[INFO] 复制适配器配置..." -ForegroundColor Green
        if (Test-Path "$ScriptDir\packages\config-templates") {
            Copy-Item -Path "$ScriptDir\packages\config-templates\*" -Destination $ConfigDir -Recurse -Force
        }

        # Add to PATH
        Write-Host "[INFO] 添加到用户环境变量..." -ForegroundColor Green
        $currentPath = [Environment]::GetEnvironmentVariable("PATH", "User")
        if ($currentPath -notlike "*$InstallDir*") {
            [Environment]::SetEnvironmentVariable("PATH", "$currentPath;$InstallDir", "User")
        }

        # Create Start Menu shortcut
        Write-Host "[INFO] 创建开始菜单快捷方式..." -ForegroundColor Green
        $WshShell = New-Object -comObject WScript.Shell
        $Shortcut = $WshShell.CreateShortcut("$env:APPDATA\Microsoft\Windows\Start Menu\Programs\OpenClaw.lnk")
        $Shortcut.TargetPath = "$InstallDir\openclaw.exe"
        $Shortcut.Save()

        Write-Host ""
        Write-Host "============================================" -ForegroundColor Green
        Write-Host "   安装完成！" -ForegroundColor Green
        Write-Host "============================================" -ForegroundColor Green
        Write-Host ""
        Write-Host "[INFO] 正在启动配置向导..." -ForegroundColor Green
        Write-Host "[INFO] 浏览器将自动打开 http://localhost:18080" -ForegroundColor Green
        Write-Host ""

        # Launch installer
        Start-Process -FilePath "$InstallDir\openclaw.exe" -WorkingDirectory $InstallDir
    }

    default {
        # Portable mode
        Write-Host ""
        Write-Host "[INFO] 便携模式 - 直接从U盘运行" -ForegroundColor Yellow
        Write-Host "[INFO] 启动安装器..." -ForegroundColor Green
        Write-Host ""

        # Launch installer from USB
        Start-Process -FilePath "$ScriptDir\installers\openclaw-installer-windows-$Arch.exe" -ArgumentList "--portable"
    }
}

Write-Host ""
Write-Host "安装程序已在后台运行。" -ForegroundColor Green
Write-Host "如果浏览器没有自动打开，请手动访问: http://localhost:18080" -ForegroundColor Yellow
Write-Host ""
Read-Host "按回车键关闭此窗口"
