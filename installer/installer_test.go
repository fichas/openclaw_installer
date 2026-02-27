package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestNewInstaller tests installer creation
func TestNewInstaller(t *testing.T) {
	platform := &Platform{OS: "linux", Arch: "amd64"}
	installer := NewInstaller(platform)

	if installer == nil {
		t.Fatal("NewInstaller() returned nil")
	}

	if installer.platform != platform {
		t.Error("Installer.platform not set correctly")
	}
}

// TestInstallerGetConfig tests getting configuration
func TestInstallerGetConfig(t *testing.T) {
	platform := &Platform{OS: "linux", Arch: "amd64"}
	installer := NewInstaller(platform)

	// Get default config when none is set
	config := installer.GetConfig()
	if config == nil {
		t.Fatal("GetConfig() returned nil")
	}

	// Verify default values
	if config.Version == "" {
		t.Error("Default config has no version")
	}
}

// TestInstallerSetConfig tests setting configuration
func TestInstallerSetConfig(t *testing.T) {
	platform := &Platform{OS: "linux", Arch: "amd64"}
	installer := NewInstaller(platform)

	newConfig := Config{
		Version: "2.0.0",
		Server:  ServerConfig{Port: 9090},
		Adapters: []AdapterConfig{
			{Name: "test", Type: "ollama"},
		},
	}

	installer.SetConfig(newConfig)

	retrievedConfig := installer.GetConfig()
	if retrievedConfig.Version != newConfig.Version {
		t.Errorf("GetConfig().Version = %s, want %s", retrievedConfig.Version, newConfig.Version)
	}
}

// TestInstallerGetStatus tests getting installation status
func TestInstallerGetStatus(t *testing.T) {
	platform := &Platform{OS: "linux", Arch: "amd64"}
	installer := NewInstaller(platform)

	status := installer.GetStatus()
	if status.Installed {
		t.Error("New installer should not be marked as installed")
	}
}

// TestInstallerInstall tests installation process
func TestInstallerInstall(t *testing.T) {
	platform := &Platform{OS: "linux", Arch: "amd64"}
	installer := NewInstaller(platform)

	// Set config first
	installer.SetConfig(Config{
		Version: "1.0.0",
		Server:  ServerConfig{Port: 8080},
		Adapters: []AdapterConfig{
			{Name: "test", Type: "ollama"},
		},
	})

	err := installer.Install()
	if err != nil {
		t.Errorf("Install() error = %v", err)
	}

	status := installer.GetStatus()
	if !status.Installed {
		t.Error("Status.Installed should be true after successful install")
	}
}

// TestInstallerInstallNoConfig tests installation without config
func TestInstallerInstallNoConfig(t *testing.T) {
	platform := &Platform{OS: "linux", Arch: "amd64"}
	installer := NewInstaller(platform)

	err := installer.Install()
	if err == nil {
		t.Error("Install() should return error when no config is set")
	}
}

// TestCopyFile tests file copying
func TestCopyFile(t *testing.T) {
	platform := &Platform{OS: "linux", Arch: "amd64"}
	installer := NewInstaller(platform)

	// Create temp directory
	tmpDir := t.TempDir()

	// Create source file
	srcPath := filepath.Join(tmpDir, "source.txt")
	content := []byte("test content")
	if err := os.WriteFile(srcPath, content, 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Copy file
	dstPath := filepath.Join(tmpDir, "dest.txt")
	if err := installer.copyFile(srcPath, dstPath); err != nil {
		t.Errorf("copyFile() error = %v", err)
	}

	// Verify destination exists
	if _, err := os.Stat(dstPath); os.IsNotExist(err) {
		t.Error("Destination file was not created")
	}

	// Verify content
	copiedContent, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("Failed to read destination file: %v", err)
	}

	if string(copiedContent) != string(content) {
		t.Errorf("Copied content = %s, want %s", string(copiedContent), string(content))
	}
}

// TestCopyFileNotExist tests copying non-existent file
func TestCopyFileNotExist(t *testing.T) {
	platform := &Platform{OS: "linux", Arch: "amd64"}
	installer := NewInstaller(platform)

	tmpDir := t.TempDir()
	srcPath := filepath.Join(tmpDir, "nonexistent.txt")
	dstPath := filepath.Join(tmpDir, "dest.txt")

	err := installer.copyFile(srcPath, dstPath)
	if err == nil {
		t.Error("copyFile() should return error for non-existent source")
	}
}

// TestInstallWithOptions tests installation with options
func TestInstallWithOptions(t *testing.T) {
	platform := &Platform{OS: "linux", Arch: "amd64"}
	installer := NewInstaller(platform)

	// Create temp directories
	srcDir := t.TempDir()
	installDir := t.TempDir()
	configDir := t.TempDir()

	// Create source binary
	srcBinary := filepath.Join(srcDir, "openclaw")
	if err := os.WriteFile(srcBinary, []byte("binary content"), 0755); err != nil {
		t.Fatalf("Failed to create source binary: %v", err)
	}

	opts := InstallOptions{
		SourceDir:  srcDir,
		InstallDir: installDir,
		ConfigDir:  configDir,
		BinaryName: "openclaw",
	}

	err := installer.InstallWithOptions(opts)
	if err != nil {
		t.Errorf("InstallWithOptions() error = %v", err)
	}

	// Verify binary was copied
	dstBinary := filepath.Join(installDir, "openclaw")
	if _, err := os.Stat(dstBinary); os.IsNotExist(err) {
		t.Error("Binary was not copied to install directory")
	}
}

// TestInstallWithAdapter tests installation with adapter
func TestInstallWithAdapter(t *testing.T) {
	platform := &Platform{OS: "linux", Arch: "amd64"}
	installer := NewInstaller(platform)

	// Create temp directories
	srcDir := t.TempDir()
	installDir := t.TempDir()
	configDir := t.TempDir()

	// Create source files
	srcBinary := filepath.Join(srcDir, "openclaw")
	srcAdapter := filepath.Join(srcDir, "adapter")
	if err := os.WriteFile(srcBinary, []byte("binary"), 0755); err != nil {
		t.Fatalf("Failed to create source binary: %v", err)
	}
	if err := os.WriteFile(srcAdapter, []byte("adapter"), 0755); err != nil {
		t.Fatalf("Failed to create source adapter: %v", err)
	}

	opts := InstallOptions{
		SourceDir:   srcDir,
		InstallDir:  installDir,
		ConfigDir:   configDir,
		BinaryName:  "openclaw",
		AdapterName: "adapter",
	}

	err := installer.InstallWithOptions(opts)
	if err != nil {
		t.Errorf("InstallWithOptions() error = %v", err)
	}

	// Verify both files were copied
	dstBinary := filepath.Join(installDir, "openclaw")
	dstAdapter := filepath.Join(installDir, "adapter")

	if _, err := os.Stat(dstBinary); os.IsNotExist(err) {
		t.Error("Binary was not copied")
	}
	if _, err := os.Stat(dstAdapter); os.IsNotExist(err) {
		t.Error("Adapter was not copied")
	}
}

// TestCopyFromUSB tests copying files from USB
func TestCopyFromUSB(t *testing.T) {
	platform := &Platform{OS: "linux", Arch: "amd64"}
	installer := NewInstaller(platform)

	// Create temp directories
	usbDir := t.TempDir()
	installDir := t.TempDir()

	// Override the install dir (we need to modify the test to use a mock)
	// For now, we'll test the copyFile function directly

	// Create source files
	files := []string{"file1.txt", "file2.txt"}
	for _, file := range files {
		path := filepath.Join(usbDir, file)
		if err := os.WriteFile(path, []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to create source file: %v", err)
		}
	}

	// Copy files directly using copyFile
	for _, file := range files {
		src := filepath.Join(usbDir, file)
		dst := filepath.Join(installDir, file)
		if err := installer.copyFile(src, dst); err != nil {
			t.Errorf("copyFile(%s) error = %v", file, err)
		}
	}

	// Verify files were copied
	for _, file := range files {
		path := filepath.Join(installDir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("File %s was not copied", file)
		}
	}
}

// TestUninstall tests file removal
func TestUninstall(t *testing.T) {
	platform := &Platform{OS: "linux", Arch: "amd64"}
	installer := NewInstaller(platform)

	// Create temp directory with files
	installDir := t.TempDir()
	files := []string{"file1.txt", "file2.txt"}

	for _, file := range files {
		path := filepath.Join(installDir, file)
		if err := os.WriteFile(path, []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}
	}

	// Uninstall
	err := installer.Uninstall(installDir, files)
	if err != nil {
		t.Errorf("Uninstall() error = %v", err)
	}

	// Verify files were removed
	for _, file := range files {
		path := filepath.Join(installDir, file)
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("File %s was not removed", file)
		}
	}
}

// TestVerifyInstallation tests installation verification
func TestVerifyInstallation(t *testing.T) {
	platform := &Platform{OS: "linux", Arch: "amd64"}
	installer := NewInstaller(platform)

	// Create temp directory with files
	installDir := t.TempDir()
	files := []string{"file1.txt", "file2.txt"}

	for _, file := range files {
		path := filepath.Join(installDir, file)
		if err := os.WriteFile(path, []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}
	}

	// Verify installation
	err := installer.VerifyInstallation(installDir, files)
	if err != nil {
		t.Errorf("VerifyInstallation() error = %v", err)
	}
}

// TestVerifyInstallationMissingFile tests verification with missing file
func TestVerifyInstallationMissingFile(t *testing.T) {
	platform := &Platform{OS: "linux", Arch: "amd64"}
	installer := NewInstaller(platform)

	installDir := t.TempDir()
	files := []string{"existing.txt", "missing.txt"}

	// Create only one file
	if err := os.WriteFile(filepath.Join(installDir, "existing.txt"), []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	err := installer.VerifyInstallation(installDir, files)
	if err == nil {
		t.Error("VerifyInstallation() should return error for missing file")
	}
}

// TestUninstallNonExistentFile tests uninstalling non-existent file
func TestUninstallNonExistentFile(t *testing.T) {
	platform := &Platform{OS: "linux", Arch: "amd64"}
	installer := NewInstaller(platform)

	installDir := t.TempDir()
	files := []string{"nonexistent.txt"}

	// Should not error for non-existent files
	err := installer.Uninstall(installDir, files)
	if err != nil {
		t.Errorf("Uninstall() error = %v", err)
	}
}

// TestInstallWithWindowsBinary tests Windows binary name handling
func TestInstallWithWindowsBinary(t *testing.T) {
	platform := &Platform{OS: "windows", Arch: "amd64"}
	installer := NewInstaller(platform)

	srcDir := t.TempDir()
	installDir := t.TempDir()
	configDir := t.TempDir()

	// Create source binary with .exe extension
	srcBinary := filepath.Join(srcDir, "openclaw.exe")
	if err := os.WriteFile(srcBinary, []byte("binary"), 0755); err != nil {
		t.Fatalf("Failed to create source binary: %v", err)
	}

	opts := InstallOptions{
		SourceDir:  srcDir,
		InstallDir: installDir,
		ConfigDir:  configDir,
		BinaryName: "openclaw",
	}

	err := installer.InstallWithOptions(opts)
	if err != nil {
		t.Errorf("InstallWithOptions() error = %v", err)
	}

	// Verify binary was copied with .exe extension
	dstBinary := filepath.Join(installDir, "openclaw.exe")
	if _, err := os.Stat(dstBinary); os.IsNotExist(err) {
		t.Error("Windows binary was not copied correctly")
	}
}

// TestCopyFilePermissions tests that file permissions are preserved
func TestCopyFilePermissions(t *testing.T) {
	platform := &Platform{OS: "linux", Arch: "amd64"}
	installer := NewInstaller(platform)

	tmpDir := t.TempDir()

	// Create source file with specific permissions
	srcPath := filepath.Join(tmpDir, "source.sh")
	content := []byte("#!/bin/bash\necho hello")
	if err := os.WriteFile(srcPath, content, 0755); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Copy file
	dstPath := filepath.Join(tmpDir, "dest.sh")
	if err := installer.copyFile(srcPath, dstPath); err != nil {
		t.Fatalf("copyFile() error = %v", err)
	}

	// Verify permissions (on Unix systems)
	srcInfo, err := os.Stat(srcPath)
	if err != nil {
		t.Fatalf("Failed to stat source file: %v", err)
	}

	dstInfo, err := os.Stat(dstPath)
	if err != nil {
		t.Fatalf("Failed to stat destination file: %v", err)
	}

	if srcInfo.Mode() != dstInfo.Mode() {
		t.Errorf("Permissions not preserved: src=%v, dst=%v", srcInfo.Mode(), dstInfo.Mode())
	}
}
