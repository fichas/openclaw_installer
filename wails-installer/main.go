package main

import (
	"context"
	"embed"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	runtime2 "github.com/wailsapp/wails/v2/pkg/runtime"

	"wails-installer/internal/config"
	"wails-installer/internal/installer"
	"wails-installer/internal/platform"
)

//go:embed frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var icon []byte

// Logger 全局日志记录器
var logger *log.Logger

// initLogger 初始化日志记录器
// Windows 无窗口模式下需要将日志写入文件，因为程序没有控制台输出
func initLogger() (*os.File, error) {
	// 获取日志目录
	var logDir string
	switch runtime.GOOS {
	case "windows":
		logDir = filepath.Join(os.Getenv("APPDATA"), "OpenClaw", "logs")
	case "darwin":
		logDir = filepath.Join(os.Getenv("HOME"), "Library", "Logs", "OpenClaw")
	default: // linux
		logDir = filepath.Join(os.Getenv("HOME"), ".local", "share", "OpenClaw", "logs")
	}

	// 创建日志目录
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("创建日志目录失败: %w", err)
	}

	// 生成日志文件名（带时间戳）
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	logFile := filepath.Join(logDir, fmt.Sprintf("installer_%s.log", timestamp))

	// 打开日志文件
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("打开日志文件失败: %w", err)
	}

	// 创建多输出日志记录器（同时输出到文件和标准错误）
	// 注意：在 Windows -H=windowsgui 模式下，标准输出/错误会被丢弃
	multiWriter := io.MultiWriter(file, os.Stderr)
	logger = log.New(multiWriter, "[OpenClaw] ", log.LstdFlags|log.Lshortfile)

	logger.Printf("日志系统初始化完成")
	logger.Printf("日志文件: %s", logFile)
	logger.Printf("操作系统: %s, 架构: %s", runtime.GOOS, runtime.GOARCH)

	return file, nil
}

// App 结构体
type App struct {
	ctx       context.Context
	platform  *platform.Platform
	config    *config.Config
	installer *installer.Installer
}

// NewApp 创建新App
func NewApp() *App {
	return &App{
		platform: platform.NewPlatform(),
		config:   config.NewConfig(""),
	}
}

// startup 应用启动时调用
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	// 初始化配置目录
	a.config = config.NewConfig(a.platform.GetConfigDir())
}

// shutdown 应用关闭时调用
func (a *App) shutdown(ctx context.Context) {
	// 清理资源
}

// domReady DOM加载完成时调用
func (a *App) domReady(ctx context.Context) {
	// DOM已准备好，可以执行前端相关操作
}

// beforeClose 应用关闭前调用，返回true则阻止关闭
func (a *App) beforeClose(ctx context.Context) bool {
	// 如果正在安装，提示用户
	if a.installer != nil && a.installer.IsRunning() {
		return true // 阻止关闭
	}
	return false
}

// GetPlatformInfo 获取平台信息
func (a *App) GetPlatformInfo() platform.PlatformInfo {
	return a.platform.GetInfo()
}

// GetInstallOptions 获取安装选项
func (a *App) GetInstallOptions() platform.InstallOptions {
	return a.platform.GetInstallOptions()
}

// Install 执行安装
func (a *App) Install(req installer.InstallRequest) installer.InstallResult {
	a.installer = installer.NewInstaller(a.platform, req)
	result := a.installer.InstallWithProgress(func(progress int, status string) {
		// 通过事件发送进度到前端
		if a.ctx != nil {
			runtime2.EventsEmit(a.ctx, "install:progress", map[string]interface{}{
				"progress": progress,
				"status":   status,
			})
		}
	})
	return result
}

// CancelInstall 取消安装
func (a *App) CancelInstall() error {
	if a.installer != nil {
		return a.installer.Cancel()
	}
	return nil
}

// SaveConfig 保存配置
func (a *App) SaveConfig(cfg config.OpenClawConfig) error {
	return a.config.Save(cfg)
}

// GetDefaultConfig 获取默认配置
func (a *App) GetDefaultConfig() config.OpenClawConfig {
	return a.config.GetDefault()
}

// ValidateInstallPath 验证安装路径
func (a *App) ValidateInstallPath(path string) installer.ValidationResult {
	return installer.ValidatePath(path, a.platform)
}

// BrowseDirectory 打开目录选择对话框
func (a *App) BrowseDirectory(defaultPath string) string {
	return installer.BrowseDirectory(defaultPath)
}

// LaunchApp 启动已安装的应用
func (a *App) LaunchApp() error {
	return a.installer.Launch()
}

// GetVersion 获取安装器版本
func (a *App) GetVersion() string {
	return "1.0.0"
}

func main() {
	// 初始化日志系统
	// Windows 无窗口模式下必须使用文件日志，因为程序没有控制台
	logFile, err := initLogger()
	if err != nil {
		// 日志初始化失败，回退到标准错误输出
		// 在 Windows -H=windowsgui 模式下，用户看不到这个输出
		// 但在调试模式下（不带 -H=windowsgui）可以看到
		fmt.Fprintf(os.Stderr, "初始化日志失败: %v\n", err)
	}
	if logFile != nil {
		defer logFile.Close()
	}

	if logger != nil {
		logger.Println("OpenClaw 安装器启动")
	}

	// 创建应用实例
	app := NewApp()

	// 配置Windows选项
	windowsOptions := &windows.Options{
		WebviewIsTransparent:              true,
		WindowIsTranslucent:               true,
		DisableWindowIcon:                 false,
		DisableFramelessWindowDecorations: false,
		WebviewUserDataPath:               "",
		WebviewBrowserPath:                "",
		Theme:                             windows.Dark,
	}

	// 配置macOS选项
	macOptions := &mac.Options{
		TitleBar: &mac.TitleBar{
			TitlebarAppearsTransparent: true,
			HideTitle:                  false,
			HideTitleBar:               false,
			FullSizeContent:          true,
			UseToolbar:               false,
			HideToolbarSeparator:     true,
		},
		Appearance:           mac.NSAppearanceNameDarkAqua,
		WebviewIsTransparent: true,
		WindowIsTranslucent:  true,
		About: &mac.AboutInfo{
			Title:   "OpenClaw Installer",
			Message: "Copyright © 2026 OpenClaw Team\n跨平台 AI 助手安装器",
			Icon:    icon,
		},
	}

	// 配置Linux选项
	linuxOptions := &linux.Options{
		Icon:                icon,
		WindowIsTranslucent: true,
	}

	// 创建Wails应用
	var runErr error
	runErr = wails.Run(&options.App{
		Title:             "OpenClaw 安装器",
		Width:             900,
		Height:            700,
		MinWidth:          800,
		MinHeight:         600,
		MaxWidth:          1200,
		MaxHeight:         900,
		DisableResize:     false,
		Fullscreen:        false,
		Frameless:         false,
		StartHidden:       false,
		HideWindowOnClose: false,
		BackgroundColour:  &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		AssetServer: &assetserver.Options{
			Assets:  assets,
			Handler: nil,
		},
		Menu:             nil,
		Logger:           nil,
		LogLevel:         0,
		OnStartup:        app.startup,
		OnDomReady:       app.domReady,
		OnShutdown:       app.shutdown,
		OnBeforeClose:    app.beforeClose,
		WindowStartState: options.Normal,
		Bind: []interface{}{
			app,
		},
		Windows: windowsOptions,
		Mac:     macOptions,
		Linux:   linuxOptions,
	})

	if runErr != nil {
		if logger != nil {
			logger.Printf("应用运行错误: %v", runErr)
		}
		// 在 Windows 无窗口模式下，fmt.Println 不会有任何可见输出
		// 错误信息已记录到日志文件
		fmt.Fprintf(os.Stderr, "Error: %v\n", runErr)
		os.Exit(1)
	}

	if logger != nil {
		logger.Println("OpenClaw 安装器正常退出")
	}
}
