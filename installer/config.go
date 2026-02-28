package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents the OpenClaw configuration
type Config struct {
	Version     string            `json:"version"`
	Platform    string            `json:"platform"`
	Server      ServerConfig      `json:"server"`
	Adapters    []AdapterConfig   `json:"adapters"`
	Settings    map[string]string `json:"settings,omitempty"`
}

// ServerConfig contains server-related configuration
type ServerConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
	TLS  bool   `json:"tls"`
}

// AdapterConfig contains adapter configuration
type AdapterConfig struct {
	Name    string            `json:"name"`
	Type    string            `json:"type"`
	Enabled bool              `json:"enabled"`
	Options map[string]string `json:"options,omitempty"`
}

// ConfigOptions contains options for generating configuration
type ConfigOptions struct {
	Version      string
	ServerHost   string
	ServerPort   int
	EnableTLS    bool
	AdapterType  string
	AdapterName  string
	CustomSettings map[string]string
}

// InstallRequest represents the installation request from CLI or API
type InstallRequest struct {
	SourceDir    string            `json:"source_dir"`
	InstallDir   string            `json:"install_dir"`
	ServerHost   string            `json:"server_host"`
	ServerPort   int               `json:"server_port"`
	EnableTLS    bool              `json:"enable_tls"`
	AdapterType  string            `json:"adapter_type"`
	AdapterName  string            `json:"adapter_name"`
	Version      string            `json:"version"`
	Settings     map[string]string `json:"settings,omitempty"`
}

// GenerateConfig creates a new configuration from options
func GenerateConfig(opts ConfigOptions) (*Config, error) {
	platform := DetectPlatform()

	config := &Config{
		Version:  opts.Version,
		Platform: platform.String(),
		Server: ServerConfig{
			Host: opts.ServerHost,
			Port: opts.ServerPort,
			TLS:  opts.EnableTLS,
		},
		Adapters: []AdapterConfig{
			{
				Name:    opts.AdapterName,
				Type:    opts.AdapterType,
				Enabled: true,
				Options: make(map[string]string),
			},
		},
		Settings: opts.CustomSettings,
	}

	return config, nil
}

// Save writes the configuration to a file
func (c *Config) Save(path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Load reads configuration from a file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// ToJSON returns the configuration as a JSON string
func (c *Config) ToJSON() (string, error) {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Version == "" {
		return fmt.Errorf("version is required")
	}

	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	if len(c.Adapters) == 0 {
		return fmt.Errorf("at least one adapter is required")
	}

	for i, adapter := range c.Adapters {
		if adapter.Name == "" {
			return fmt.Errorf("adapter %d: name is required", i)
		}
		if adapter.Type == "" {
			return fmt.Errorf("adapter %d: type is required", i)
		}
	}

	return nil
}

// GetDefaultConfigPath returns the default configuration file path
func GetDefaultConfigPath() string {
	platform := DetectPlatform()
	configDir := platform.GetConfigDir()
	return filepath.Join(configDir, "openclaw.json")
}
