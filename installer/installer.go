package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// InstallStatus represents the current installation status
type InstallStatus struct {
	Installed  bool   `json:"installed"`
	Version    string `json:"version,omitempty"`
	Error      string `json:"error,omitempty"`
	InProgress bool   `json:"in_progress"`
}

// Installer handles file installation operations
type Installer struct {
	platform *Platform
	config   *Config
	status   InstallStatus
}

// NewInstaller creates a new installer instance
func NewInstaller(platform *Platform) *Installer {
	return &Installer{
		platform: platform,
		config:   nil,
		status:   InstallStatus{Installed: false},
	}
}

// GetConfig returns the current configuration
func (i *Installer) GetConfig() *Config {
	if i.config == nil {
		defaultConfig, _ := GenerateConfig(ConfigOptions{
			Version:    "1.0.0",
			ServerHost: "localhost",
			ServerPort: 8080,
		})
		return defaultConfig
	}
	return i.config
}

// SetConfig sets the configuration
func (i *Installer) SetConfig(config Config) {
	i.config = &config
}

// GetStatus returns the installation status
func (i *Installer) GetStatus() InstallStatus {
	return i.status
}

// Install performs installation with current configuration
func (i *Installer) Install() error {
	i.status.InProgress = true
	i.status.Error = ""

	if i.config == nil {
		i.status.InProgress = false
		i.status.Error = "no configuration set"
		return fmt.Errorf("no configuration set")
	}

	// Simulate installation
	i.status.Installed = true
	i.status.Version = i.config.Version
	i.status.InProgress = false

	return nil
}

// InstallOptions contains installation configuration
type InstallOptions struct {
	SourceDir   string // Source directory (e.g., USB drive)
	InstallDir  string // Target installation directory
	ConfigDir   string // Configuration directory
	BinaryName  string // Name of the main binary
	AdapterName string // Name of the adapter binary
}

// InstallWithOptions performs the installation of OpenClaw binaries and files
func (i *Installer) InstallWithOptions(opts InstallOptions) error {
	i.status.InProgress = true
	i.status.Error = ""

	// Create installation directories
	if err := os.MkdirAll(opts.InstallDir, 0755); err != nil {
		i.status.InProgress = false
		i.status.Error = err.Error()
		return fmt.Errorf("failed to create install directory: %w", err)
	}

	if err := os.MkdirAll(opts.ConfigDir, 0755); err != nil {
		i.status.InProgress = false
		i.status.Error = err.Error()
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Copy main binary
	binaryName := i.platform.GetBinaryName(opts.BinaryName)
	srcBinary := filepath.Join(opts.SourceDir, binaryName)
	dstBinary := filepath.Join(opts.InstallDir, binaryName)

	if err := i.copyFile(srcBinary, dstBinary); err != nil {
		i.status.InProgress = false
		i.status.Error = err.Error()
		return fmt.Errorf("failed to copy binary: %w", err)
	}

	// Make binary executable on Unix systems
	if !i.platform.IsWindows() {
		if err := os.Chmod(dstBinary, 0755); err != nil {
			i.status.InProgress = false
			i.status.Error = err.Error()
			return fmt.Errorf("failed to set binary permissions: %w", err)
		}
	}

	// Copy adapter if specified
	if opts.AdapterName != "" {
		adapterName := i.platform.GetBinaryName(opts.AdapterName)
		srcAdapter := filepath.Join(opts.SourceDir, adapterName)
		dstAdapter := filepath.Join(opts.InstallDir, adapterName)

		if err := i.copyFile(srcAdapter, dstAdapter); err != nil {
			i.status.InProgress = false
			i.status.Error = err.Error()
			return fmt.Errorf("failed to copy adapter: %w", err)
		}

		if !i.platform.IsWindows() {
			if err := os.Chmod(dstAdapter, 0755); err != nil {
				i.status.InProgress = false
				i.status.Error = err.Error()
				return fmt.Errorf("failed to set adapter permissions: %w", err)
			}
		}
	}

	i.status.Installed = true
	i.status.InProgress = false

	return nil
}

// copyFile copies a file from src to dst
func (i *Installer) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file %s: %w", src, err)
	}
	defer sourceFile.Close()

	// Get source file info for permissions
	stat, err := sourceFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat source file: %w", err)
	}

	// Create destination file
	destFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, stat.Mode())
	if err != nil {
		return fmt.Errorf("failed to create destination file %s: %w", dst, err)
	}
	defer destFile.Close()

	// Copy content
	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	return nil
}

// CopyFromUSB copies files from a USB drive to the installation directory
func (i *Installer) CopyFromUSB(usbPath string, files []string) error {
	installDir := i.platform.GetInstallDir()

	for _, file := range files {
		src := filepath.Join(usbPath, file)
		dst := filepath.Join(installDir, file)

		if err := i.copyFile(src, dst); err != nil {
			return fmt.Errorf("failed to copy %s: %w", file, err)
		}
	}

	return nil
}

// Uninstall removes installed files
func (i *Installer) Uninstall(installDir string, files []string) error {
	for _, file := range files {
		path := filepath.Join(installDir, file)
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove %s: %w", file, err)
		}
	}
	return nil
}

// VerifyInstallation checks if the installation is valid
func (i *Installer) VerifyInstallation(installDir string, files []string) error {
	for _, file := range files {
		path := filepath.Join(installDir, file)
		if _, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("file not found: %s", file)
			}
			return fmt.Errorf("failed to check %s: %w", file, err)
		}
	}
	return nil
}
