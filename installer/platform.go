package main

import (
	"fmt"
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
