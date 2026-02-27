// Package updater 提供 Windows 平台自动更新功能
package updater

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/openclaw/updater/internal/backup"
	"github.com/openclaw/updater/internal/rollback"
	"github.com/openclaw/updater/pkg/types"
)

// WindowsUpdater 实现 Windows 平台的自动更新功能
type WindowsUpdater struct {
	installDir    string
	backupDir     string
	tempDir       string
	manifestURL   string
	httpClient    *http.Client
	backupManager *backup.Manager
	rollbackMgr   *rollback.Manager
	logger        Logger
}

// Logger 定义日志接口
type Logger interface {
	Info(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Error(format string, args ...interface{})
}

// defaultLogger 默认日志实现
type defaultLogger struct{}

func (l *defaultLogger) Info(format string, args ...interface{})  { fmt.Printf("[INFO] "+format+"\n", args...) }
func (l *defaultLogger) Warn(format string, args ...interface{})  { fmt.Printf("[WARN] "+format+"\n", args...) }
func (l *defaultLogger) Error(format string, args ...interface{}) { fmt.Printf("[ERROR] "+format+"\n", args...) }

// WindowsUpdaterOptions 包含 Windows 更新器配置选项
type WindowsUpdaterOptions struct {
	InstallDir  string
	BackupDir   string
	ManifestURL string
	Logger      Logger
}

// NewWindowsUpdater 创建新的 Windows 更新器实例
func NewWindowsUpdater(opts WindowsUpdaterOptions) *WindowsUpdater {
	if opts.InstallDir == "" {
		opts.InstallDir = `C:\Program Files\OpenClaw`
	}
	if opts.BackupDir == "" {
		opts.BackupDir = `C:\ProgramData\OpenClaw\backups`
	}
	if opts.Logger == nil {
		opts.Logger = &defaultLogger{}
	}

	tempDir := filepath.Join(os.TempDir(), "openclaw-updates")
	bm := backup.NewManager(opts.BackupDir, 5)
	rm := rollback.NewManager(bm)

	return &WindowsUpdater{
		installDir:    opts.InstallDir,
		backupDir:     opts.BackupDir,
		tempDir:       tempDir,
		manifestURL:   opts.ManifestURL,
		httpClient:    &http.Client{Timeout: 5 * time.Minute},
		backupManager: bm,
		rollbackMgr:   rm,
		logger:        opts.Logger,
	}
}

// CheckForUpdates 检查是否有可用更新
func (w *WindowsUpdater) CheckForUpdates() ([]types.UpdateInfo, error) {
	w.logger.Info("Checking for updates from %s", w.manifestURL)

	// 获取远程清单
	manifest, err := w.fetchRemoteManifest()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch manifest: %w", err)
	}

	// 获取已安装版本
	installed := w.getInstalledVersions()

	// 检查可用更新
	var updates []types.UpdateInfo

	// 检查核心更新
	if coreVer, ok := installed[types.ComponentCore]; ok {
		if compareVersions(coreVer, manifest.Core.Version) < 0 {
			updates = append(updates, types.UpdateInfo{
				Component:    types.ComponentCore,
				Name:         "OpenClaw Core",
				CurrentVer:   coreVer,
				AvailableVer: manifest.Core.Version,
				DownloadURL:  w.expandURL(manifest.Core.DownloadURL),
				Checksum:     manifest.Core.Checksum,
				ReleaseNotes: manifest.Core.ReleaseNotes,
			})
		}
	}

	// 检查适配器更新
	for id, adapter := range manifest.Adapters {
		componentType := adapterIDToComponent(id)
		if installedVer, ok := installed[componentType]; ok {
			if compareVersions(installedVer, adapter.Version) < 0 {
				// 检查核心版本兼容性
				if adapter.MinCoreVersion != "" {
					if coreVer, ok := installed[types.ComponentCore]; ok {
						if compareVersions(coreVer, adapter.MinCoreVersion) < 0 {
							w.logger.Warn("Skipping adapter %s update: requires core %s, have %s",
								id, adapter.MinCoreVersion, coreVer)
							continue
						}
					}
				}

				updates = append(updates, types.UpdateInfo{
					Component:    componentType,
					Name:         getAdapterName(id),
					CurrentVer:   installedVer,
					AvailableVer: adapter.Version,
					DownloadURL:  w.expandURL(adapter.DownloadURL),
					Checksum:     adapter.Checksum,
					ReleaseNotes: adapter.ReleaseNotes,
				})
			}
		}
	}

	if len(updates) > 0 {
		w.logger.Info("Found %d available updates", len(updates))
	} else {
		w.logger.Info("No updates available")
	}

	return updates, nil
}

// DownloadUpdate 下载更新包（后台静默下载）
func (w *WindowsUpdater) DownloadUpdate(update types.UpdateInfo) (string, error) {
	w.logger.Info("Downloading update for %s from %s", update.Name, update.DownloadURL)

	// 创建临时目录
	if err := os.MkdirAll(w.tempDir, 0750); err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	// 生成文件名
	filename := filepath.Base(update.DownloadURL)
	if filename == "" {
		filename = fmt.Sprintf("%s-%s.zip", update.Component, update.AvailableVer)
	}

	downloadPath := filepath.Join(w.tempDir, filename)

	// 创建请求
	req, err := http.NewRequest("GET", update.DownloadURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// 发送请求
	resp, err := w.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// 创建临时文件
	tempFile := downloadPath + ".tmp"
	file, err := os.Create(tempFile)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}

	// 下载并计算校验和
	hasher := sha256.New()
	writer := io.MultiWriter(file, hasher)

	_, err = io.Copy(writer, resp.Body)
	file.Close()

	if err != nil {
		os.Remove(tempFile)
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	// 验证校验和
	if update.Checksum != "" {
		actualChecksum := hex.EncodeToString(hasher.Sum(nil))
		expectedChecksum := strings.TrimPrefix(update.Checksum, "sha256:")

		if actualChecksum != expectedChecksum {
			os.Remove(tempFile)
			return "", fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
		}

		w.logger.Info("Checksum verification passed for %s", update.Name)
	}

	// 重命名为最终文件名
	if err := os.Rename(tempFile, downloadPath); err != nil {
		os.Remove(tempFile)
		return "", fmt.Errorf("failed to rename file: %w", err)
	}

	w.logger.Info("Successfully downloaded %s to %s", update.Name, downloadPath)

	// 保存更新信息
	infoPath := downloadPath + ".json"
	infoData, _ := json.MarshalIndent(update, "", "  ")
	os.WriteFile(infoPath, infoData, 0644)

	return downloadPath, nil
}

// VerifySignature 验证更新包签名（Windows 使用 Authenticode）
func (w *WindowsUpdater) VerifySignature(packagePath string) error {
	w.logger.Info("Verifying signature for %s", packagePath)

	// 检查文件是否存在
	if _, err := os.Stat(packagePath); err != nil {
		return fmt.Errorf("package not found: %w", err)
	}

	// 使用 Windows signtool 验证签名
	// 如果 signtool 不可用，则使用 PowerShell Get-AuthenticodeSignature
	signtoolPath := w.findSigntool()
	if signtoolPath != "" {
		cmd := exec.Command(signtoolPath, "verify", "/pa", packagePath)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("signature verification failed: %w, output: %s", err, string(output))
		}
	} else {
		// 使用 PowerShell 作为备选
		psCmd := fmt.Sprintf(
			"Get-AuthenticodeSignature -FilePath '%s' | Select-Object -ExpandProperty Status",
			packagePath,
		)
		cmd := exec.Command("powershell.exe", "-Command", psCmd)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to check signature via PowerShell: %w", err)
		}

		status := strings.TrimSpace(string(output))
		if status != "Valid" {
			return fmt.Errorf("invalid signature status: %s", status)
		}
	}

	w.logger.Info("Signature verification passed for %s", packagePath)
	return nil
}

// findSigntool 查找 signtool.exe 路径
func (w *WindowsUpdater) findSigntool() string {
	// 常见 signtool 路径
	paths := []string{
		`C:\Program Files (x86)\Windows Kits\10\bin\10.0.22621.0\x64\signtool.exe`,
		`C:\Program Files (x86)\Windows Kits\10\bin\10.0.22000.0\x64\signtool.exe`,
		`C:\Program Files (x86)\Windows Kits\10\bin\10.0.19041.0\x64\signtool.exe`,
		`C:\Program Files (x86)\Windows Kits\10\bin\10.0.18362.0\x64\signtool.exe`,
		`C:\Program Files (x86)\Windows Kits\8.1\bin\x64\signtool.exe`,
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

// InstallUpdate 安装更新
func (w *WindowsUpdater) InstallUpdate(update types.UpdateInfo, packagePath string) error {
	w.logger.Info("Installing update for %s", update.Name)

	// 验证签名
	if err := w.VerifySignature(packagePath); err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}

	// 创建备份
	components := []types.InstalledComponent{
		{
			Type:        update.Component,
			Name:        update.Name,
			Version:     update.CurrentVer,
			InstallPath: w.getComponentPath(update.Component),
		},
	}

	w.logger.Info("Creating backup before update...")
	backupInfo, err := w.backupManager.CreateBackup(components, update.CurrentVer)
	if err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}
	w.logger.Info("Backup created: %s", backupInfo.ID)

	// 解压更新包到临时目录
	extractDir := filepath.Join(w.tempDir, "extract", string(update.Component))
	if err := os.RemoveAll(extractDir); err != nil {
		w.logger.Warn("Failed to clean extract directory: %v", err)
	}

	if err := w.extractPackage(packagePath, extractDir); err != nil {
		return fmt.Errorf("failed to extract package: %w", err)
	}

	// 停止服务（如果正在运行）
	if err := w.stopService(); err != nil {
		w.logger.Warn("Failed to stop service: %v", err)
	}

	// 执行文件替换
	installPath := w.getComponentPath(update.Component)
	if err := w.replaceFiles(extractDir, installPath); err != nil {
		// 回滚
		w.logger.Error("Update failed, rolling back...")
		if rbErr := w.rollbackMgr.Rollback(backupInfo.ID); rbErr != nil {
			return fmt.Errorf("update failed and rollback failed: %v (rollback error: %v)", err, rbErr)
		}
		return fmt.Errorf("update failed, rolled back to previous version: %w", err)
	}

	// 更新版本文件
	if err := w.saveVersion(update.Component, update.AvailableVer); err != nil {
		w.logger.Warn("Failed to save version file: %v", err)
	}

	// 启动服务
	if err := w.startService(); err != nil {
		w.logger.Warn("Failed to start service: %v", err)
	}

	w.logger.Info("Successfully installed update for %s to version %s", update.Name, update.AvailableVer)

	// 清理临时文件
	os.Remove(packagePath)
	os.Remove(packagePath + ".json")

	return nil
}

// replaceFiles 替换文件
func (w *WindowsUpdater) replaceFiles(sourceDir, destDir string) error {
	// 确保目标目录存在
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// 遍历源目录
	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 计算相对路径
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(destDir, relPath)

		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		}

		// 如果目标文件存在，先删除（Windows 不能直接覆盖正在使用的文件）
		if _, err := os.Stat(destPath); err == nil {
			// 尝试删除，如果失败可能是因为文件正在使用
			if err := os.Remove(destPath); err != nil {
				// 重命名旧文件
				backupPath := destPath + ".old"
				if err := os.Rename(destPath, backupPath); err != nil {
					return fmt.Errorf("failed to backup existing file %s: %w", destPath, err)
				}
			}
		}

		// 复制新文件
		if err := w.copyFile(path, destPath); err != nil {
			return fmt.Errorf("failed to copy file %s: %w", path, err)
		}

		return nil
	})
}

// copyFile 复制文件
func (w *WindowsUpdater) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// extractPackage 解压安装包
func (w *WindowsUpdater) extractPackage(packagePath, destDir string) error {
	if strings.HasSuffix(packagePath, ".zip") {
		return w.extractZip(packagePath, destDir)
	}
	return fmt.Errorf("unsupported package format: %s", packagePath)
}

// extractZip 解压 ZIP 文件
func (w *WindowsUpdater) extractZip(zipPath, destDir string) error {
	// 使用 PowerShell 解压（Windows 原生支持）
	psCmd := fmt.Sprintf(
		"Expand-Archive -Path '%s' -DestinationPath '%s' -Force",
		zipPath, destDir,
	)
	cmd := exec.Command("powershell.exe", "-Command", psCmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to extract zip: %w, output: %s", err, string(output))
	}
	return nil
}

// stopService 停止服务
func (w *WindowsUpdater) stopService() error {
	// 使用 SC 命令停止服务
	cmd := exec.Command("sc", "stop", "OpenClaw")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// 服务可能不存在或已经停止
		outputStr := string(output)
		if strings.Contains(outputStr, "1060") || strings.Contains(outputStr, "1062") {
			// 1060: 服务不存在
			// 1062: 服务未启动
			return nil
		}
		return fmt.Errorf("failed to stop service: %w, output: %s", err, outputStr)
	}

	// 等待服务停止
	time.Sleep(2 * time.Second)
	return nil
}

// startService 启动服务
func (w *WindowsUpdater) startService() error {
	cmd := exec.Command("sc", "start", "OpenClaw")
	output, err := cmd.CombinedOutput()
	if err != nil {
		outputStr := string(output)
		if strings.Contains(outputStr, "1060") {
			// 服务不存在，尝试直接启动程序
			return w.startApplication()
		}
		return fmt.Errorf("failed to start service: %w, output: %s", err, outputStr)
	}
	return nil
}

// startApplication 直接启动应用程序
func (w *WindowsUpdater) startApplication() error {
	exePath := filepath.Join(w.installDir, "openclaw.exe")
	if _, err := os.Stat(exePath); err != nil {
		return fmt.Errorf("executable not found: %w", err)
	}

	cmd := exec.Command(exePath)
	cmd.Dir = w.installDir

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start application: %w", err)
	}

	return nil
}

// getComponentPath 获取组件安装路径
func (w *WindowsUpdater) getComponentPath(component types.ComponentType) string {
	switch component {
	case types.ComponentCore:
		return w.installDir
	case types.ComponentWecom:
		return filepath.Join(w.installDir, "adapters", "wecom")
	case types.ComponentDingtalk:
		return filepath.Join(w.installDir, "adapters", "dingtalk")
	case types.ComponentFeishu:
		return filepath.Join(w.installDir, "adapters", "feishu")
	default:
		return filepath.Join(w.installDir, string(component))
	}
}

// saveVersion 保存版本信息
func (w *WindowsUpdater) saveVersion(component types.ComponentType, version string) error {
	versionFile := filepath.Join(w.getComponentPath(component), ".version")
	info := types.InstalledComponent{
		Type:        component,
		Version:     version,
		InstallDate: time.Now(),
	}

	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(versionFile, data, 0644)
}

// fetchRemoteManifest 获取远程清单
func (w *WindowsUpdater) fetchRemoteManifest() (*types.ReleaseManifest, error) {
	resp, err := w.httpClient.Get(w.manifestURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var manifest types.ReleaseManifest
	if err := json.Unmarshal(body, &manifest); err != nil {
		return nil, err
	}

	return &manifest, nil
}

// getInstalledVersions 获取已安装版本
func (w *WindowsUpdater) getInstalledVersions() map[types.ComponentType]string {
	installed := make(map[types.ComponentType]string)

	// 检查核心版本
	coreVersion := w.readVersion(types.ComponentCore)
	installed[types.ComponentCore] = coreVersion

	// 检查适配器版本
	adapters := []types.ComponentType{
		types.ComponentWecom,
		types.ComponentDingtalk,
		types.ComponentFeishu,
	}

	for _, adapter := range adapters {
		version := w.readVersion(adapter)
		if version != "0.0.0" {
			installed[adapter] = version
		}
	}

	return installed
}

// readVersion 读取版本文件
func (w *WindowsUpdater) readVersion(component types.ComponentType) string {
	versionFile := filepath.Join(w.getComponentPath(component), ".version")
	data, err := os.ReadFile(versionFile)
	if err != nil {
		return "0.0.0"
	}

	var info types.InstalledComponent
	if err := json.Unmarshal(data, &info); err != nil {
		return "0.0.0"
	}

	return info.Version
}

// expandURL 扩展 URL 模板变量
func (w *WindowsUpdater) expandURL(url string) string {
	url = strings.ReplaceAll(url, "{os}", "windows")
	url = strings.ReplaceAll(url, "{arch}", runtime.GOARCH)
	url = strings.ReplaceAll(url, "{ext}", "zip")
	return url
}

// Rollback 回滚到之前的版本
func (w *WindowsUpdater) Rollback(backupID string) error {
	w.logger.Info("Rolling back to backup %s", backupID)

	if err := w.stopService(); err != nil {
		w.logger.Warn("Failed to stop service before rollback: %v", err)
	}

	if err := w.rollbackMgr.Rollback(backupID); err != nil {
		return fmt.Errorf("rollback failed: %w", err)
	}

	if err := w.startService(); err != nil {
		w.logger.Warn("Failed to start service after rollback: %v", err)
	}

	w.logger.Info("Rollback completed successfully")
	return nil
}

// RunAutoUpdate 执行完整的自动更新流程
func (w *WindowsUpdater) RunAutoUpdate(ctx context.Context) error {
	w.logger.Info("Starting automatic update process")

	// 1. 检查更新
	updates, err := w.CheckForUpdates()
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	if len(updates) == 0 {
		w.logger.Info("No updates available")
		return nil
	}

	// 2. 下载并安装每个更新
	for _, update := range updates {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// 下载更新
		packagePath, err := w.DownloadUpdate(update)
		if err != nil {
			w.logger.Error("Failed to download update for %s: %v", update.Name, err)
			continue
		}

		// 安装更新
		if err := w.InstallUpdate(update, packagePath); err != nil {
			w.logger.Error("Failed to install update for %s: %v", update.Name, err)
			continue
		}
	}

	w.logger.Info("Automatic update process completed")
	return nil
}

// ScheduleUpdate 计划更新（用于后台服务）
func (w *WindowsUpdater) ScheduleUpdate(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
			if err := w.RunAutoUpdate(ctx); err != nil {
				w.logger.Error("Scheduled update failed: %v", err)
			}
			cancel()
		}
	}()
}

// 辅助函数

func adapterIDToComponent(id string) types.ComponentType {
	switch id {
	case "wecom":
		return types.ComponentWecom
	case "dingtalk":
		return types.ComponentDingtalk
	case "feishu":
		return types.ComponentFeishu
	default:
		return types.ComponentType(id)
	}
}

func getAdapterName(id string) string {
	names := map[string]string{
		"wecom":    "企业微信适配器",
		"dingtalk": "钉钉适配器",
		"feishu":   "飞书适配器",
	}
	if name, ok := names[id]; ok {
		return name
	}
	return id
}

// compareVersions 比较版本号
// 返回 -1 如果 v1 < v2, 0 如果 v1 == v2, 1 如果 v1 > v2
func compareVersions(v1, v2 string) int {
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var n1, n2 int
		if i < len(parts1) {
			fmt.Sscanf(parts1[i], "%d", &n1)
		}
		if i < len(parts2) {
			fmt.Sscanf(parts2[i], "%d", &n2)
		}

		if n1 < n2 {
			return -1
		}
		if n1 > n2 {
			return 1
		}
	}

	return 0
}
