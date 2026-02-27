// Package types 定义更新程序使用的共享类型
package types

import (
	"time"
)

// VersionInfo 表示组件版本信息
type VersionInfo struct {
	Version        string    `json:"version"`
	DownloadURL    string    `json:"download_url"`
	Checksum       string    `json:"checksum"`
	ChecksumType   string    `json:"checksum_type,omitempty"`
	ReleaseDate    time.Time `json:"release_date"`
	ReleaseNotes   string    `json:"release_notes,omitempty"`
	MinCoreVersion string    `json:"min_core_version,omitempty"`
}

// ComponentType 表示组件类型
type ComponentType string

const (
	ComponentCore     ComponentType = "core"
	ComponentWecom    ComponentType = "wecom-adapter"
	ComponentDingtalk ComponentType = "dingtalk-adapter"
	ComponentFeishu   ComponentType = "feishu-adapter"
)

// ReleaseManifest 表示发布清单
type ReleaseManifest struct {
	Core     VersionInfo            `json:"core"`
	Adapters map[string]VersionInfo `json:"adapters"`
}

// InstalledComponent 表示已安装的组件
type InstalledComponent struct {
	Type       ComponentType `json:"type"`
	Name       string        `json:"name"`
	Version    string        `json:"version"`
	InstallPath string       `json:"install_path"`
	ConfigPath  string       `json:"config_path,omitempty"`
	InstallDate time.Time    `json:"install_date"`
}

// InstallationRecord 表示安装记录
type InstallationRecord struct {
	Version     string                `json:"version"`
	InstallDate time.Time             `json:"install_date"`
	Components  []InstalledComponent  `json:"components"`
}

// UpdateInfo 表示更新信息
type UpdateInfo struct {
	Component   ComponentType `json:"component"`
	Name        string        `json:"name"`
	CurrentVer  string        `json:"current_version"`
	AvailableVer string       `json:"available_version"`
	DownloadURL string        `json:"download_url"`
	Checksum    string        `json:"checksum"`
	ReleaseNotes string       `json:"release_notes,omitempty"`
}

// BackupInfo 表示备份信息
type BackupInfo struct {
	ID          string    `json:"id"`
	Timestamp   time.Time `json:"timestamp"`
	Version     string    `json:"version"`
	Components  []string  `json:"components"`
	BackupPath  string    `json:"backup_path"`
	Size        int64     `json:"size"`
}

// BackupManifest 表示备份清单
type BackupManifest struct {
	BackupID    string                `json:"backup_id"`
	Timestamp   time.Time             `json:"timestamp"`
	Version     string                `json:"version"`
	Components  []BackupComponent     `json:"components"`
}

// BackupComponent 表示备份的组件
type BackupComponent struct {
	Type        ComponentType `json:"type"`
	Name        string        `json:"name"`
	Version     string        `json:"version"`
	Files       []BackupFile  `json:"files"`
}

// BackupFile 表示备份的文件
type BackupFile struct {
	SourcePath string `json:"source_path"`
	BackupPath string `json:"backup_path"`
	Hash       string `json:"hash"`
	Mode       uint32 `json:"mode"`
}

// Config 表示更新程序配置
type Config struct {
	Update struct {
		CheckInterval string `json:"check_interval"`
		AutoUpdate    bool   `json:"auto_update"`
		Channel       string `json:"channel"`
	} `json:"update"`

	Source struct {
		Type       string `json:"type"`
		URL        string `json:"url"`
		LocalPath  string `json:"local_path"`
	} `json:"source"`

	Components struct {
		Core     bool     `json:"core"`
		Adapters []string `json:"adapters"`
	} `json:"components"`

	Backup struct {
		KeepCount  int    `json:"keep_count"`
		BackupDir  string `json:"backup_dir"`
	} `json:"backup"`

	Proxy struct {
		Enabled    bool   `json:"enabled"`
		HTTPProxy  string `json:"http_proxy"`
		HTTPSProxy string `json:"https_proxy"`
	} `json:"proxy"`
}

// UpdateResult 表示更新结果
type UpdateResult struct {
	Success     bool          `json:"success"`
	Component   ComponentType `json:"component"`
	FromVersion string        `json:"from_version"`
	ToVersion   string        `json:"to_version"`
	Error       string        `json:"error,omitempty"`
	Timestamp   time.Time     `json:"timestamp"`
}

// UpdateSummary 表示更新摘要
type UpdateSummary struct {
	Total       int            `json:"total"`
	Successful  int            `json:"successful"`
	Failed      int            `json:"failed"`
	Results     []UpdateResult `json:"results"`
	StartTime   time.Time      `json:"start_time"`
	EndTime     time.Time      `json:"end_time"`
}

// PlatformInfo 表示平台信息
type PlatformInfo struct {
	OS   string `json:"os"`
	Arch string `json:"arch"`
}

// DownloadProgress 表示下载进度
type DownloadProgress struct {
	Component   ComponentType `json:"component"`
	URL         string        `json:"url"`
	TotalBytes  int64         `json:"total_bytes"`
	Downloaded  int64         `json:"downloaded"`
	Percentage  float64       `json:"percentage"`
	Speed       float64       `json:"speed"` // bytes per second
}

// UpdateOptions 表示更新选项
type UpdateOptions struct {
	Components   []ComponentType
	SourcePath   string
	DryRun       bool
	Force        bool
	SkipBackup   bool
	SkipVerify   bool
}

// CheckOptions 表示检查选项
type CheckOptions struct {
	Components []ComponentType
}

// RollbackOptions 表示回滚选项
type RollbackOptions struct {
	BackupID string
}
