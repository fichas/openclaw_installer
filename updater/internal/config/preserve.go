// Package config 提供配置保留功能
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Preserver 提供配置保留功能
type Preserver struct {
	configPaths map[string]string
}

// NewPreserver 创建新的配置保留器
func NewPreserver(configPaths map[string]string) *Preserver {
	return &Preserver{
		configPaths: configPaths,
	}
}

// PreserveAndMerge 保留用户配置并合并新配置
func (p *Preserver) PreserveAndMerge(componentName, newConfigDir string) error {
	userConfigPath := p.configPaths[componentName]
	if userConfigPath == "" {
		// 没有现有配置，直接使用新配置
		return nil
	}

	// 检查用户配置是否存在
	if _, err := os.Stat(userConfigPath); os.IsNotExist(err) {
		// 用户配置不存在，直接使用新配置
		return nil
	}

	// 读取用户配置
	userConfig, err := p.loadConfig(userConfigPath)
	if err != nil {
		return fmt.Errorf("failed to load user config: %w", err)
	}

	// 读取新配置（作为模板）
	newConfigPath := filepath.Join(newConfigDir, componentName+".yaml")
	newConfig, err := p.loadConfig(newConfigPath)
	if err != nil {
		// 新配置可能不存在，使用用户配置
		return nil
	}

	// 合并配置：用户配置优先，新配置中的新字段使用默认值
	mergedConfig := p.mergeConfigs(userConfig, newConfig)

	// 写回合并后的配置
	if err := p.saveConfig(userConfigPath, mergedConfig); err != nil {
		return fmt.Errorf("failed to save merged config: %w", err)
	}

	return nil
}

// loadConfig 加载 YAML 配置文件
func (p *Preserver) loadConfig(path string) (map[string]interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config map[string]interface{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return config, nil
}

// saveConfig 保存 YAML 配置文件
func (p *Preserver) saveConfig(path string, config map[string]interface{}) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	// 保留原文件权限
	var mode os.FileMode = 0640
	if info, err := os.Stat(path); err == nil {
		mode = info.Mode()
	}

	return os.WriteFile(path, data, mode)
}

// mergeConfigs 合并配置（递归）
func (p *Preserver) mergeConfigs(userConfig, newConfig map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// 首先复制所有新配置的键（获取新字段的默认值）
	for key, value := range newConfig {
		result[key] = value
	}

	// 然后用用户配置覆盖（保留用户设置）
	for key, userValue := range userConfig {
		if newValue, exists := newConfig[key]; exists {
			// 键存在于新旧配置中，递归合并
			if userMap, ok := userValue.(map[string]interface{}); ok {
				if newMap, ok := newValue.(map[string]interface{}); ok {
					result[key] = p.mergeConfigs(userMap, newMap)
					continue
				}
			}
		}
		// 直接使用用户值
		result[key] = userValue
	}

	return result
}

// BackupConfig 备份配置文件
func (p *Preserver) BackupConfig(componentName, backupDir string) (string, error) {
	configPath := p.configPaths[componentName]
	if configPath == "" {
		return "", nil
	}

	// 检查配置是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return "", nil
	}

	// 创建备份目录
	if err := os.MkdirAll(backupDir, 0750); err != nil {
		return "", err
	}

	// 复制配置文件
	backupPath := filepath.Join(backupDir, filepath.Base(configPath))
	if err := p.copyFile(configPath, backupPath); err != nil {
		return "", err
	}

	return backupPath, nil
}

// copyFile 复制文件
func (p *Preserver) copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	// 保留原文件权限
	info, err := os.Stat(src)
	var mode os.FileMode = 0640
	if err == nil {
		mode = info.Mode()
	}

	return os.WriteFile(dst, data, mode)
}

// GetConfigPaths 获取所有配置文件路径
func (p *Preserver) GetConfigPaths() map[string]string {
	return p.configPaths
}

// ValidateConfig 验证配置文件有效性
func (p *Preserver) ValidateConfig(configPath string) error {
	config, err := p.loadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// 基本验证：检查关键字段
	if len(config) == 0 {
		return fmt.Errorf("config file is empty")
	}

	return nil
}

// MigrateConfig 迁移旧版本配置到新版本
func (p *Preserver) MigrateConfig(configPath string, fromVersion, toVersion string) error {
	// 读取配置
	config, err := p.loadConfig(configPath)
	if err != nil {
		return err
	}

	// 版本特定的迁移逻辑
	// 这里可以添加针对不同版本差异的迁移代码

	// 更新版本号
	config["version"] = toVersion

	// 保存
	return p.saveConfig(configPath, config)
}

// ListConfigFiles 列出所有配置文件
func (p *Preserver) ListConfigFiles() ([]string, error) {
	var files []string
	for _, path := range p.configPaths {
		if path == "" {
			continue
		}
		// 查找所有 YAML 文件
		matches, err := filepath.Glob(filepath.Join(path, "*.yaml"))
		if err != nil {
			continue
		}
		files = append(files, matches...)

		matches, err = filepath.Glob(filepath.Join(path, "*.yml"))
		if err != nil {
			continue
		}
		files = append(files, matches...)
	}
	return files, nil
}

// IsProtectedKey 检查配置键是否受保护（不应被覆盖）
func (p *Preserver) IsProtectedKey(key string) bool {
	protectedKeys := []string{
		"webhook_url",
		"corp_id",
		"corp_secret",
		"app_key",
		"app_secret",
		"token",
		"secret",
		"password",
		"api_key",
		"private_key",
	}

	lowerKey := strings.ToLower(key)
	for _, protected := range protectedKeys {
		if strings.Contains(lowerKey, protected) {
			return true
		}
	}
	return false
}
