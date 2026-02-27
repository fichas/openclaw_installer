package installer

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"wails-installer/internal/platform"
)

// InstallRequest 安装请求
type InstallRequest struct {
	Mode           string `json:"mode"`
	InstallPath    string `json:"installPath,omitempty"`
	CreateShortcut bool   `json:"createShortcut"`
	AddToPath      bool   `json:"addToPath"`
}

// InstallResult 安装结果
type InstallResult struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// ValidationResult 路径验证结果
type ValidationResult struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// ProgressCallback 进度回调函数
type ProgressCallback func(progress int, status string)

// Installer 安装器
type Installer struct {
	platform   *platform.Platform
	request    InstallRequest
	cancelled  bool
	running    bool
	mu         sync.RWMutex
	ctx        context.Context
	cancelFunc context.CancelFunc
}

// NewInstaller 创建新安装器
func NewInstaller(p *platform.Platform, req InstallRequest) *Installer {
	ctx, cancel := context.WithCancel(context.Background())
	return &Installer{
		platform:   p,
		request:    req,
		ctx:        ctx,
		cancelFunc: cancel,
	}
}

// IsRunning 检查是否正在安装
func (i *Installer) IsRunning() bool {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.running
}

// Cancel 取消安装
func (i *Installer) Cancel() error {
	i.mu.Lock()
	defer i.mu.Unlock()
	if !i.running {
		return nil
	}
	i.cancelled = true
	i.cancelFunc()
	return nil
}

// InstallWithProgress 执行安装并报告进度
func (i *Installer) InstallWithProgress(callback ProgressCallback) InstallResult {
	i.mu.Lock()
	i.running = true
	i.cancelled = false
	i.mu.Unlock()

	defer func() {
		i.mu.Lock()
		i.running = false
		i.mu.Unlock()
	}()

	reportProgress := func(progress int, status string) {
		if callback != nil {
			callback(progress, status)
		}
	}

	// 检查是否已取消
	if i.isCancelled() {
		return InstallResult{Success: false, Error: "安装已取消"}
	}

	// 获取安装目录
	reportProgress(5, "准备安装目录...")
	installDir, err := i.getInstallDir()
	if err != nil {
		return InstallResult{Success: false, Error: fmt.Sprintf("获取安装路径失败: %v", err)}
	}

	// 检查是否已取消
	if i.isCancelled() {
		return InstallResult{Success: false, Error: "安装已取消"}
	}

	// 创建安装目录
	reportProgress(10, "创建安装目录...")
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return InstallResult{Success: false, Error: fmt.Sprintf("创建安装目录失败: %v", err)}
	}

	// 检查是否已取消
	if i.isCancelled() {
		return InstallResult{Success: false, Error: "安装已取消"}
	}

	// 复制二进制文件
	reportProgress(20, "复制程序文件...")
	if err := i.copyBinary(installDir); err != nil {
		return InstallResult{Success: false, Error: fmt.Sprintf("复制程序失败: %v", err)}
	}

	// 检查是否已取消
	if i.isCancelled() {
		return InstallResult{Success: false, Error: "安装已取消"}
	}

	// 创建配置目录
	reportProgress(40, "创建配置目录...")
	configDir := i.platform.GetConfigDir()
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return InstallResult{Success: false, Error: fmt.Sprintf("创建配置目录失败: %v", err)}
	}

	// 检查是否已取消
	if i.isCancelled() {
		return InstallResult{Success: false, Error: "安装已取消"}
	}

	// 添加到PATH
	if i.request.AddToPath {
		reportProgress(60, "配置环境变量...")
		if err := i.addToPath(installDir); err != nil {
			// 非致命错误，继续安装
			reportProgress(65, "配置环境变量失败，继续安装...")
		}
	}

	// 检查是否已取消
	if i.isCancelled() {
		return InstallResult{Success: false, Error: "安装已取消"}
	}

	// 创建快捷方式
	if i.request.CreateShortcut {
		reportProgress(80, "创建快捷方式...")
		if err := i.createShortcut(installDir); err != nil {
			// 非致命错误，继续安装
			reportProgress(85, "创建快捷方式失败，继续安装...")
		}
	}

	// 检查是否已取消
	if i.isCancelled() {
		return InstallResult{Success: false, Error: "安装已取消"}
	}

	// 完成
	reportProgress(100, "安装完成！")
	time.Sleep(500 * time.Millisecond)

	return InstallResult{
		Success: true,
		Message: fmt.Sprintf("OpenClaw 已成功安装到 %s", installDir),
	}
}

// Install 执行安装（兼容旧接口）
func (i *Installer) Install() error {
	result := i.InstallWithProgress(nil)
	if !result.Success {
		return fmt.Errorf(result.Error)
	}
	return nil
}

func (i *Installer) isCancelled() bool {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.cancelled
}

func (i *Installer) getInstallDir() (string, error) {
	opts := i.platform.GetInstallOptions()

	switch i.request.Mode {
	case "system":
		return opts.SystemDir, nil
	case "user":
		return opts.UserDir, nil
	case "portable":
		return filepath.Join(os.TempDir(), "openclaw-portable"), nil
	default:
		if i.request.InstallPath != "" {
			return i.request.InstallPath, nil
		}
		return opts.UserDir, nil
	}
}

func (i *Installer) copyBinary(installDir string) error {
	execPath, err := os.Executable()
	if err != nil {
		return err
	}
	execDir := filepath.Dir(execPath)

	info := i.platform.GetInfo()
	var srcName string
	switch {
	case info.IsWindows:
		srcName = fmt.Sprintf("openclaw-windows-%s.exe", info.Arch)
	case info.IsMacOS:
		srcName = fmt.Sprintf("openclaw-darwin-%s", info.Arch)
	case info.IsLinux:
		srcName = fmt.Sprintf("openclaw-linux-%s", info.Arch)
	}

	srcPath := filepath.Join(execDir, "binaries", srcName)
	// 如果找不到，尝试当前目录
	if _, err := os.Stat(srcPath); os.IsNotExist(err) {
		srcPath = filepath.Join(execDir, srcName)
	}
	// 如果还是找不到，尝试使用自身
	if _, err := os.Stat(srcPath); os.IsNotExist(err) {
		srcPath = execPath
	}

	dstPath := filepath.Join(installDir, i.platform.GetBinaryName())

	return copyFile(srcPath, dstPath)
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// 确保目标目录存在
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	return os.Chmod(dst, 0755)
}

func (i *Installer) addToPath(installDir string) error {
	info := i.platform.GetInfo()

	switch {
	case info.IsWindows:
		return i.addToPathWindows(installDir)
	case info.IsMacOS, info.IsLinux:
		return i.addToPathUnix(installDir)
	}

	return nil
}

func (i *Installer) addToPathWindows(installDir string) error {
	// Windows: 修改用户环境变量
	// 使用 setx 命令添加用户 PATH
	currentPath := os.Getenv("PATH")
	if contains(currentPath, installDir) {
		// 已经在PATH中
		return nil
	}

	// 使用 PowerShell 添加 PATH
	psCmd := fmt.Sprintf(
		`[Environment]::SetEnvironmentVariable("Path", $env:Path + ";%s", "User")`,
		installDir,
	)
	cmd := exec.Command("powershell", "-Command", psCmd)
	return cmd.Run()
}

func (i *Installer) addToPathUnix(installDir string) error {
	// Unix: 修改 shell 配置文件
	home := os.Getenv("HOME")
	if home == "" {
		return fmt.Errorf("无法获取 HOME 目录")
	}

	// 检测使用的 shell
	shell := os.Getenv("SHELL")
	var profileFile string
	if contains(shell, "zsh") {
		profileFile = filepath.Join(home, ".zshrc")
	} else if contains(shell, "bash") {
		profileFile = filepath.Join(home, ".bashrc")
	} else {
		profileFile = filepath.Join(home, ".profile")
	}

	// 检查是否已经在 PATH 中
	content, err := os.ReadFile(profileFile)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	pathLine := fmt.Sprintf("export PATH=\"%s:$PATH\"", installDir)
	if contains(string(content), pathLine) {
		return nil // 已经存在
	}

	// 追加到配置文件
	f, err := os.OpenFile(profileFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(fmt.Sprintf("\n# OpenClaw\n%s\n", pathLine))
	return err
}

func (i *Installer) createShortcut(installDir string) error {
	info := i.platform.GetInfo()

	switch {
	case info.IsWindows:
		return i.createShortcutWindows(installDir)
	case info.IsMacOS:
		return i.createShortcutMacOS(installDir)
	case info.IsLinux:
		return i.createShortcutLinux(installDir)
	}

	return nil
}

func (i *Installer) createShortcutWindows(installDir string) error {
	// 创建开始菜单快捷方式
	startMenuPath := filepath.Join(os.Getenv("APPDATA"), "Microsoft", "Windows", "Start Menu", "Programs")
	shortcutPath := filepath.Join(startMenuPath, "OpenClaw.lnk")

	binaryPath := filepath.Join(installDir, "openclaw.exe")

	// 使用 PowerShell 创建快捷方式
	psScript := fmt.Sprintf(`
$WshShell = New-Object -comObject WScript.Shell
$Shortcut = $WshShell.CreateShortcut("%s")
$Shortcut.TargetPath = "%s"
$Shortcut.WorkingDirectory = "%s"
$Shortcut.Save()
`, shortcutPath, binaryPath, installDir)

	cmd := exec.Command("powershell", "-Command", psScript)
	return cmd.Run()
}

func (i *Installer) createShortcutMacOS(installDir string) error {
	// macOS: 创建 Applications 链接或 alias
	_ = "/Applications"
	_ = filepath.Join("/Applications", "OpenClaw.app")

	// 简化处理：创建一个启动脚本
	binaryPath := filepath.Join(installDir, "openclaw")

	// 创建简单的 AppleScript 应用
	appleScript := fmt.Sprintf(`do shell script "%s &"`, binaryPath)

	cmd := exec.Command("osascript", "-e", appleScript)
	return cmd.Run()
}

func (i *Installer) createShortcutLinux(installDir string) error {
	// Linux: 创建 .desktop 文件
	applicationsDir := filepath.Join(os.Getenv("HOME"), ".local", "share", "applications")
	if err := os.MkdirAll(applicationsDir, 0755); err != nil {
		return err
	}

	desktopFile := filepath.Join(applicationsDir, "openclaw.desktop")

	binaryPath := filepath.Join(installDir, "openclaw")
	content := fmt.Sprintf(`[Desktop Entry]
Name=OpenClaw
Comment=跨平台 AI 助手
Exec=%s
Type=Application
Terminal=false
Categories=Utility;
`, binaryPath)

	return os.WriteFile(desktopFile, []byte(content), 0644)
}

// Launch 启动已安装的应用
func (i *Installer) Launch() error {
	installDir, err := i.getInstallDir()
	if err != nil {
		return err
	}

	binaryPath := filepath.Join(installDir, i.platform.GetBinaryName())

	// 检查文件是否存在
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		return fmt.Errorf("找不到可执行文件: %s", binaryPath)
	}

	// 启动应用
	cmd := exec.Command(binaryPath)
	cmd.Dir = installDir
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	return cmd.Start()
}

// ValidatePath 验证安装路径
func ValidatePath(path string, platform *platform.Platform) ValidationResult {
	if path == "" {
		return ValidationResult{
			Valid:   false,
			Message: "路径不能为空",
		}
	}

	// 检查路径长度
	if len(path) > 200 {
		return ValidationResult{
			Valid:   false,
			Message: "路径过长",
		}
	}

	// 检查非法字符
	info := platform.GetInfo()
	if info.IsWindows {
		invalidChars := []string{"<", ">", ":", "\"", "|", "?", "*"}
		for _, char := range invalidChars {
			if contains(path, char) {
				return ValidationResult{
					Valid:   false,
					Message: fmt.Sprintf("路径包含非法字符: %s", char),
				}
			}
		}
	}

	// 检查目录是否可写
	parentDir := filepath.Dir(path)
	if parentDir != "" && parentDir != "." {
		if err := os.MkdirAll(parentDir, 0755); err != nil {
			return ValidationResult{
				Valid:   false,
				Message: fmt.Sprintf("无法创建目录: %v", err),
			}
		}

		// 尝试创建临时文件测试写入权限
		testFile := filepath.Join(parentDir, ".write_test")
		if f, err := os.Create(testFile); err == nil {
			f.Close()
			os.Remove(testFile)
		} else {
			return ValidationResult{
				Valid:   false,
				Message: "没有写入权限",
			}
		}
	}

	return ValidationResult{
		Valid:   true,
		Message: "路径有效",
	}
}

// BrowseDirectory 打开目录选择对话框
func BrowseDirectory(defaultPath string) string {
	// 由于无法直接调用系统对话框，返回默认路径
	// 实际使用时可以通过前端调用 Wails 的 OpenDialog
	if defaultPath == "" {
		home, _ := os.UserHomeDir()
		return home
	}
	return defaultPath
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
