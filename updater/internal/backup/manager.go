// Package backup 提供备份管理功能
package backup

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/openclaw/updater/pkg/types"
)

// Manager 提供备份管理功能
type Manager struct {
	backupDir string
	keepCount int
}

// NewManager 创建新的备份管理器
func NewManager(backupDir string, keepCount int) *Manager {
	return &Manager{
		backupDir: backupDir,
		keepCount: keepCount,
	}
}

// Init 初始化备份目录
func (m *Manager) Init() error {
	return os.MkdirAll(m.backupDir, 0750)
}

// CreateBackup 创建备份
func (m *Manager) CreateBackup(components []types.InstalledComponent, version string) (*types.BackupInfo, error) {
	// 生成备份 ID
	backupID := time.Now().Format("20060102-150405")
	backupPath := filepath.Join(m.backupDir, fmt.Sprintf("backup-%s", backupID))

	// 创建备份目录
	if err := os.MkdirAll(backupPath, 0750); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %w", err)
	}

	manifest := types.BackupManifest{
		BackupID:  backupID,
		Timestamp: time.Now(),
		Version:   version,
	}

	// 备份每个组件
	for _, component := range components {
		backupComponent, err := m.backupComponent(component, backupPath)
		if err != nil {
			return nil, fmt.Errorf("failed to backup component %s: %w", component.Name, err)
		}
		manifest.Components = append(manifest.Components, *backupComponent)
	}

	// 保存备份清单
	manifestPath := filepath.Join(backupPath, "manifest.json")
	manifestData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal manifest: %w", err)
	}

	if err := os.WriteFile(manifestPath, manifestData, 0640); err != nil {
		return nil, fmt.Errorf("failed to write manifest: %w", err)
	}

	// 计算备份大小
	size, err := m.calculateBackupSize(backupPath)
	if err != nil {
		size = 0
	}

	// 清理旧备份
	if err := m.cleanupOldBackups(); err != nil {
		// 记录错误但不中断
		fmt.Fprintf(os.Stderr, "Warning: failed to cleanup old backups: %v\n", err)
	}

	return &types.BackupInfo{
		ID:         backupID,
		Timestamp:  time.Now(),
		Version:    version,
		Components: m.getComponentNames(components),
		BackupPath: backupPath,
		Size:       size,
	}, nil
}

// backupComponent 备份单个组件
func (m *Manager) backupComponent(component types.InstalledComponent, backupPath string) (*types.BackupComponent, error) {
	backupComponent := &types.BackupComponent{
		Type:    component.Type,
		Name:    component.Name,
		Version: component.Version,
	}

	// 创建组件备份目录
	componentBackupDir := filepath.Join(backupPath, string(component.Type))
	if err := os.MkdirAll(componentBackupDir, 0750); err != nil {
		return nil, err
	}

	// 备份安装目录中的文件
	if component.InstallPath != "" {
		files, err := m.backupDirectory(component.InstallPath, componentBackupDir)
		if err != nil {
			return nil, err
		}
		backupComponent.Files = append(backupComponent.Files, files...)
	}

	// 备份配置文件
	if component.ConfigPath != "" {
		configBackupDir := filepath.Join(componentBackupDir, "config")
		if err := os.MkdirAll(configBackupDir, 0750); err != nil {
			return nil, err
		}

		files, err := m.backupDirectory(component.ConfigPath, configBackupDir)
		if err != nil {
			return nil, err
		}
		backupComponent.Files = append(backupComponent.Files, files...)
	}

	return backupComponent, nil
}

// backupDirectory 备份目录中的所有文件
func (m *Manager) backupDirectory(sourceDir, backupDir string) ([]types.BackupFile, error) {
	var files []types.BackupFile

	err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录
		if info.IsDir() {
			return nil
		}

		// 计算相对路径
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		// 目标备份路径
		backupPath := filepath.Join(backupDir, relPath)

		// 创建子目录
		if err := os.MkdirAll(filepath.Dir(backupPath), 0750); err != nil {
			return err
		}

		// 复制文件
		if err := m.copyFile(path, backupPath); err != nil {
			return err
		}

		// 计算文件哈希
		hash, err := m.hashFile(backupPath)
		if err != nil {
			return err
		}

		files = append(files, types.BackupFile{
			SourcePath: path,
			BackupPath: backupPath,
			Hash:       hash,
			Mode:       uint32(info.Mode()),
		})

		return nil
	})

	return files, err
}

// copyFile 复制文件
func (m *Manager) copyFile(src, dst string) error {
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

// hashFile 计算文件 SHA256 哈希
func (m *Manager) hashFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// calculateBackupSize 计算备份大小
func (m *Manager) calculateBackupSize(backupPath string) (int64, error) {
	var size int64

	err := filepath.Walk(backupPath, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	return size, err
}

// getComponentNames 获取组件名称列表
func (m *Manager) getComponentNames(components []types.InstalledComponent) []string {
	names := make([]string, len(components))
	for i, c := range components {
		names[i] = c.Name
	}
	return names
}

// cleanupOldBackups 清理旧备份
func (m *Manager) cleanupOldBackups() error {
	entries, err := os.ReadDir(m.backupDir)
	if err != nil {
		return err
	}

	// 收集备份目录
	var backups []os.DirEntry
	for _, entry := range entries {
		if entry.IsDir() && len(entry.Name()) > 7 && entry.Name()[:7] == "backup-" {
			backups = append(backups, entry)
		}
	}

	// 如果备份数量未超过限制，不清理
	if len(backups) <= m.keepCount {
		return nil
	}

	// 按修改时间排序（旧的在前）
	for i := 0; i < len(backups)-1; i++ {
		for j := i + 1; j < len(backups); j++ {
			infoI, _ := backups[i].Info()
			infoJ, _ := backups[j].Info()
			if infoI.ModTime().After(infoJ.ModTime()) {
				backups[i], backups[j] = backups[j], backups[i]
			}
		}
	}

	// 删除多余的旧备份
	toDelete := len(backups) - m.keepCount
	for i := 0; i < toDelete; i++ {
		backupPath := filepath.Join(m.backupDir, backups[i].Name())
		if err := os.RemoveAll(backupPath); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to remove old backup %s: %v\n", backupPath, err)
		}
	}

	return nil
}

// ListBackups 列出所有备份
func (m *Manager) ListBackups() ([]types.BackupInfo, error) {
	entries, err := os.ReadDir(m.backupDir)
	if err != nil {
		return nil, err
	}

	var backups []types.BackupInfo
	for _, entry := range entries {
		if !entry.IsDir() || len(entry.Name()) <= 7 || entry.Name()[:7] != "backup-" {
			continue
		}

		backupPath := filepath.Join(m.backupDir, entry.Name())
		manifestPath := filepath.Join(backupPath, "manifest.json")

		// 读取清单
		data, err := os.ReadFile(manifestPath)
		if err != nil {
			continue
		}

		var manifest types.BackupManifest
		if err := json.Unmarshal(data, &manifest); err != nil {
			continue
		}

		// 计算大小
		size, _ := m.calculateBackupSize(backupPath)

		backups = append(backups, types.BackupInfo{
			ID:         manifest.BackupID,
			Timestamp:  manifest.Timestamp,
			Version:    manifest.Version,
			Components: m.getComponentNamesFromManifest(manifest.Components),
			BackupPath: backupPath,
			Size:       size,
		})
	}

	return backups, nil
}

// getComponentNamesFromManifest 从清单获取组件名称
func (m *Manager) getComponentNamesFromManifest(components []types.BackupComponent) []string {
	names := make([]string, len(components))
	for i, c := range components {
		names[i] = c.Name
	}
	return names
}

// GetBackup 获取指定备份信息
func (m *Manager) GetBackup(backupID string) (*types.BackupManifest, error) {
	backupPath := filepath.Join(m.backupDir, fmt.Sprintf("backup-%s", backupID))
	manifestPath := filepath.Join(backupPath, "manifest.json")

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup manifest: %w", err)
	}

	var manifest types.BackupManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse backup manifest: %w", err)
	}

	return &manifest, nil
}
