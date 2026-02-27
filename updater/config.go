package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// UpdateConfig represents the updater configuration
type UpdateConfig struct {
	Version        string            `json:"version"`
	AutoUpdate     bool              `json:"auto_update"`
	CheckInterval  string            `json:"check_interval"` // cron expression or duration
	MaxBackups     int               `json:"max_backups"`
	BackupBefore   bool              `json:"backup_before"`
	PreserveConfig bool              `json:"preserve_config"`
	Channels       UpdateChannels    `json:"channels"`
	Adapters       AdapterConfig     `json:"adapters"`
	Notifications  NotificationConfig `json:"notifications"`
	LastCheck      time.Time         `json:"last_check"`
	LastUpdate     time.Time         `json:"last_update"`
	CurrentVersion string            `json:"current_version"`
	CustomSettings map[string]string `json:"custom_settings,omitempty"`
}

// UpdateChannels defines update channels for different components
type UpdateChannels struct {
	Core    string `json:"core"`    // stable, beta, nightly
	Wecom   string `json:"wecom"`   // stable, beta, nightly
	Dingtalk string `json:"dingtalk"` // stable, beta, nightly
	Feishu  string `json:"feishu"`  // stable, beta, nightly
}

// AdapterConfig contains adapter-specific update settings
type AdapterConfig struct {
	AutoUpdate []string `json:"auto_update"` // List of adapters to auto-update
	PinVersions map[string]string `json:"pin_versions"` // Adapter -> pinned version
}

// NotificationConfig contains notification settings
type NotificationConfig struct {
	Enabled     bool     `json:"enabled"`
	OnSuccess   bool     `json:"on_success"`
	OnFailure   bool     `json:"on_failure"`
	Email       string   `json:"email,omitempty"`
	WebhookURL  string   `json:"webhook_url,omitempty"`
}

// DefaultUpdateConfig returns the default update configuration
func DefaultUpdateConfig() *UpdateConfig {
	return &UpdateConfig{
		Version:        "1.0.0",
		AutoUpdate:     false,
		CheckInterval:  "0 2 * * *", // Daily at 2 AM
		MaxBackups:     5,
		BackupBefore:   true,
		PreserveConfig: true,
		Channels: UpdateChannels{
			Core:     "stable",
			Wecom:    "stable",
			Dingtalk: "stable",
			Feishu:   "stable",
		},
		Adapters: AdapterConfig{
			AutoUpdate:  []string{}, // Empty means none auto-update
			PinVersions: make(map[string]string),
		},
		Notifications: NotificationConfig{
			Enabled:   true,
			OnSuccess: false,
			OnFailure: true,
		},
		CustomSettings: make(map[string]string),
	}
}

// LoadUpdateConfig loads the update configuration from a file
func LoadUpdateConfig(path string) (*UpdateConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default config if file doesn't exist
			return DefaultUpdateConfig(), nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config UpdateConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set defaults for missing values
	if config.MaxBackups == 0 {
		config.MaxBackups = 5
	}
	if config.CheckInterval == "" {
		config.CheckInterval = "0 2 * * *"
	}
	if config.Channels.Core == "" {
		config.Channels.Core = "stable"
	}

	return &config, nil
}

// Save writes the configuration to a file
func (c *UpdateConfig) Save(path string) error {
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

// ToJSON returns the configuration as a JSON string
func (c *UpdateConfig) ToJSON() (string, error) {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Validate checks if the configuration is valid
func (c *UpdateConfig) Validate() error {
	if c.MaxBackups < 0 {
		return fmt.Errorf("max_backups must be non-negative")
	}

	validChannels := map[string]bool{"stable": true, "beta": true, "nightly": true}
	if !validChannels[c.Channels.Core] {
		return fmt.Errorf("invalid core channel: %s", c.Channels.Core)
	}

	return nil
}

// ShouldAutoUpdateAdapter checks if an adapter should be auto-updated
func (c *UpdateConfig) ShouldAutoUpdateAdapter(name string) bool {
	for _, adapter := range c.Adapters.AutoUpdate {
		if adapter == name || adapter == "all" {
			// Check if version is pinned
			if _, pinned := c.Adapters.PinVersions[name]; pinned {
				return false
			}
			return true
		}
	}
	return false
}

// GetAdapterChannel returns the update channel for an adapter
func (c *UpdateConfig) GetAdapterChannel(name string) string {
	switch name {
	case "wecom":
		return c.Channels.Wecom
	case "dingtalk":
		return c.Channels.Dingtalk
	case "feishu":
		return c.Channels.Feishu
	default:
		return "stable"
	}
}

// UpdateLastCheck updates the last check timestamp
func (c *UpdateConfig) UpdateLastCheck() {
	c.LastCheck = time.Now()
}

// UpdateLastUpdate updates the last update timestamp
func (c *UpdateConfig) UpdateLastUpdate() {
	c.LastUpdate = time.Now()
}

// IsUpdateDue checks if an update check is due based on the check interval
func (c *UpdateConfig) IsUpdateDue() bool {
	if c.LastCheck.IsZero() {
		return true
	}

	// Parse check interval as duration (simplified)
	// For cron expressions, this would need a cron parser
	duration, err := time.ParseDuration(c.CheckInterval)
	if err != nil {
		// Default to 24 hours if parsing fails
		duration = 24 * time.Hour
	}

	return time.Since(c.LastCheck) >= duration
}
