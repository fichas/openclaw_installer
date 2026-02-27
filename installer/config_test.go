package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestGenerateConfig tests configuration generation
func TestGenerateConfig(t *testing.T) {
	opts := ConfigOptions{
		Version:      "1.0.0",
		ServerHost:   "localhost",
		ServerPort:   8080,
		EnableTLS:    true,
		AdapterType:  "ollama",
		AdapterName:  "local-ollama",
		CustomSettings: map[string]string{
			"log_level": "debug",
		},
	}

	config, err := GenerateConfig(opts)
	if err != nil {
		t.Fatalf("GenerateConfig() error = %v", err)
	}

	if config == nil {
		t.Fatal("GenerateConfig() returned nil")
	}

	// Verify version
	if config.Version != opts.Version {
		t.Errorf("Config.Version = %s, want %s", config.Version, opts.Version)
	}

	// Verify server configuration
	if config.Server.Host != opts.ServerHost {
		t.Errorf("Config.Server.Host = %s, want %s", config.Server.Host, opts.ServerHost)
	}

	if config.Server.Port != opts.ServerPort {
		t.Errorf("Config.Server.Port = %d, want %d", config.Server.Port, opts.ServerPort)
	}

	if config.Server.TLS != opts.EnableTLS {
		t.Errorf("Config.Server.TLS = %v, want %v", config.Server.TLS, opts.EnableTLS)
	}

	// Verify adapters
	if len(config.Adapters) != 1 {
		t.Fatalf("len(Config.Adapters) = %d, want 1", len(config.Adapters))
	}

	adapter := config.Adapters[0]
	if adapter.Type != opts.AdapterType {
		t.Errorf("Adapter.Type = %s, want %s", adapter.Type, opts.AdapterType)
	}

	if adapter.Name != opts.AdapterName {
		t.Errorf("Adapter.Name = %s, want %s", adapter.Name, opts.AdapterName)
	}

	if !adapter.Enabled {
		t.Error("Adapter.Enabled = false, want true")
	}

	// Verify platform is set
	if config.Platform == "" {
		t.Error("Config.Platform is empty")
	}
}

// TestConfigValidate tests configuration validation
func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: Config{
				Version: "1.0.0",
				Server: ServerConfig{
					Host: "localhost",
					Port: 8080,
				},
				Adapters: []AdapterConfig{
					{Name: "test", Type: "ollama", Enabled: true},
				},
			},
			wantErr: false,
		},
		{
			name: "missing version",
			config: Config{
				Version: "",
				Server:  ServerConfig{Port: 8080},
				Adapters: []AdapterConfig{
					{Name: "test", Type: "ollama"},
				},
			},
			wantErr: true,
			errMsg:  "version is required",
		},
		{
			name: "invalid port - zero",
			config: Config{
				Version: "1.0.0",
				Server:  ServerConfig{Port: 0},
				Adapters: []AdapterConfig{
					{Name: "test", Type: "ollama"},
				},
			},
			wantErr: true,
			errMsg:  "invalid server port",
		},
		{
			name: "invalid port - too high",
			config: Config{
				Version: "1.0.0",
				Server:  ServerConfig{Port: 70000},
				Adapters: []AdapterConfig{
					{Name: "test", Type: "ollama"},
				},
			},
			wantErr: true,
			errMsg:  "invalid server port",
		},
		{
			name: "no adapters",
			config: Config{
				Version:  "1.0.0",
				Server:   ServerConfig{Port: 8080},
				Adapters: []AdapterConfig{},
			},
			wantErr: true,
			errMsg:  "at least one adapter is required",
		},
		{
			name: "adapter missing name",
			config: Config{
				Version: "1.0.0",
				Server:  ServerConfig{Port: 8080},
				Adapters: []AdapterConfig{
					{Name: "", Type: "ollama"},
				},
			},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "adapter missing type",
			config: Config{
				Version: "1.0.0",
				Server:  ServerConfig{Port: 8080},
				Adapters: []AdapterConfig{
					{Name: "test", Type: ""},
				},
			},
			wantErr: true,
			errMsg:  "type is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" {
				if err.Error() != tt.errMsg && !contains(err.Error(), tt.errMsg) {
					t.Errorf("Config.Validate() error message = %v, want containing %v", err, tt.errMsg)
				}
			}
		})
	}
}

// TestConfigSaveAndLoad tests saving and loading configuration
func TestConfigSaveAndLoad(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test_config.json")

	// Create test config
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
			"log_level": "info",
		},
	}

	// Save config
	err := originalConfig.Save(configPath)
	if err != nil {
		t.Fatalf("Config.Save() error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("Config file was not created")
	}

	// Load config
	loadedConfig, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	// Verify loaded config matches original
	if loadedConfig.Version != originalConfig.Version {
		t.Errorf("Loaded Version = %s, want %s", loadedConfig.Version, originalConfig.Version)
	}

	if loadedConfig.Server.Port != originalConfig.Server.Port {
		t.Errorf("Loaded Server.Port = %d, want %d", loadedConfig.Server.Port, originalConfig.Server.Port)
	}

	if len(loadedConfig.Adapters) != len(originalConfig.Adapters) {
		t.Errorf("Loaded Adapters count = %d, want %d", len(loadedConfig.Adapters), len(originalConfig.Adapters))
	}
}

// TestConfigToJSON tests JSON conversion
func TestConfigToJSON(t *testing.T) {
	config := &Config{
		Version:  "1.0.0",
		Platform: "linux/amd64",
		Server: ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
		Adapters: []AdapterConfig{
			{Name: "test", Type: "ollama", Enabled: true},
		},
	}

	jsonStr, err := config.ToJSON()
	if err != nil {
		t.Fatalf("Config.ToJSON() error = %v", err)
	}

	if jsonStr == "" {
		t.Error("Config.ToJSON() returned empty string")
	}

	// Verify it's valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		t.Errorf("Config.ToJSON() returned invalid JSON: %v", err)
	}

	// Verify key fields are present
	if _, ok := parsed["version"]; !ok {
		t.Error("JSON missing 'version' field")
	}
	if _, ok := parsed["server"]; !ok {
		t.Error("JSON missing 'server' field")
	}
}

// TestLoadConfigNotExist tests loading non-existent config
func TestLoadConfigNotExist(t *testing.T) {
	_, err := LoadConfig("/nonexistent/path/config.json")
	if err == nil {
		t.Error("LoadConfig() should return error for non-existent file")
	}
}

// TestLoadConfigInvalidJSON tests loading invalid JSON
func TestLoadConfigInvalidJSON(t *testing.T) {
	// Create temporary file with invalid JSON
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.json")

	err := os.WriteFile(configPath, []byte("not valid json"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err = LoadConfig(configPath)
	if err == nil {
		t.Error("LoadConfig() should return error for invalid JSON")
	}
}

// TestGetDefaultConfigPath tests default config path generation
func TestGetDefaultConfigPath(t *testing.T) {
	path := GetDefaultConfigPath()
	if path == "" {
		t.Error("GetDefaultConfigPath() returned empty string")
	}

	// Verify it ends with openclaw.json
	if filepath.Base(path) != "openclaw.json" {
		t.Errorf("GetDefaultConfigPath() = %s, want path ending with openclaw.json", path)
	}
}

// TestConfigSaveCreatesDirectory tests that Save creates parent directories
func TestConfigSaveCreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nested", "deep", "config.json")

	config := &Config{
		Version: "1.0.0",
		Server:  ServerConfig{Port: 8080},
		Adapters: []AdapterConfig{
			{Name: "test", Type: "ollama"},
		},
	}

	err := config.Save(configPath)
	if err != nil {
		t.Fatalf("Config.Save() error = %v", err)
	}

	// Verify directories were created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created in nested directory")
	}
}

// TestConfigWithMultipleAdapters tests config with multiple adapters
func TestConfigWithMultipleAdapters(t *testing.T) {
	config := &Config{
		Version:  "1.0.0",
		Platform: "linux/amd64",
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 8080,
		},
		Adapters: []AdapterConfig{
			{Name: "ollama-local", Type: "ollama", Enabled: true},
			{Name: "openai-api", Type: "openai", Enabled: true},
			{Name: "anthropic-api", Type: "anthropic", Enabled: false},
		},
	}

	if err := config.Validate(); err != nil {
		t.Errorf("Config.Validate() error = %v", err)
	}

	if len(config.Adapters) != 3 {
		t.Errorf("len(Adapters) = %d, want 3", len(config.Adapters))
	}
}

// TestConfigWithSettings tests config with custom settings
func TestConfigWithSettings(t *testing.T) {
	config := &Config{
		Version: "1.0.0",
		Server:  ServerConfig{Port: 8080},
		Adapters: []AdapterConfig{
			{Name: "test", Type: "ollama"},
		},
		Settings: map[string]string{
			"log_level":   "debug",
			"max_tokens":  "4096",
			"timeout":     "30s",
		},
	}

	if err := config.Validate(); err != nil {
		t.Errorf("Config.Validate() error = %v", err)
	}

	if len(config.Settings) != 3 {
		t.Errorf("len(Settings) = %d, want 3", len(config.Settings))
	}
}

// Helper function
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
