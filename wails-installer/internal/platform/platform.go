package platform

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Platform 平台信息
type Platform struct {
	OS   string
	Arch string
}

// PlatformInfo 平台信息
type PlatformInfo struct {
	OS         string `json:"os"`
	Arch       string `json:"arch"`
	IsWindows  bool   `json:"isWindows"`
	IsMacOS    bool   `json:"isMacOS"`
	IsLinux    bool   `json:"isLinux"`
	Version    string `json:"version,omitempty"`
	IsAdmin    bool   `json:"isAdmin"`
}

// InstallOptions 安装选项
type InstallOptions struct {
	SystemDir string `json:"systemDir"`
	UserDir   string `json:"userDir"`
	HasAdmin  bool   `json:"hasAdmin"`
	RecommendedMode string `json:"recommendedMode"`
}

// SystemInfo 系统详细信息
type SystemInfo struct {
	OSVersion     string `json:"osVersion"`
	TotalMemory   uint64 `json:"totalMemory"`
	FreeDiskSpace uint64 `json:"freeDiskSpace"`
	CPUCount      int    `json:"cpuCount"`
}

// NewPlatform 创建新平台检测器
func NewPlatform() *Platform {
	return &Platform{
		OS:   runtime.GOOS,
		Arch: normalizeArch(runtime.GOARCH),
	}
}

// GetInfo 获取平台信息
func (p *Platform) GetInfo() PlatformInfo {
	osVersion := p.getOSVersion()
	return PlatformInfo{
		OS:        p.OS,
		Arch:      p.Arch,
		IsWindows: p.OS == "windows",
		IsMacOS:   p.OS == "darwin",
		IsLinux:   p.OS == "linux",
		Version:   osVersion,
		IsAdmin:   p.hasAdmin(),
	}
}

// GetInstallOptions 获取安装选项
func (p *Platform) GetInstallOptions() InstallOptions {
	opts := InstallOptions{
		HasAdmin: p.hasAdmin(),
	}

	switch p.OS {
	case "windows":
		opts.SystemDir = `C:\Program Files\OpenClaw`
		opts.UserDir = filepath.Join(os.Getenv("LOCALAPPDATA"), "OpenClaw")
		if opts.HasAdmin {
			opts.RecommendedMode = "system"
		} else {
			opts.RecommendedMode = "user"
		}
	case "darwin":
		opts.SystemDir = "/usr/local/bin"
		opts.UserDir = filepath.Join(os.Getenv("HOME"), ".local", "bin")
		opts.RecommendedMode = "user"
	case "linux":
		opts.SystemDir = "/usr/local/bin"
		opts.UserDir = filepath.Join(os.Getenv("HOME"), ".local", "bin")
		opts.RecommendedMode = "user"
	}

	return opts
}

// GetBinaryName 获取二进制文件名
func (p *Platform) GetBinaryName() string {
	if p.OS == "windows" {
		return "openclaw.exe"
	}
	return "openclaw"
}

// GetConfigDir 获取配置目录
func (p *Platform) GetConfigDir() string {
	switch p.OS {
	case "windows":
		return filepath.Join(os.Getenv("APPDATA"), "OpenClaw")
	case "darwin":
		return filepath.Join(os.Getenv("HOME"), ".openclaw")
	case "linux":
		return filepath.Join(os.Getenv("HOME"), ".openclaw")
	}
	return ".openclaw"
}

// GetLogDir 获取日志目录
func (p *Platform) GetLogDir() string {
	switch p.OS {
	case "windows":
		return filepath.Join(os.Getenv("APPDATA"), "OpenClaw", "logs")
	case "darwin":
		return filepath.Join(os.Getenv("HOME"), "Library", "Logs", "OpenClaw")
	case "linux":
		return filepath.Join(os.Getenv("HOME"), ".local", "share", "OpenClaw", "logs")
	}
	return "logs"
}

// GetDataDir 获取数据目录
func (p *Platform) GetDataDir() string {
	switch p.OS {
	case "windows":
		return filepath.Join(os.Getenv("LOCALAPPDATA"), "OpenClaw", "data")
	case "darwin":
		return filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "OpenClaw")
	case "linux":
		return filepath.Join(os.Getenv("HOME"), ".local", "share", "OpenClaw")
	}
	return "data"
}

// hasAdmin 检查是否有管理员权限
func (p *Platform) hasAdmin() bool {
	switch p.OS {
	case "windows":
		// Windows: 尝试打开一个需要管理员权限的资源
		_, err := os.Open(`\\.\PHYSICALDRIVE0`)
		if err == nil {
			return true
		}
		// 备用方法：检查是否能写入系统目录
		testFile := filepath.Join(os.Getenv("ProgramFiles"), ".admin_test")
		f, err := os.Create(testFile)
		if err == nil {
			f.Close()
			os.Remove(testFile)
			return true
		}
		return false
	case "darwin", "linux":
		return os.Getuid() == 0
	}
	return false
}

// getOSVersion 获取操作系统版本
func (p *Platform) getOSVersion() string {
	switch p.OS {
	case "windows":
		return p.getWindowsVersion()
	case "darwin":
		return p.getMacOSVersion()
	case "linux":
		return p.getLinuxVersion()
	}
	return "unknown"
}

// getWindowsVersion 获取 Windows 版本
func (p *Platform) getWindowsVersion() string {
	// 使用 ver 命令
	cmd := exec.Command("cmd", "/c", "ver")
	output, err := cmd.Output()
	if err == nil {
		version := strings.TrimSpace(string(output))
		// 提取版本号
		if idx := strings.Index(version, "["); idx != -1 {
			version = version[:idx]
		}
		return strings.TrimSpace(version)
	}
	return "Windows"
}

// getMacOSVersion 获取 macOS 版本
func (p *Platform) getMacOSVersion() string {
	cmd := exec.Command("sw_vers", "-productVersion")
	output, err := cmd.Output()
	if err == nil {
		return fmt.Sprintf("macOS %s", strings.TrimSpace(string(output)))
	}
	return "macOS"
}

// getLinuxVersion 获取 Linux 版本
func (p *Platform) getLinuxVersion() string {
	// 尝试读取 /etc/os-release
	data, err := os.ReadFile("/etc/os-release")
	if err == nil {
		content := string(data)
		// 查找 PRETTY_NAME
		if idx := strings.Index(content, "PRETTY_NAME="); idx != -1 {
			line := content[idx:]
			if endIdx := strings.Index(line, "\n"); endIdx != -1 {
				line = line[:endIdx]
			}
			// 提取值
			value := strings.TrimPrefix(line, "PRETTY_NAME=")
			value = strings.Trim(value, `"`)
			return value
		}
	}
	return "Linux"
}

// GetSystemInfo 获取系统详细信息
func (p *Platform) GetSystemInfo() SystemInfo {
	info := SystemInfo{
		CPUCount: runtime.NumCPU(),
	}

	// 获取磁盘空间
	if freeSpace := p.getFreeDiskSpace(); freeSpace > 0 {
		info.FreeDiskSpace = freeSpace
	}

	return info
}

// getFreeDiskSpace 获取磁盘剩余空间
func (p *Platform) getFreeDiskSpace() uint64 {
	var path string
	switch p.OS {
	case "windows":
		path = os.Getenv("SystemDrive")
		if path == "" {
			path = "C:"
		}
	default:
		path = "/"
	}

	// 使用 syscall 获取磁盘空间
	// 这里简化处理，实际实现需要平台特定的代码
	_ = path
	return 0
}

// normalizeArch 标准化架构名称
func normalizeArch(arch string) string {
	switch arch {
	case "amd64", "x86_64":
		return "amd64"
	case "arm64", "aarch64":
		return "arm64"
	case "386", "i386", "x86":
		return "386"
	case "arm":
		return "arm"
	}
	return arch
}

// IsSupportedArch 检查架构是否受支持
func IsSupportedArch(arch string) bool {
	switch normalizeArch(arch) {
	case "amd64", "arm64":
		return true
	}
	return false
}

// GetSupportedInstallModes 获取支持的安装模式
func (p *Platform) GetSupportedInstallModes() []map[string]string {
	modes := []map[string]string{
		{
			"id":          "system",
			"name":        "系统安装",
			"description": "安装到系统目录，所有用户可用",
		},
		{
			"id":          "user",
			"name":        "用户安装",
			"description": "安装到用户目录，仅当前用户可用",
		},
		{
			"id":          "portable",
			"name":        "便携模式",
			"description": "直接从当前位置运行，不复制文件",
		},
	}

	// 如果没有管理员权限，移除系统安装选项
	if !p.hasAdmin() {
		modes = modes[1:]
	}

	return modes
}
