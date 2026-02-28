// OpenClaw Updater - 定期更新程序
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/openclaw/updater/internal/backup"
	"github.com/openclaw/updater/internal/config"
	"github.com/openclaw/updater/internal/download"
	"github.com/openclaw/updater/internal/install"
	"github.com/openclaw/updater/internal/rollback"
	"github.com/openclaw/updater/internal/version"
	"github.com/openclaw/updater/pkg/types"
	"gopkg.in/yaml.v3"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

// Config 更新程序配置
type Config struct {
	Update struct {
		CheckInterval string `json:"check_interval"`
		AutoUpdate    bool   `json:"auto_update"`
		Channel       string `json:"channel"`
	} `json:"update"`

	Source struct {
		Type      string `json:"type"`
		URL       string `json:"url"`
		LocalPath string `json:"local_path"`
	} `json:"source"`

	Components struct {
		Core     bool     `json:"core"`
		Adapters []string `json:"adapters"`
	} `json:"components"`

	Backup struct {
		KeepCount int    `json:"keep_count"`
		BackupDir string `json:"backup_dir"`
	} `json:"backup"`

	Paths struct {
		InstallDir string `json:"install_dir"`
		ConfigDir  string `json:"config_dir"`
	} `json:"paths"`
}

func main() {
	var (
		configPath   = flag.String("config", "/etc/openclaw/updater.json", "配置文件路径")
		checkOnly    = flag.Bool("check", false, "仅检查更新")
		component    = flag.String("component", "", "指定要更新的组件 (core, wecom, dingtalk, feishu)")
		sourcePath   = flag.String("source", "", "本地更新源路径")
		dryRun       = flag.Bool("dry-run", false, "模拟运行，不实际执行更新")
		force        = flag.Bool("force", false, "强制更新，即使版本相同")
		skipBackup   = flag.Bool("skip-backup", false, "跳过备份")
		rollbackFlag = flag.Bool("rollback", false, "回滚到上一版本")
		versionFlag  = flag.Bool("version", false, "显示版本信息")
	)
	flag.Parse()

	if *versionFlag {
		printVersion()
		return
	}

	// 加载配置
	cfg, err := loadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// 命令分发
	switch {
	case *rollbackFlag:
		if err := doRollback(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Rollback failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Rollback completed successfully")

	case *checkOnly:
		if err := doCheck(cfg, *sourcePath); err != nil {
			fmt.Fprintf(os.Stderr, "Check failed: %v\n", err)
			os.Exit(1)
		}

	default:
		options := types.UpdateOptions{
			DryRun:     *dryRun,
			Force:      *force,
			SkipBackup: *skipBackup,
			SourcePath: *sourcePath,
		}

		if *component != "" {
			options.Components = []types.ComponentType{types.ComponentType(*component)}
		}

		if err := doUpdate(cfg, options); err != nil {
			fmt.Fprintf(os.Stderr, "Update failed: %v\n", err)
			os.Exit(1)
		}
	}
}

func printVersion() {
	fmt.Printf("OpenClaw Updater %s\n", Version)
	fmt.Printf("  Build Time: %s\n", BuildTime)
	fmt.Printf("  Git Commit: %s\n", GitCommit)
	fmt.Printf("  Go Version: %s\n", runtime.Version())
	fmt.Printf("  OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
}

func loadConfig(path string) (*Config, error) {
	// 默认配置
	cfg := &Config{}
	cfg.Update.CheckInterval = "7d"
	cfg.Update.Channel = "stable"
	cfg.Source.Type = "remote"
	cfg.Source.URL = "https://releases.openclaw.io/manifest.json"
	cfg.Components.Core = true
	cfg.Components.Adapters = []string{"wecom", "dingtalk", "feishu"}
	cfg.Backup.KeepCount = 3

	// 设置默认路径
	switch runtime.GOOS {
	case "darwin":
		cfg.Backup.BackupDir = "/usr/local/var/backups/openclaw"
		cfg.Paths.InstallDir = "/usr/local/bin"
		cfg.Paths.ConfigDir = "/usr/local/etc/openclaw"
	case "windows":
		cfg.Backup.BackupDir = filepath.Join(os.Getenv("ProgramData"), "OpenClaw", "backups")
		cfg.Paths.InstallDir = filepath.Join(os.Getenv("ProgramFiles"), "OpenClaw")
		cfg.Paths.ConfigDir = filepath.Join(os.Getenv("ProgramData"), "OpenClaw", "config")
	default: // linux
		cfg.Backup.BackupDir = "/var/backups/openclaw"
		cfg.Paths.InstallDir = "/usr/local/bin"
		cfg.Paths.ConfigDir = "/etc/openclaw"
	}

	// 尝试读取配置文件（如果存在）
	if data, err := os.ReadFile(path); err == nil {
		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".yaml", ".yml":
			if err := yaml.Unmarshal(data, cfg); err != nil {
				return nil, fmt.Errorf("failed to parse yaml config: %w", err)
			}
		default:
			if err := json.Unmarshal(data, cfg); err != nil {
				// 为兼容历史 YAML 配置，JSON 失败后尝试 YAML。
				if yamlErr := yaml.Unmarshal(data, cfg); yamlErr != nil {
					return nil, fmt.Errorf("failed to parse config as json(%v) or yaml(%v)", err, yamlErr)
				}
			}
		}
	}

	return cfg, nil
}

func doCheck(cfg *Config, sourcePath string) error {
	fmt.Println("Checking for updates...")

	checker := version.NewChecker(cfg.Source.URL)

	var manifest *types.ReleaseManifest
	var err error

	if sourcePath != "" || cfg.Source.Type == "local" {
		localPath := sourcePath
		if localPath == "" {
			localPath = cfg.Source.LocalPath
		}
		manifest, err = checker.FetchLocalManifest(localPath)
	} else {
		manifest, err = checker.FetchRemoteManifest()
	}

	if err != nil {
		return fmt.Errorf("failed to fetch manifest: %w", err)
	}

	// 获取已安装版本
	installed := getInstalledVersions(cfg)

	// 检查更新
	updates, err := checker.CheckForUpdates(manifest, installed)
	if err != nil {
		return fmt.Errorf("failed to check updates: %w", err)
	}

	if len(updates) == 0 {
		fmt.Println("No updates available. All components are up to date.")
		return nil
	}

	fmt.Printf("Found %d update(s):\n\n", len(updates))
	for _, u := range updates {
		fmt.Printf("  %s: %s -> %s\n", u.Name, u.CurrentVer, u.AvailableVer)
		if u.ReleaseNotes != "" {
			fmt.Printf("    Release notes: %s\n", u.ReleaseNotes)
		}
		fmt.Println()
	}

	return nil
}

func doUpdate(cfg *Config, options types.UpdateOptions) error {
	fmt.Println("Starting update process...")

	// 1. 获取发布清单
	checker := version.NewChecker(cfg.Source.URL)

	var manifest *types.ReleaseManifest
	var err error

	if options.SourcePath != "" {
		manifest, err = checker.FetchLocalManifest(options.SourcePath)
	} else {
		manifest, err = checker.FetchRemoteManifest()
	}

	if err != nil {
		return fmt.Errorf("failed to fetch manifest: %w", err)
	}

	// 2. 获取已安装版本
	installed := getInstalledVersions(cfg)

	// 3. 检查更新
	updates, err := checker.CheckForUpdates(manifest, installed)
	if err != nil {
		return fmt.Errorf("failed to check updates: %w", err)
	}

	// 过滤指定组件
	if len(options.Components) > 0 {
		updates = filterUpdates(updates, options.Components)
	}

	if len(updates) == 0 {
		fmt.Println("No updates available.")
		return nil
	}

	fmt.Printf("Will update %d component(s):\n", len(updates))
	for _, u := range updates {
		fmt.Printf("  - %s: %s -> %s\n", u.Name, u.CurrentVer, u.AvailableVer)
	}

	if options.DryRun {
		fmt.Println("\nDry run mode - no changes will be made.")
		return nil
	}

	// 4. 创建临时目录
	tempDir, err := os.MkdirTemp("", "openclaw-update-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// 5. 下载更新包
	fmt.Println("\nDownloading updates...")
	downloader := download.NewDownloader()
	downloaded, err := downloader.DownloadMultiple(updates, tempDir)
	if err != nil {
		return fmt.Errorf("failed to download updates: %w", err)
	}

	// 6. 备份当前版本
	if !options.SkipBackup {
		fmt.Println("Creating backup...")
		backupMgr := backup.NewManager(cfg.Backup.BackupDir, cfg.Backup.KeepCount)
		if err := backupMgr.Init(); err != nil {
			return fmt.Errorf("failed to initialize backup: %w", err)
		}

		installedComponents := getInstalledComponents(cfg)
		_, err = backupMgr.CreateBackup(installedComponents, installed[types.ComponentCore])
		if err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
	}

	// 7. 安装更新
	fmt.Println("Installing updates...")
	installPaths := getInstallPaths(cfg)
	configPaths := getConfigPaths(cfg)
	installer := install.NewInstaller(installPaths, configPaths)
	configPreserver := config.NewPreserver(configPaths)

	for component, packagePath := range downloaded {
		fmt.Printf("  Installing %s...\n", component)

		if err := installer.InstallComponent(component, packagePath, options); err != nil {
			// 安装失败，尝试回滚
			fmt.Fprintf(os.Stderr, "Failed to install %s: %v\n", component, err)
			fmt.Println("Attempting rollback...")
			if rbErr := doRollback(cfg); rbErr != nil {
				fmt.Fprintf(os.Stderr, "Rollback failed: %v\n", rbErr)
			}
			return fmt.Errorf("installation failed for %s", component)
		}

		// 保留配置
		if err := configPreserver.PreserveAndMerge(component, packagePath); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to preserve config for %s: %v\n", component, err)
		}
	}

	// 8. 验证安装
	fmt.Println("Verifying installation...")
	for component := range downloaded {
		if err := installer.VerifyInstallation(component); err != nil {
			return fmt.Errorf("verification failed for %s: %w", component, err)
		}
	}

	fmt.Println("\nUpdate completed successfully!")
	return nil
}

func doRollback(cfg *Config) error {
	backupMgr := backup.NewManager(cfg.Backup.BackupDir, cfg.Backup.KeepCount)
	rollbackMgr := rollback.NewManager(backupMgr)

	if !rollbackMgr.CanRollback() {
		return fmt.Errorf("no backups available for rollback")
	}

	return rollbackMgr.RollbackToLatest()
}

func getInstalledVersions(cfg *Config) map[types.ComponentType]string {
	versions := make(map[types.ComponentType]string)

	// 这里应该实际检测已安装版本
	// 简化处理：从版本文件读取或返回默认值
	versions[types.ComponentCore] = "0.0.0"
	versions[types.ComponentWecom] = "0.0.0"
	versions[types.ComponentDingtalk] = "0.0.0"
	versions[types.ComponentFeishu] = "0.0.0"

	return versions
}

func getInstalledComponents(cfg *Config) []types.InstalledComponent {
	var components []types.InstalledComponent

	// Core
	components = append(components, types.InstalledComponent{
		Type:        types.ComponentCore,
		Name:        "OpenClaw Core",
		Version:     "0.0.0", // 应该从实际检测获取
		InstallPath: cfg.Paths.InstallDir,
		ConfigPath:  cfg.Paths.ConfigDir,
		InstallDate: time.Now(),
	})

	// Adapters
	for _, adapter := range cfg.Components.Adapters {
		var compType types.ComponentType
		switch adapter {
		case "wecom":
			compType = types.ComponentWecom
		case "dingtalk":
			compType = types.ComponentDingtalk
		case "feishu":
			compType = types.ComponentFeishu
		}

		components = append(components, types.InstalledComponent{
			Type:        compType,
			Name:        adapter + "-adapter",
			Version:     "0.0.0",
			InstallPath: filepath.Join(cfg.Paths.InstallDir, adapter+"-adapter"),
			ConfigPath:  filepath.Join(cfg.Paths.ConfigDir, adapter+"-adapter.yaml"),
			InstallDate: time.Now(),
		})
	}

	return components
}

func getInstallPaths(cfg *Config) map[types.ComponentType]string {
	paths := make(map[types.ComponentType]string)
	paths[types.ComponentCore] = cfg.Paths.InstallDir
	paths[types.ComponentWecom] = filepath.Join(cfg.Paths.InstallDir, "wecom-adapter")
	paths[types.ComponentDingtalk] = filepath.Join(cfg.Paths.InstallDir, "dingtalk-adapter")
	paths[types.ComponentFeishu] = filepath.Join(cfg.Paths.InstallDir, "feishu-adapter")
	return paths
}

func getConfigPaths(cfg *Config) map[types.ComponentType]string {
	paths := make(map[types.ComponentType]string)
	paths[types.ComponentCore] = filepath.Join(cfg.Paths.ConfigDir, "openclaw.yaml")
	paths[types.ComponentWecom] = filepath.Join(cfg.Paths.ConfigDir, "wecom-adapter.yaml")
	paths[types.ComponentDingtalk] = filepath.Join(cfg.Paths.ConfigDir, "dingtalk-adapter.yaml")
	paths[types.ComponentFeishu] = filepath.Join(cfg.Paths.ConfigDir, "feishu-adapter.yaml")
	return paths
}

func filterUpdates(updates []types.UpdateInfo, components []types.ComponentType) []types.UpdateInfo {
	componentSet := make(map[types.ComponentType]bool)
	for _, c := range components {
		componentSet[c] = true
	}

	var filtered []types.UpdateInfo
	for _, u := range updates {
		if componentSet[u.Component] {
			filtered = append(filtered, u)
		}
	}
	return filtered
}
