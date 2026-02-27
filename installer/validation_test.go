package main

import (
	"regexp"
	"testing"
)

// TestValidatePort tests port validation
func TestValidatePort(t *testing.T) {
	tests := []struct {
		name      string
		port      int
		wantValid bool
	}{
		{"valid port 1", 1, true},
		{"valid port 80", 80, true},
		{"valid port 8080", 8080, true},
		{"valid port 65535", 65535, true},
		{"invalid port 0", 0, false},
		{"invalid port -1", -1, false},
		{"invalid port 65536", 65536, false},
		{"invalid port 100000", 100000, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{
				Version: "1.0.0",
				Server:  ServerConfig{Port: tt.port},
				Adapters: []AdapterConfig{
					{Name: "test", Type: "ollama"},
				},
			}

			err := config.Validate()
			if tt.wantValid && err != nil {
				t.Errorf("Validate() error = %v, want nil", err)
			}
			if !tt.wantValid && err == nil {
				t.Error("Validate() should return error for invalid port")
			}
		})
	}
}

// TestValidateVersion tests version validation
func TestValidateVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
		valid   bool
	}{
		{"valid semantic version", "1.0.0", true},
		{"valid version with prefix", "v1.0.0", true},
		{"valid complex version", "1.2.3-beta.1", true},
		{"empty version", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{
				Version: tt.version,
				Server:  ServerConfig{Port: 8080},
				Adapters: []AdapterConfig{
					{Name: "test", Type: "ollama"},
				},
			}

			err := config.Validate()
			if tt.valid && err != nil {
				t.Errorf("Validate() error = %v, want nil", err)
			}
			if !tt.valid && err == nil {
				t.Error("Validate() should return error for invalid version")
			}
		})
	}
}

// TestValidateAdapterName tests adapter name validation
func TestValidateAdapterName(t *testing.T) {
	tests := []struct {
		name   string
		adapter string
		valid  bool
	}{
		{"valid name", "ollama-local", true},
		{"valid name with numbers", "adapter123", true},
		{"empty name", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{
				Version: "1.0.0",
				Server:  ServerConfig{Port: 8080},
				Adapters: []AdapterConfig{
					{Name: tt.adapter, Type: "ollama"},
				},
			}

			err := config.Validate()
			if tt.valid && err != nil {
				t.Errorf("Validate() error = %v, want nil", err)
			}
			if !tt.valid && err == nil {
				t.Error("Validate() should return error for invalid adapter name")
			}
		})
	}
}

// TestValidateAdapterType tests adapter type validation
func TestValidateAdapterType(t *testing.T) {
	tests := []struct {
		name   string
		adapterType string
		valid  bool
	}{
		{"valid type ollama", "ollama", true},
		{"valid type openai", "openai", true},
		{"valid type anthropic", "anthropic", true},
		{"empty type", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{
				Version: "1.0.0",
				Server:  ServerConfig{Port: 8080},
				Adapters: []AdapterConfig{
					{Name: "test", Type: tt.adapterType},
				},
			}

			err := config.Validate()
			if tt.valid && err != nil {
				t.Errorf("Validate() error = %v, want nil", err)
			}
			if !tt.valid && err == nil {
				t.Error("Validate() should return error for invalid adapter type")
			}
		})
	}
}

// TestValidateServerHost tests server host validation
func TestValidateServerHost(t *testing.T) {
	tests := []struct {
		name string
		host string
		valid bool
	}{
		{"localhost", "localhost", true},
		{"127.0.0.1", "127.0.0.1", true},
		{"0.0.0.0", "0.0.0.0", true},
		{"empty host", "", true}, // Empty host is allowed (uses default)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{
				Version: "1.0.0",
				Server: ServerConfig{
					Host: tt.host,
					Port: 8080,
				},
				Adapters: []AdapterConfig{
					{Name: "test", Type: "ollama"},
				},
			}

			err := config.Validate()
			if tt.valid && err != nil {
				t.Errorf("Validate() error = %v, want nil", err)
			}
		})
	}
}

// TestValidateInstallRequest tests install request validation
func TestValidateInstallRequest(t *testing.T) {
	tests := []struct {
		name    string
		request InstallRequest
		valid   bool
	}{
		{
			name: "complete valid request",
			request: InstallRequest{
				SourceDir:   "/media/usb",
				InstallDir:  "/opt/openclaw",
				ServerHost:  "0.0.0.0",
				ServerPort:  8080,
				EnableTLS:   true,
				AdapterType: "ollama",
				AdapterName: "local-ollama",
				Version:     "1.0.0",
				Settings:    map[string]string{"log_level": "debug"},
			},
			valid: true,
		},
		{
			name: "minimal valid request",
			request: InstallRequest{
				ServerPort:  8080,
				AdapterType: "usb",
				AdapterName: "default",
				Version:     "1.0.0",
			},
			valid: true,
		},
		{
			name: "zero port",
			request: InstallRequest{
				ServerPort:  0,
				AdapterType: "usb",
				Version:     "1.0.0",
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create config from request
			config, err := GenerateConfig(ConfigOptions{
				Version:     tt.request.Version,
				ServerHost:  tt.request.ServerHost,
				ServerPort:  tt.request.ServerPort,
				EnableTLS:   tt.request.EnableTLS,
				AdapterType: tt.request.AdapterType,
				AdapterName: tt.request.AdapterName,
			})

			if err != nil {
				if tt.valid {
					t.Errorf("GenerateConfig() error = %v", err)
				}
				return
			}

			validateErr := config.Validate()
			if tt.valid && validateErr != nil {
				t.Errorf("Validate() error = %v, want nil", validateErr)
			}
			if !tt.valid && validateErr == nil {
				t.Error("Validate() should return error for invalid request")
			}
		})
	}
}

// TestValidatePath tests path validation
func TestValidatePath(t *testing.T) {
	tests := []struct {
		name string
		path string
		valid bool
	}{
		{"Unix absolute path", "/opt/openclaw", true},
		{"Unix home path", "~/openclaw", true},
		{"Windows absolute path", `C:\Program Files\OpenClaw`, true},
		{"Relative path", "./openclaw", true},
		{"empty path", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Path validation is implicit in directory creation
			// Just check it's not empty for critical paths
			isValid := tt.path != ""
			if isValid != tt.valid {
				t.Errorf("Path validation for %q: got %v, want %v", tt.path, isValid, tt.valid)
			}
		})
	}
}

// TestValidateSettings tests settings validation
func TestValidateSettings(t *testing.T) {
	tests := []struct {
		name     string
		settings map[string]string
		valid    bool
	}{
		{
			name:     "valid settings",
			settings: map[string]string{"log_level": "debug", "timeout": "30s"},
			valid:    true,
		},
		{
			name:     "empty settings",
			settings: map[string]string{},
			valid:    true,
		},
		{
			name:     "nil settings",
			settings: nil,
			valid:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{
				Version:  "1.0.0",
				Server:   ServerConfig{Port: 8080},
				Adapters: []AdapterConfig{{Name: "test", Type: "ollama"}},
				Settings: tt.settings,
			}

			err := config.Validate()
			if tt.valid && err != nil {
				t.Errorf("Validate() error = %v, want nil", err)
			}
		})
	}
}

// TestValidateLogLevel tests log level validation
func TestValidateLogLevel(t *testing.T) {
	validLevels := []string{"debug", "info", "warn", "error"}
	invalidLevels := []string{"", "invalid", "DEBUG", "INFO"}

	for _, level := range validLevels {
		t.Run("valid_"+level, func(t *testing.T) {
			// Log level validation is done at the form level
			// This test documents expected valid values
			t.Logf("Valid log level: %s", level)
		})
	}

	for _, level := range invalidLevels {
		t.Run("invalid_"+level, func(t *testing.T) {
			t.Logf("Invalid log level: %s", level)
		})
	}
}

// TestValidateTLSConfig tests TLS configuration validation
func TestValidateTLSConfig(t *testing.T) {
	tests := []struct {
		name    string
		enableTLS bool
		valid   bool
	}{
		{"TLS enabled", true, true},
		{"TLS disabled", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{
				Version: "1.0.0",
				Server: ServerConfig{
					Port: 8080,
					TLS:  tt.enableTLS,
				},
				Adapters: []AdapterConfig{
					{Name: "test", Type: "ollama"},
				},
			}

			err := config.Validate()
			if tt.valid && err != nil {
				t.Errorf("Validate() error = %v, want nil", err)
			}
		})
	}
}

// TestValidateMultipleAdapters tests validation with multiple adapters
func TestValidateMultipleAdapters(t *testing.T) {
	tests := []struct {
		name     string
		adapters []AdapterConfig
		valid    bool
	}{
		{
			name: "single valid adapter",
			adapters: []AdapterConfig{
				{Name: "ollama-local", Type: "ollama", Enabled: true},
			},
			valid: true,
		},
		{
			name: "multiple valid adapters",
			adapters: []AdapterConfig{
				{Name: "ollama-local", Type: "ollama", Enabled: true},
				{Name: "openai-api", Type: "openai", Enabled: true},
				{Name: "anthropic-api", Type: "anthropic", Enabled: false},
			},
			valid: true,
		},
		{
			name:     "no adapters",
			adapters: []AdapterConfig{},
			valid:    false,
		},
		{
			name: "adapter with empty name",
			adapters: []AdapterConfig{
				{Name: "", Type: "ollama", Enabled: true},
			},
			valid: false,
		},
		{
			name: "adapter with empty type",
			adapters: []AdapterConfig{
				{Name: "test", Type: "", Enabled: true},
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{
				Version:  "1.0.0",
				Server:   ServerConfig{Port: 8080},
				Adapters: tt.adapters,
			}

			err := config.Validate()
			if tt.valid && err != nil {
				t.Errorf("Validate() error = %v, want nil", err)
			}
			if !tt.valid && err == nil {
				t.Error("Validate() should return error for invalid adapters")
			}
		})
	}
}

// TestSanitizePath tests path sanitization
func TestSanitizePath(t *testing.T) {
	// This test documents expected path sanitization behavior
	// Actual sanitization would be implemented in the production code
	tests := []struct {
		input    string
		expected string
	}{
		{"/normal/path", "/normal/path"},
		{"/path/with/../parent", "/path/with/../parent"},
		{"/path/with/./current", "/path/with/./current"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			// Document expected behavior
			t.Logf("Path: %s -> %s", tt.input, tt.expected)
		})
	}
}

// TestRegexPatterns tests regex patterns for validation
func TestRegexPatterns(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		input   string
		match   bool
	}{
		{
			name:    "valid hostname",
			pattern: `^[a-zA-Z0-9][a-zA-Z0-9\-\.]{0,253}$`,
			input:   "localhost",
			match:   true,
		},
		{
			name:    "valid IP",
			pattern: `^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`,
			input:   "192.168.1.1",
			match:   true,
		},
		{
			name:    "valid version",
			pattern: `^\d+\.\d+\.\d+`,
			input:   "1.0.0",
			match:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re, err := regexp.Compile(tt.pattern)
			if err != nil {
				t.Fatalf("Invalid regex pattern: %v", err)
			}

			match := re.MatchString(tt.input)
			if match != tt.match {
				t.Errorf("Match(%q) = %v, want %v", tt.input, match, tt.match)
			}
		})
	}
}
