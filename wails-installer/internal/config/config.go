package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config 配置管理器
type Config struct {
	configDir string
}

// OpenClawConfig OpenClaw配置
type OpenClawConfig struct {
	Version     string                 `json:"version"`
	Server      ServerConfig           `json:"server"`
	Adapters    []AdapterConfig        `json:"adapters"`
	LogLevel    string                 `json:"logLevel,omitempty"`
	AutoStart   bool                   `json:"autoStart,omitempty"`
	Theme       string                 `json:"theme,omitempty"`
	CustomVars  map[string]interface{} `json:"customVars,omitempty"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port        int    `json:"port"`
	Host        string `json:"host"`
	TLS         bool   `json:"tls,omitempty"`
	CertFile    string `json:"certFile,omitempty"`
	KeyFile     string `json:"keyFile,omitempty"`
	APIPrefix   string `json:"apiPrefix,omitempty"`
}

// AdapterConfig 适配器配置
type AdapterConfig struct {
	Name    string                 `json:"name"`
	Type    string                 `json:"type"`
	Enable  bool                   `json:"enable"`
	Config  map[string]interface{} `json:"config"`
	Webhook WebhookConfig          `json:"webhook,omitempty"`
}

// WebhookConfig Webhook配置
type WebhookConfig struct {
	URL       string `json:"url,omitempty"`
	Token     string `json:"token,omitempty"`
	Secret    string `json:"secret,omitempty"`
	EncryptKey string `json:"encryptKey,omitempty"`
}

// NewConfig 创建新配置管理器
func NewConfig(configDir string) *Config {
	return &Config{configDir: configDir}
}

// GetDefault 获取默认配置
func (c *Config) GetDefault() OpenClawConfig {
	return OpenClawConfig{
		Version:   "1.0.0",
		LogLevel:  "info",
		AutoStart: true,
		Theme:     "auto",
		Server: ServerConfig{
			Port:      18789,
			Host:      "0.0.0.0",
			TLS:       false,
			APIPrefix: "/api/v1",
		},
		Adapters: []AdapterConfig{
			{
				Name:   "企业微信",
				Type:   "wecom",
				Enable: false,
				Config: map[string]interface{}{
					"corp_id":      "",
					"agent_id":     "",
					"secret":       "",
					"token":        "",
					"aes_key":      "",
					"callback_url": "",
				},
				Webhook: WebhookConfig{
					URL:    "",
					Token:  "",
					Secret: "",
				},
			},
			{
				Name:   "钉钉",
				Type:   "dingtalk",
				Enable: false,
				Config: map[string]interface{}{
					"app_key":      "",
					"app_secret":   "",
					"webhook":      "",
					"robot_code":   "",
				},
				Webhook: WebhookConfig{
					URL:    "",
					Token:  "",
					Secret: "",
				},
			},
			{
				Name:   "飞书",
				Type:   "feishu",
				Enable: false,
				Config: map[string]interface{}{
					"app_id":       "",
					"app_secret":   "",
					"encrypt_key":  "",
					"token":        "",
					"verification_token": "",
				},
				Webhook: WebhookConfig{
					URL:        "",
					Token:      "",
					EncryptKey: "",
				},
			},
			{
				Name:   "Slack",
				Type:   "slack",
				Enable: false,
				Config: map[string]interface{}{
					"bot_token":          "",
					"signing_secret":     "",
					"socket_mode":        false,
					"app_token":          "",
				},
				Webhook: WebhookConfig{
					URL:    "",
					Token:  "",
					Secret: "",
				},
			},
			{
				Name:   "Discord",
				Type:   "discord",
				Enable: false,
				Config: map[string]interface{}{
					"bot_token":      "",
					"client_id":      "",
					"client_secret":  "",
					"guild_id":       "",
				},
				Webhook: WebhookConfig{
					URL:   "",
					Token: "",
				},
			},
		},
		CustomVars: make(map[string]interface{}),
	}
}

// Save 保存配置
func (c *Config) Save(config OpenClawConfig) error {
	if err := os.MkdirAll(c.configDir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	configPath := filepath.Join(c.configDir, "openclaw.json")
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	return nil
}

// Load 加载配置
func (c *Config) Load() (OpenClawConfig, error) {
	configPath := filepath.Join(c.configDir, "openclaw.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// 配置文件不存在，返回默认配置
			return c.GetDefault(), nil
		}
		return OpenClawConfig{}, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var config OpenClawConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return OpenClawConfig{}, fmt.Errorf("解析配置文件失败: %w", err)
	}

	return config, nil
}

// Exists 检查配置文件是否存在
func (c *Config) Exists() bool {
	configPath := filepath.Join(c.configDir, "openclaw.json")
	_, err := os.Stat(configPath)
	return err == nil
}

// Backup 备份配置文件
func (c *Config) Backup() error {
	if !c.Exists() {
		return nil
	}

	sourcePath := filepath.Join(c.configDir, "openclaw.json")
	backupPath := filepath.Join(c.configDir, fmt.Sprintf("openclaw.json.backup.%d", os.Getpid()))

	data, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return fmt.Errorf("写入备份文件失败: %w", err)
	}

	return nil
}

// Validate 验证配置
func (c *Config) Validate(config OpenClawConfig) []string {
	var errors []string

	// 验证服务器配置
	if config.Server.Port < 1 || config.Server.Port > 65535 {
		errors = append(errors, "服务器端口必须在 1-65535 之间")
	}

	if config.Server.Host == "" {
		errors = append(errors, "服务器主机不能为空")
	}

	// 验证适配器配置
	for _, adapter := range config.Adapters {
		if adapter.Enable {
			if adapter.Type == "" {
				errors = append(errors, fmt.Sprintf("适配器 '%s' 类型不能为空", adapter.Name))
			}
		}
	}

	return errors
}

// GetAdapterTypes 获取支持的适配器类型
func GetAdapterTypes() []map[string]string {
	return []map[string]string{
		{
			"id":          "wecom",
			"name":        "企业微信",
			"description": "腾讯企业微信",
			"icon":        "💬",
		},
		{
			"id":          "dingtalk",
			"name":        "钉钉",
			"description": "阿里巴巴钉钉",
			"icon":        "📱",
		},
		{
			"id":          "feishu",
			"name":        "飞书",
			"description": "字节跳动飞书",
			"icon":        "🚀",
		},
		{
			"id":          "slack",
			"name":        "Slack",
			"description": "Slack 团队协作",
			"icon":        "💼",
		},
		{
			"id":          "discord",
			"name":        "Discord",
			"description": "Discord 社区",
			"icon":        "🎮",
		},
	}
}

// MergeWithDefault 将用户配置与默认配置合并
func (c *Config) MergeWithDefault(userConfig OpenClawConfig) OpenClawConfig {
	defaultConfig := c.GetDefault()

	// 如果版本号为空，使用默认版本
	if userConfig.Version == "" {
		userConfig.Version = defaultConfig.Version
	}

	// 如果服务器配置为空，使用默认配置
	if userConfig.Server.Port == 0 {
		userConfig.Server = defaultConfig.Server
	}

	// 如果适配器列表为空，使用默认配置
	if len(userConfig.Adapters) == 0 {
		userConfig.Adapters = defaultConfig.Adapters
	}

	// 合并自定义变量
	if userConfig.CustomVars == nil {
		userConfig.CustomVars = make(map[string]interface{})
	}

	return userConfig
}
