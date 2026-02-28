; NSIS Installer Script for OpenClaw
; This script creates a Windows installer for OpenClaw

!include "MUI2.nsh"
!include "LogicLib.nsh"

; General
Name "OpenClaw Installer"
OutFile "OpenClaw-Installer-Windows.exe"
InstallDir "$PROGRAMFILES64\OpenClaw"
InstallDirRegKey HKLM "Software\OpenClaw" "InstallDir"
RequestExecutionLevel admin

; Interface Settings
!define MUI_ABORTWARNING
!define MUI_ICON "icon.ico"
!define MUI_UNICON "icon.ico"
!define MUI_WELCOMEFINISHPAGE_BITMAP "sidebar.bmp"
!define MUI_HEADERIMAGE
!define MUI_HEADERIMAGE_BITMAP "header.bmp"

; Pages
!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_LICENSE "license.txt"
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH

!insertmacro MUI_UNPAGE_WELCOME
!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES
!insertmacro MUI_UNPAGE_FINISH

; Languages
!insertmacro MUI_LANGUAGE "SimpChinese"
!insertmacro MUI_LANGUAGE "English"

; Installer Sections
Section "OpenClaw" SecMain
    SetOutPath "$INSTDIR"

    ; Main executable
    File "OpenClaw-Installer.exe"
    File "openclaw.exe"

    ; Adapters
    SetOutPath "$INSTDIR\adapters"
    File "wecom-adapter.exe"
    File "dingtalk-adapter.exe"
    File "feishu-adapter.exe"

    ; Create shortcuts
    CreateDirectory "$SMPROGRAMS\OpenClaw"
    CreateShortcut "$SMPROGRAMS\OpenClaw\OpenClaw.lnk" "$INSTDIR\openclaw.exe"
    CreateShortcut "$SMPROGRAMS\OpenClaw\Uninstall.lnk" "$INSTDIR\uninstall.exe"
    CreateShortcut "$DESKTOP\OpenClaw.lnk" "$INSTDIR\openclaw.exe"

    ; Registry entries
    WriteRegStr HKLM "Software\OpenClaw" "InstallDir" "$INSTDIR"
    WriteRegStr HKLM "Software\OpenClaw" "Version" "1.0.0"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\OpenClaw" "DisplayName" "OpenClaw"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\OpenClaw" "UninstallString" "$INSTDIR\uninstall.exe"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\OpenClaw" "DisplayIcon" "$INSTDIR\openclaw.exe"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\OpenClaw" "Publisher" "OpenClaw Team"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\OpenClaw" "Version" "1.0.0"

    ; Create uninstaller
    WriteUninstaller "$INSTDIR\uninstall.exe"
SectionEnd

; Uninstaller Section
Section "Uninstall"
    ; Remove files
    Delete "$INSTDIR\openclaw.exe"
    Delete "$INSTDIR\OpenClaw-Installer.exe"
    Delete "$INSTDIR\adapters\wecom-adapter.exe"
    Delete "$INSTDIR\adapters\dingtalk-adapter.exe"
    Delete "$INSTDIR\adapters\feishu-adapter.exe"
    RMDir "$INSTDIR\adapters"
    Delete "$INSTDIR\uninstall.exe"
    RMDir "$INSTDIR"

    ; Remove shortcuts
    Delete "$SMPROGRAMS\OpenClaw\OpenClaw.lnk"
    Delete "$SMPROGRAMS\OpenClaw\Uninstall.lnk"
    RMDir "$SMPROGRAMS\OpenClaw"
    Delete "$DESKTOP\OpenClaw.lnk"

    ; Remove registry entries
    DeleteRegKey HKLM "Software\OpenClaw"
    DeleteRegKey HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\OpenClaw"
SectionEnd
