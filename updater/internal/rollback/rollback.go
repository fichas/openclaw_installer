// Package rollback 提供回滚功能
package rollback

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/openclaw/updater/internal/backup"
	"github.com/openclaw/updater/pkg/types"
)

// Manager 提供回滚管理功能
type Manager struct {
	backupManager *backup.Manager
}

// NewManager 创建新的回滚管理器
func NewManager(backupManager *backup.Manager) *Manager {
	return &Manager{
		backupManager: backupManager,
	}
}

// Rollback 执行回滚到指定备份
func (m *Manager) Rollback(backupID string) error {
	// 获取备份信息
	manifest, err := m.backupManager.GetBackup(backupID)
	if err != nil {
		return fmt.Errorf("failed to get backup %s: %w", backupID, err)
	}

	// 验证备份完整性
	if err := m.verifyBackup(manifest); err != nil {
		return fmt.Errorf("backup verification failed: %w", err)
	}

	// 回滚每个组件
	for _, component := range manifest.Components {
		if err := m.rollbackComponent(component); err != nil {
			return fmt.Errorf("failed to rollback component %s: %w", component.Name, err)
		}
	}

	return nil
}

// RollbackToLatest 回滚到最新的备份
func (m *Manager) RollbackToLatest() error {
	backups, err := m.backupManager.ListBackups()
	if err != nil {
		return fmt.Errorf("failed to list backups: %w", err)
	}

	if len(backups) == 0 {
		return fmt.Errorf("no backups available for rollback")
	}

	// 找到最新的备份
	latest := backups[0]
	for _, b := range backups {
		if b.Timestamp.After(latest.Timestamp) {
			latest = b
		}
	}

	return m.Rollback(latest.ID)
}

// verifyBackup 验证备份完整性
func (m *Manager) verifyBackup(manifest *types.BackupManifest) error {
	for _, component := range manifest.Components {
		for _, file := range component.Files {
			// 检查备份文件是否存在
			if _, err := os.Stat(file.BackupPath); err != nil {
				return fmt.Errorf("backup file missing: %s", file.BackupPath)
			}
		}
	}
	return nil
}

// rollbackComponent 回滚单个组件
func (m *Manager) rollbackComponent(component types.BackupComponent) error {
	for _, file := range component.Files {
		// 确保目标目录存在
		targetDir := filepath.Dir(file.SourcePath)
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", targetDir, err)
		}

		// 如果目标文件存在，先删除
		if _, err := os.Stat(file.SourcePath); err == nil {
			if err := os.Remove(file.SourcePath); err != nil {
				return fmt.Errorf("failed to remove existing file %s: %w", file.SourcePath, err)
			}
		}

		// 复制备份文件到原位置
		if err := m.copyFile(file.BackupPath, file.SourcePath); err != nil {
			return fmt.Errorf("failed to restore file %s: %w", file.SourcePath, err)
		}

		// 恢复原始权限
		if err := os.Chmod(file.SourcePath, os.FileMode(file.Mode)); err != nil {
			// 权限恢复失败不中断，仅记录
			fmt.Fprintf(os.Stderr, "Warning: failed to restore permissions for %s: %v\n", file.SourcePath, err)
		}
	}

	return nil
}

// copyFile 复制文件
func (m *Manager) copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, data, 0644)
}

// CanRollback 检查是否可以回滚
func (m *Manager) CanRollback() bool {
	backups, err := m.backupManager.ListBackups()
	if err != nil {
		return false
	}
	return len(backups) > 0
}

// GetRollbackHistory 获取可回滚的历史版本
func (m *Manager) GetRollbackHistory() ([]types.BackupInfo, error) {
	return m.backupManager.ListBackups()
}
