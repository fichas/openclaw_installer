package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// Platform represents the detected platform information
type Platform struct {
	OS   string // windows, darwin, linux
	Arch string // amd64, arm64, 386
}

// DetectPlatform detects the current operating system and architecture
func DetectPlatform() *Platform {
	return &Platform{
		OS:   runtime.GOOS,
		Arch: normalizeArch(runtime.GOARCH),
	}
}

// normalizeArch normalizes architecture names
func normalizeArch(arch string) string {
	switch arch {
	case "amd64", "x86_64":
		return "amd64"
	case "arm64", "aarch64":
		return "arm64"
	case "386", "x86":
		return "386"
	default:
		return arch
	}
}

// IsWindows returns true if the platform is Windows
func (p *Platform) IsWindows() bool {
	return p.OS == "windows"
}

// IsMacOS returns true if the platform is macOS
func (p *Platform) IsMacOS() bool {
	return p.OS == "darwin"
}

// IsLinux returns true if the platform is Linux
func (p *Platform) IsLinux() bool {
	return p.OS == "linux"
}

// GetInstallDir returns the default installation directory for the platform
func (p *Platform) GetInstallDir() string {
	switch p.OS {
	case "windows":
		return `C:\Program Files\OpenClaw`
	case "darwin":
		return "/usr/local/bin"
	case "linux":
		return "/usr/local/bin"
	default:
		return "/usr/local/bin"
	}
}

// GetConfigDir returns the configuration directory for the platform
func (p *Platform) GetConfigDir() string {
	switch p.OS {
	case "windows":
		return `C:\ProgramData\OpenClaw`
	case "darwin":
		return "/etc/openclaw"
	case "linux":
		return "/etc/openclaw"
	default:
		return "/etc/openclaw"
	}
}

// GetConfigPath returns the default configuration file path
func (p *Platform) GetConfigPath() string {
	return filepath.Join(p.GetConfigDir(), "openclaw.json")
}

// GetBackupDir returns the default backup directory for the platform
func (p *Platform) GetBackupDir() string {
	switch p.OS {
	case "windows":
		return `C:\ProgramData\OpenClaw\backups`
	case "darwin":
		return "/var/lib/openclaw/backups"
	case "linux":
		return "/var/lib/openclaw/backups"
	default:
		return "/var/lib/openclaw/backups"
	}
}

// GetDataDir returns the default data directory for the platform
func (p *Platform) GetDataDir() string {
	switch p.OS {
	case "windows":
		return `C:\ProgramData\OpenClaw\data`
	case "darwin":
		return "/var/lib/openclaw"
	case "linux":
		return "/var/lib/openclaw"
	default:
		return "/var/lib/openclaw"
	}
}

// GetLogDir returns the default log directory for the platform
func (p *Platform) GetLogDir() string {
	switch p.OS {
	case "windows":
		return `C:\ProgramData\OpenClaw\logs`
	case "darwin":
		return "/var/log/openclaw"
	case "linux":
		return "/var/log/openclaw"
	default:
		return "/var/log/openclaw"
	}
}

// GetBinaryName returns the binary name with extension for the platform
func (p *Platform) GetBinaryName(name string) string {
	if p.IsWindows() {
		return name + ".exe"
	}
	return name
}

// String returns a string representation of the platform
func (p *Platform) String() string {
	return fmt.Sprintf("%s/%s", p.OS, p.Arch)
}

// GetPackageExtension returns the package extension for the platform
func (p *Platform) GetPackageExtension() string {
	if p.IsWindows() {
		return ".zip"
	}
	return ".tar.gz"
}

// GetPackageName generates the package name for a component
func (p *Platform) GetPackageName(component string) string {
	ext := p.GetPackageExtension()
	return fmt.Sprintf("%s-%s-%s%s", component, p.OS, p.Arch, ext)
}

// EnsureDirectories creates all necessary directories
func (p *Platform) EnsureDirectories() error {
	dirs := []string{
		p.GetInstallDir(),
		p.GetConfigDir(),
		p.GetDataDir(),
		p.GetLogDir(),
		p.GetBackupDir(),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}
