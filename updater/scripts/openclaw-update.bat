@echo off
REM OpenClaw Update Script for Windows
REM Wrapper script for the OpenClaw updater

setlocal EnableDelayedExpansion

REM Configuration
set "SCRIPT_DIR=%~dp0"
set "INSTALL_DIR=%INSTALL_DIR%"
if "%INSTALL_DIR%"=="" set "INSTALL_DIR=C:\Program Files\OpenClaw"

set "UPDATER_BIN=%UPDATER_BIN%"
if "%UPDATER_BIN%"=="" set "UPDATER_BIN=%SCRIPT_DIR%openclaw-updater.exe"

set "LOG_DIR=%LOG_DIR%"
if "%LOG_DIR%"=="" set "LOG_DIR=C:\ProgramData\OpenClaw\logs"

set "CONFIG_FILE=%CONFIG_FILE%"
if "%CONFIG_FILE%"=="" set "CONFIG_FILE=C:\ProgramData\OpenClaw\updater.json"

set "LOCK_FILE=%TEMP%\openclaw-update.lock"

REM Parse arguments
set COMMAND=update
set UPDATER_ARGS=
set YES_FLAG=0
set FORCE_FLAG=0
set DRY_RUN_FLAG=0
set VERBOSE_FLAG=0
set ADAPTER=

:parse_args
if "%~1"=="" goto :execute

if /i "%~1"=="check" (
    set COMMAND=check
    shift
    goto :parse_args
)
if /i "%~1"=="update" (
    set COMMAND=update
    shift
    goto :parse_args
)
if /i "%~1"=="rollback" (
    set COMMAND=rollback
    shift
    goto :parse_args
)
if /i "%~1"=="list-backups" (
    set COMMAND=list-backups
    shift
    goto :parse_args
)
if /i "%~1"=="install-task" (
    set COMMAND=install-task
    shift
    goto :parse_args
)
if /i "%~1"=="status" (
    set COMMAND=status
    shift
    goto :parse_args
)
if /i "%~1"=="-y" (
    set YES_FLAG=1
    set UPDATER_ARGS=!UPDATER_ARGS! -yes
    shift
    goto :parse_args
)
if /i "%~1"=="--yes" (
    set YES_FLAG=1
    set UPDATER_ARGS=!UPDATER_ARGS! -yes
    shift
    goto :parse_args
)
if /i "%~1"=="-f" (
    set FORCE_FLAG=1
    set UPDATER_ARGS=!UPDATER_ARGS! -force
    shift
    goto :parse_args
)
if /i "%~1"=="--force" (
    set FORCE_FLAG=1
    set UPDATER_ARGS=!UPDATER_ARGS! -force
    shift
    goto :parse_args
)
if /i "%~1"=="-d" (
    set DRY_RUN_FLAG=1
    set UPDATER_ARGS=!UPDATER_ARGS! -dry-run
    shift
    goto :parse_args
)
if /i "%~1"=="--dry-run" (
    set DRY_RUN_FLAG=1
    set UPDATER_ARGS=!UPDATER_ARGS! -dry-run
    shift
    goto :parse_args
)
if /i "%~1"=="-v" (
    set VERBOSE_FLAG=1
    set UPDATER_ARGS=!UPDATER_ARGS! -v
    shift
    goto :parse_args
)
if /i "%~1"=="--verbose" (
    set VERBOSE_FLAG=1
    set UPDATER_ARGS=!UPDATER_ARGS! -v
    shift
    goto :parse_args
)
if /i "%~1"=="-a" (
    set ADAPTER=%~2
    set UPDATER_ARGS=!UPDATER_ARGS! -adapter "%~2"
    shift
    shift
    goto :parse_args
)
if /i "%~1"=="--adapter" (
    set ADAPTER=%~2
    set UPDATER_ARGS=!UPDATER_ARGS! -adapter "%~2"
    shift
    shift
    goto :parse_args
)
if /i "%~1"=="-c" (
    set CONFIG_FILE=%~2
    set UPDATER_ARGS=!UPDATER_ARGS! -config "%~2"
    shift
    shift
    goto :parse_args
)
if /i "%~1"=="--config" (
    set CONFIG_FILE=%~2
    set UPDATER_ARGS=!UPDATER_ARGS! -config "%~2"
    shift
    shift
    goto :parse_args
)
if /i "%~1"=="-h" goto :usage
if /i "%~1"=="--help" goto :usage

echo [ERROR] Unknown option: %~1
goto :usage

:execute
REM Execute command
goto :%COMMAND%

:check
    call :check_updater
    call :ensure_log_dir
    call :acquire_lock
    call :run_updater -check %UPDATER_ARGS%
    call :release_lock
    goto :end

:update
    call :check_updater
    call :ensure_log_dir
    call :acquire_lock
    call :cleanup_logs
    call :run_updater %UPDATER_ARGS%
    call :release_lock
    goto :end

:rollback
    call :check_updater
    call :ensure_log_dir
    call :acquire_lock
    call :run_updater -rollback %UPDATER_ARGS%
    call :release_lock
    goto :end

:list-backups
    call :check_updater
    call :run_updater -list-backups %UPDATER_ARGS%
    goto :end

:install-task
    call :install_scheduled_task
    goto :end

:status
    call :show_status
    goto :end

:end
endlocal
exit /b %ERRORLEVEL%

REM Functions
:check_updater
    if exist "%UPDATER_BIN%" goto :eof
    where openclaw-updater.exe >nul 2>&1
    if %ERRORLEVEL%==0 (
        for /f "tokens=*" %%a in ('where openclaw-updater.exe') do set "UPDATER_BIN=%%a"
        goto :eof
    )
    echo [ERROR] Updater binary not found: %UPDATER_BIN%
    echo [INFO] Please ensure openclaw-updater.exe is installed
    exit /b 1

goto :eof

:ensure_log_dir
    if not exist "%LOG_DIR%" (
        mkdir "%LOG_DIR%" 2>nul
        if %ERRORLEVEL% neq 0 (
            echo [WARN] Cannot create log directory: %LOG_DIR%
            set "LOG_DIR=%TEMP%"
        )
    )
goto :eof

:acquire_lock
    if exist "%LOCK_FILE%" (
        echo [ERROR] Another update process is running
        exit /b 1
    )
    echo %DATE% %TIME% > "%LOCK_FILE%"
goto :eof

:release_lock
    if exist "%LOCK_FILE%" del "%LOCK_FILE%"
goto :eof

:run_updater
    set "TIMESTAMP=%DATE:~-4%%DATE:~4,2%%DATE:~7,2%-%TIME:~0,2%%TIME:~3,2%%TIME:~6,2%"
    set "TIMESTAMP=!TIMESTAMP: =0!"
    set "LOG_FILE=%LOG_DIR%\update-!TIMESTAMP!.log"

    echo [INFO] Running updater with args: %*
    echo [INFO] Log file: %LOG_FILE%

    "%UPDATER_BIN%" %* > "%LOG_FILE%" 2>&1
    set EXIT_CODE=%ERRORLEVEL%

    type "%LOG_FILE%"

    if %EXIT_CODE%==0 (
        echo [OK] Updater completed successfully
    ) else (
        echo [ERROR] Updater failed with exit code: %EXIT_CODE%
        echo [INFO] Check log file for details: %LOG_FILE%
    )
    exit /b %EXIT_CODE%
goto :eof

:cleanup_logs
    echo [INFO] Cleaning up old log files...
    forfiles /P "%LOG_DIR%" /S /M "update-*.log" /D -30 /C "cmd /c del @path" 2>nul
goto :eof

:install_scheduled_task
    echo [INFO] Installing scheduled task for automatic updates...

    set "TASK_NAME=OpenClawUpdate"
    set "SCRIPT_PATH=%~f0"

    schtasks /query /tn "%TASK_NAME%" >nul 2>&1
    if %ERRORLEVEL%==0 (
        echo [INFO] Task already exists, removing old task...
        schtasks /delete /tn "%TASK_NAME%" /f >nul 2>&1
    )

    schtasks /create /tn "%TASK_NAME%" /tr "\"%SCRIPT_PATH%\" update --yes" /sc daily /st 02:00 /ru SYSTEM /rl HIGHEST /f

    if %ERRORLEVEL%==0 (
        echo [OK] Scheduled task installed successfully
        echo [INFO] Updates will run daily at 2:00 AM
        schtasks /query /tn "%TASK_NAME%"
    ) else (
        echo [ERROR] Failed to install scheduled task
        echo [INFO] Try running as Administrator
    )
goto :eof

:show_status
    echo [INFO] OpenClaw Update Status
    echo ========================

    if exist "%UPDATER_BIN%" (
        echo [OK] Updater binary: %UPDATER_BIN%
    ) else (
        echo [ERROR] Updater binary not found: %UPDATER_BIN%
    )

    if exist "%INSTALL_DIR%" (
        echo [OK] Install directory: %INSTALL_DIR%
    ) else (
        echo [WARN] Install directory not found: %INSTALL_DIR%
    )

    if exist "%LOG_DIR%" (
        echo [OK] Log directory: %LOG_DIR%
    ) else (
        echo [WARN] Log directory not found: %LOG_DIR%
    )

    if exist "%CONFIG_FILE%" (
        echo [OK] Configuration file: %CONFIG_FILE%
    ) else (
        echo [WARN] Configuration file not found: %CONFIG_FILE%
    )

    schtasks /query /tn "OpenClawUpdate" >nul 2>&1
    if %ERRORLEVEL%==0 (
        echo [INFO] Scheduled task: installed
        schtasks /query /tn "OpenClawUpdate" /fo list /v | findstr "Task Name"
        schtasks /query /tn "OpenClawUpdate" /fo list /v | findstr "Next Run Time"
    ) else (
        echo [INFO] Scheduled task: not installed
    )

    echo.
    echo [INFO] Recent update logs:
    dir /b /o-d "%LOG_DIR%\update-*.log" 2>nul | head -5

goto :eof

:usage
echo OpenClaw Update Script for Windows
echo.
echo Usage: %~nx0 [OPTIONS] [COMMAND]
echo.
echo Commands:
echo     check           Check for available updates only
echo     update          Perform update ^(default^)
echo     rollback        Rollback to previous version
echo     list-backups    List available backups
echo     install-task    Install scheduled task for automatic updates
echo     status          Show update status
echo.
echo Options:
echo     -y, --yes               Auto-confirm updates
echo     -f, --force             Force update even if versions match
echo     -d, --dry-run           Simulate update without making changes
echo     -a, --adapter NAME      Update specific adapter
echo     -c, --config PATH       Path to configuration file
echo     -v, --verbose           Verbose output
echo     -h, --help              Show this help message
echo.
echo Examples:
echo     %~nx0 check                    Check for updates
echo     %~nx0 update --yes             Update with auto-confirm
echo     %~nx0 update --adapter=wecom   Update only wecom adapter
echo     %~nx0 rollback                 Rollback to previous version
echo     %~nx0 install-task             Setup automatic updates
goto :end
