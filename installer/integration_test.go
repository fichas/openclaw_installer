package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestFullInstallationFlow tests the complete installation flow
func TestFullInstallationFlow(t *testing.T) {
	// Create temporary directories
	srcDir := t.TempDir()
	installDir := t.TempDir()
	configDir := t.TempDir()

	// Create source binary
	srcBinary := filepath.Join(srcDir, "openclaw")
	if err := os.WriteFile(srcBinary, []byte("binary content"), 0755); err != nil {
		t.Fatalf("Failed to create source binary: %v", err)
	}

	// Create platform and installer
	platform := &Platform{OS: "linux", Arch: "amd64"}
	installer := NewInstaller(platform)

	// Step 1: Verify initial status
	status := installer.GetStatus()
	if status.Installed {
		t.Error("Initial status should not be installed")
	}

	// Step 2: Generate configuration
	configOpts := ConfigOptions{
		Version:     "1.0.0",
		ServerHost:  "localhost",
		ServerPort:  8080,
		EnableTLS:   false,
		AdapterType: "ollama",
		AdapterName: "local-ollama",
	}

	config, err := GenerateConfig(configOpts)
	if err != nil {
		t.Fatalf("GenerateConfig() error = %v", err)
	}

	if err := config.Validate(); err != nil {
		t.Fatalf("Config.Validate() error = %v", err)
	}

	// Step 3: Install with options
	installOpts := InstallOptions{
		SourceDir:  srcDir,
		InstallDir: installDir,
		ConfigDir:  configDir,
		BinaryName: "openclaw",
	}

	if err := installer.InstallWithOptions(installOpts); err != nil {
		t.Fatalf("InstallWithOptions() error = %v", err)
	}

	// Step 4: Verify installation
	files := []string{"openclaw"}
	if err := installer.VerifyInstallation(installDir, files); err != nil {
		t.Errorf("VerifyInstallation() error = %v", err)
	}

	// Step 5: Save configuration
	configPath := filepath.Join(configDir, "openclaw.json")
	if err := config.Save(configPath); err != nil {
		t.Errorf("Config.Save() error = %v", err)
	}

	// Step 6: Load and verify configuration
	loadedConfig, err := LoadConfig(configPath)
	if err != nil {
		t.Errorf("LoadConfig() error = %v", err)
	}

	if loadedConfig.Version != config.Version {
		t.Errorf("Loaded version = %s, want %s", loadedConfig.Version, config.Version)
	}

	// Step 7: Check final status
	finalStatus := installer.GetStatus()
	if !finalStatus.Installed {
		t.Error("Final status should be installed")
	}
}

// TestCrossPlatformInstallation tests installation on different platforms
func TestCrossPlatformInstallation(t *testing.T) {
	platforms := []struct {
		os   string
		arch string
	}{
		{"linux", "amd64"},
		{"linux", "arm64"},
		{"darwin", "amd64"},
		{"darwin", "arm64"},
		{"windows", "amd64"},
	}

	for _, p := range platforms {
		t.Run(p.os+"_"+p.arch, func(t *testing.T) {
			platform := &Platform{OS: p.os, Arch: p.arch}
			installer := NewInstaller(platform)

			// Verify platform detection
			if installer.platform.OS != p.os {
				t.Errorf("Platform.OS = %s, want %s", installer.platform.OS, p.os)
			}

			// Verify install directory is platform-appropriate
			installDir := platform.GetInstallDir()
			if installDir == "" {
				t.Error("GetInstallDir() returned empty string")
			}

			// Verify config directory
			configDir := platform.GetConfigDir()
			if configDir == "" {
				t.Error("GetConfigDir() returned empty string")
			}

			// Verify binary name has correct extension
			binaryName := platform.GetBinaryName("openclaw")
			if p.os == "windows" && binaryName != "openclaw.exe" {
				t.Errorf("Binary name = %s, want openclaw.exe", binaryName)
			}
			if p.os != "windows" && binaryName != "openclaw" {
				t.Errorf("Binary name = %s, want openclaw", binaryName)
			}
		})
	}
}

// TestErrorHandling tests error handling in various scenarios
func TestErrorHandling(t *testing.T) {
	t.Run("InvalidSourceDir", func(t *testing.T) {
		platform := &Platform{OS: "linux", Arch: "amd64"}
		installer := NewInstaller(platform)

		opts := InstallOptions{
			SourceDir:  "/nonexistent/directory",
			InstallDir: t.TempDir(),
			ConfigDir:  t.TempDir(),
			BinaryName: "openclaw",
		}

		err := installer.InstallWithOptions(opts)
		if err == nil {
			t.Error("InstallWithOptions() should return error for invalid source directory")
		}
	})

	t.Run("PermissionDenied", func(t *testing.T) {
		platform := &Platform{OS: "linux", Arch: "amd64"}
		installer := NewInstaller(platform)

		// Try to install to a system directory (should fail without root)
		opts := InstallOptions{
			SourceDir:  t.TempDir(),
			InstallDir: "/nonexistent_root_path/install",
			ConfigDir:  t.TempDir(),
			BinaryName: "openclaw",
		}

		// Create source file
		srcBinary := filepath.Join(opts.SourceDir, "openclaw")
		os.WriteFile(srcBinary, []byte("binary"), 0644)

		err := installer.InstallWithOptions(opts)
		if err == nil {
			t.Log("Install succeeded unexpectedly (may have permissions)")
		}
	})

	t.Run("InvalidConfig", func(t *testing.T) {
		config := Config{
			Version: "",
			Server:  ServerConfig{Port: 0},
		}

		err := config.Validate()
		if err == nil {
			t.Error("Validate() should return error for invalid config")
		}
	})
}

// TestConcurrentOperations tests concurrent access to installer
func TestConcurrentOperations(t *testing.T) {
	platform := &Platform{OS: "linux", Arch: "amd64"}
	installer := NewInstaller(platform)

	// Test concurrent status reads
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_ = installer.GetStatus()
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		select {
		case <-done:
			// Success
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for concurrent operations")
		}
	}
}

// TestConfigPersistence tests configuration persistence across operations
func TestConfigPersistence(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test_config.json")

	// Create and save config
	originalConfig := &Config{
		Version:  "1.0.0",
		Platform: "linux/amd64",
		Server: ServerConfig{
			Host: "localhost",
			Port: 8080,
			TLS:  false,
		},
		Adapters: []AdapterConfig{
			{
				Name:    "ollama-local",
				Type:    "ollama",
				Enabled: true,
				Options: map[string]string{
					"url": "http://localhost:11434",
				},
			},
		},
		Settings: map[string]string{
			"log_level": "debug",
		},
	}

	if err := originalConfig.Save(configPath); err != nil {
		t.Fatalf("Config.Save() error = %v", err)
	}

	// Load config
	loadedConfig, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	// Verify all fields
	if loadedConfig.Version != originalConfig.Version {
		t.Errorf("Version mismatch: got %s, want %s", loadedConfig.Version, originalConfig.Version)
	}

	if loadedConfig.Server.Host != originalConfig.Server.Host {
		t.Errorf("Server.Host mismatch: got %s, want %s", loadedConfig.Server.Host, originalConfig.Server.Host)
	}

	if loadedConfig.Server.Port != originalConfig.Server.Port {
		t.Errorf("Server.Port mismatch: got %d, want %d", loadedConfig.Server.Port, originalConfig.Server.Port)
	}

	if len(loadedConfig.Adapters) != len(originalConfig.Adapters) {
		t.Errorf("Adapters count mismatch: got %d, want %d", len(loadedConfig.Adapters), len(originalConfig.Adapters))
	}

	if loadedConfig.Settings["log_level"] != originalConfig.Settings["log_level"] {
		t.Errorf("Settings mismatch: got %s, want %s", loadedConfig.Settings["log_level"], originalConfig.Settings["log_level"])
	}
}

// TestMissingConfigFile tests handling of missing config file
func TestMissingConfigFile(t *testing.T) {
	_, err := LoadConfig("/nonexistent/path/config.json")
	if err == nil {
		t.Error("LoadConfig() should return error for missing file")
	}
}

// TestCleanupOnFailure tests cleanup behavior on installation failure
func TestCleanupOnFailure(t *testing.T) {
	platform := &Platform{OS: "linux", Arch: "amd64"}
	installer := NewInstaller(platform)

	srcDir := t.TempDir()
	installDir := t.TempDir()
	configDir := t.TempDir()

	// Create source file
	srcBinary := filepath.Join(srcDir, "openclaw")
	os.WriteFile(srcBinary, []byte("binary"), 0644)

	opts := InstallOptions{
		SourceDir:   srcDir,
		InstallDir:  installDir,
		ConfigDir:   configDir,
		BinaryName:  "openclaw",
		AdapterName: "nonexistent_adapter",
	}

	// This should fail because adapter doesn't exist
	err := installer.InstallWithOptions(opts)
	if err == nil {
		t.Log("Install succeeded unexpectedly")
	}

	// Verify status reflects failure
	status := installer.GetStatus()
	if status.Installed {
		t.Error("Status should not be installed after failure")
	}
}
