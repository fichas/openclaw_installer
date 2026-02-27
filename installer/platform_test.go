package main

import (
	"runtime"
	"testing"
)

// TestDetectPlatform tests the platform detection function
func TestDetectPlatform(t *testing.T) {
	platform := DetectPlatform()

	if platform == nil {
		t.Fatal("DetectPlatform() returned nil")
	}

	// Verify OS is set
	if platform.OS == "" {
		t.Error("Platform.OS is empty")
	}

	// Verify Arch is set
	if platform.Arch == "" {
		t.Error("Platform.Arch is empty")
	}

	// Verify it matches runtime values
	if platform.OS != runtime.GOOS {
		t.Errorf("Platform.OS = %s, want %s", platform.OS, runtime.GOOS)
	}

	expectedArch := normalizeArch(runtime.GOARCH)
	if platform.Arch != expectedArch {
		t.Errorf("Platform.Arch = %s, want %s", platform.Arch, expectedArch)
	}
}

// TestNormalizeArch tests architecture normalization
func TestNormalizeArch(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"amd64", "amd64"},
		{"x86_64", "amd64"},
		{"arm64", "arm64"},
		{"aarch64", "arm64"},
		{"386", "386"},
		{"x86", "386"},
		{"unknown", "unknown"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalizeArch(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeArch(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestPlatformIsWindows tests Windows detection
func TestPlatformIsWindows(t *testing.T) {
	tests := []struct {
		os       string
		expected bool
	}{
		{"windows", true},
		{"darwin", false},
		{"linux", false},
		{"freebsd", false},
	}

	for _, tt := range tests {
		t.Run(tt.os, func(t *testing.T) {
			p := &Platform{OS: tt.os, Arch: "amd64"}
			if got := p.IsWindows(); got != tt.expected {
				t.Errorf("Platform.IsWindows() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestPlatformIsMacOS tests macOS detection
func TestPlatformIsMacOS(t *testing.T) {
	tests := []struct {
		os       string
		expected bool
	}{
		{"darwin", true},
		{"windows", false},
		{"linux", false},
		{"freebsd", false},
	}

	for _, tt := range tests {
		t.Run(tt.os, func(t *testing.T) {
			p := &Platform{OS: tt.os, Arch: "amd64"}
			if got := p.IsMacOS(); got != tt.expected {
				t.Errorf("Platform.IsMacOS() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestPlatformIsLinux tests Linux detection
func TestPlatformIsLinux(t *testing.T) {
	tests := []struct {
		os       string
		expected bool
	}{
		{"linux", true},
		{"windows", false},
		{"darwin", false},
		{"freebsd", false},
	}

	for _, tt := range tests {
		t.Run(tt.os, func(t *testing.T) {
			p := &Platform{OS: tt.os, Arch: "amd64"}
			if got := p.IsLinux(); got != tt.expected {
				t.Errorf("Platform.IsLinux() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestPlatformGetInstallDir tests installation directory paths
func TestPlatformGetInstallDir(t *testing.T) {
	tests := []struct {
		os       string
		expected string
	}{
		{"windows", `C:\Program Files\OpenClaw`},
		{"darwin", "/usr/local/bin"},
		{"linux", "/usr/local/bin"},
		{"freebsd", "/usr/local/bin"},
	}

	for _, tt := range tests {
		t.Run(tt.os, func(t *testing.T) {
			p := &Platform{OS: tt.os, Arch: "amd64"}
			if got := p.GetInstallDir(); got != tt.expected {
				t.Errorf("Platform.GetInstallDir() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestPlatformGetConfigDir tests configuration directory paths
func TestPlatformGetConfigDir(t *testing.T) {
	tests := []struct {
		os       string
		expected string
	}{
		{"windows", `C:\ProgramData\OpenClaw`},
		{"darwin", "/etc/openclaw"},
		{"linux", "/etc/openclaw"},
		{"freebsd", "/etc/openclaw"},
	}

	for _, tt := range tests {
		t.Run(tt.os, func(t *testing.T) {
			p := &Platform{OS: tt.os, Arch: "amd64"}
			if got := p.GetConfigDir(); got != tt.expected {
				t.Errorf("Platform.GetConfigDir() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestPlatformGetBinaryName tests binary name generation
func TestPlatformGetBinaryName(t *testing.T) {
	tests := []struct {
		os       string
		name     string
		expected string
	}{
		{"windows", "openclaw", "openclaw.exe"},
		{"darwin", "openclaw", "openclaw"},
		{"linux", "openclaw", "openclaw"},
		{"windows", "adapter", "adapter.exe"},
		{"linux", "adapter", "adapter"},
	}

	for _, tt := range tests {
		t.Run(tt.os+"_"+tt.name, func(t *testing.T) {
			p := &Platform{OS: tt.os, Arch: "amd64"}
			if got := p.GetBinaryName(tt.name); got != tt.expected {
				t.Errorf("Platform.GetBinaryName(%q) = %q, want %q", tt.name, got, tt.expected)
			}
		})
	}
}

// TestPlatformString tests string representation
func TestPlatformString(t *testing.T) {
	p := &Platform{OS: "linux", Arch: "amd64"}
	expected := "linux/amd64"
	if got := p.String(); got != expected {
		t.Errorf("Platform.String() = %q, want %q", got, expected)
	}
}

// TestCrossPlatformSupport tests all supported platform combinations
func TestCrossPlatformSupport(t *testing.T) {
	supportedPlatforms := []struct {
		os   string
		arch string
	}{
		{"windows", "amd64"},
		{"windows", "arm64"},
		{"darwin", "amd64"},
		{"darwin", "arm64"},
		{"linux", "amd64"},
		{"linux", "arm64"},
	}

	for _, tt := range supportedPlatforms {
		t.Run(tt.os+"_"+tt.arch, func(t *testing.T) {
			p := &Platform{OS: tt.os, Arch: tt.arch}

			// Verify no panic on method calls
			_ = p.IsWindows()
			_ = p.IsMacOS()
			_ = p.IsLinux()
			_ = p.GetInstallDir()
			_ = p.GetConfigDir()
			_ = p.GetBinaryName("test")
			_ = p.String()
		})
	}
}

// MockPlatform allows testing with specific platform values
func MockPlatform(goos, goarch string) *Platform {
	return &Platform{
		OS:   goos,
		Arch: normalizeArch(goarch),
	}
}

// TestMockPlatform tests the mock platform helper
func TestMockPlatform(t *testing.T) {
	// Test mocking different platforms
	platforms := []struct {
		goos   string
		goarch string
	}{
		{"windows", "amd64"},
		{"darwin", "arm64"},
		{"linux", "amd64"},
	}

	for _, p := range platforms {
		mock := MockPlatform(p.goos, p.goarch)
		if mock.OS != p.goos {
			t.Errorf("MockPlatform OS = %s, want %s", mock.OS, p.goos)
		}
		if mock.Arch != normalizeArch(p.goarch) {
			t.Errorf("MockPlatform Arch = %s, want %s", mock.Arch, normalizeArch(p.goarch))
		}
	}
}
