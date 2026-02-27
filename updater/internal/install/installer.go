// Package install 提供安装功能
package install

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/openclaw/updater/pkg/types"
)

// Installer 提供安装功能
type Installer struct {
	installPaths map[types.ComponentType]string
	configPaths  map[types.ComponentType]string
}

// NewInstaller 创建新的安装器
func NewInstaller(installPaths, configPaths map[types.ComponentType]string) *Installer {
	return &Installer{
		installPaths: installPaths,
		configPaths:  configPaths,
	}
}

// InstallComponent 安装组件
func (i *Installer) InstallComponent(componentType types.ComponentType, packagePath string, options types.UpdateOptions) error {
	installPath := i.installPaths[componentType]
	if installPath == "" {
		return fmt.Errorf("no install path configured for component %s", componentType)
	}

	// 创建临时目录
	tempDir := filepath.Join(os.TempDir(), fmt.Sprintf("openclaw-install-%s", componentType))
	if err := os.MkdirAll(tempDir, 0750); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// 解压包
	if err := i.extractPackage(packagePath, tempDir); err != nil {
		return fmt.Errorf("failed to extract package: %w", err)
	}

	// 如果是 dry run，到此结束
	if options.DryRun {
		return nil
	}

	// 确保安装目录存在
	if err := os.MkdirAll(installPath, 0755); err != nil {
		return fmt.Errorf("failed to create install directory: %w", err)
	}

	// 复制文件到安装目录
	if err := i.copyInstallFiles(tempDir, installPath); err != nil {
		return fmt.Errorf("failed to copy files: %w", err)
	}

	// 设置权限
	if err := i.setPermissions(installPath); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	return nil
}

// extractPackage 解压安装包
func (i *Installer) extractPackage(packagePath, destDir string) error {
	if strings.HasSuffix(packagePath, ".zip") {
		return i.extractZip(packagePath, destDir)
	}
	if strings.HasSuffix(packagePath, ".tar.gz") || strings.HasSuffix(packagePath, ".tgz") {
		return i.extractTarGz(packagePath, destDir)
	}
	return fmt.Errorf("unsupported package format: %s", packagePath)
}

// extractZip 解压 ZIP 文件
func (i *Installer) extractZip(zipPath, destDir string) error {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer reader.Close()

	for _, file := range reader.File {
		path := filepath.Join(destDir, file.Name)

		// 安全检查：防止路径遍历
		if !strings.HasPrefix(path, filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			os.MkdirAll(path, file.Mode())
			continue
		}

		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}

		outFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}

		rc, err := file.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}

	return nil
}

// extractTarGz 解压 tar.gz 文件
func (i *Installer) extractTarGz(tarGzPath, destDir string) error {
	file, err := os.Open(tarGzPath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		path := filepath.Join(destDir, header.Name)

		// 安全检查：防止路径遍历
		if !strings.HasPrefix(path, filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(path, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				return err
			}

			outFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}

	return nil
}

// copyInstallFiles 复制安装文件
func (i *Installer) copyInstallFiles(sourceDir, destDir string) error {
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

		// 复制文件
		return i.copyFile(path, destPath)
	})
}

// copyFile 复制单个文件
func (i *Installer) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// 获取源文件信息
	stat, err := sourceFile.Stat()
	if err != nil {
		return err
	}

	// 创建目标文件
	destFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, stat.Mode())
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// setPermissions 设置文件权限
func (i *Installer) setPermissions(installPath string) error {
	// Windows 不需要特殊权限设置
	if runtime.GOOS == "windows" {
		return nil
	}

	return filepath.Walk(installPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return os.Chmod(path, 0755)
		}

		// 可执行文件设置 0755，其他文件 0644
		if i.isExecutable(path) {
			return os.Chmod(path, 0755)
		}
		return os.Chmod(path, 0644)
	})
}

// isExecutable 检查文件是否为可执行文件
func (i *Installer) isExecutable(path string) bool {
	// 检查文件扩展名
	if runtime.GOOS == "windows" {
		ext := strings.ToLower(filepath.Ext(path))
		return ext == ".exe" || ext == ".bat" || ext == ".cmd"
	}

	// Unix 系统：检查是否有可执行权限
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	return info.Mode()&0111 != 0
}

// VerifyInstallation 验证安装
func (i *Installer) VerifyInstallation(componentType types.ComponentType) error {
	installPath := i.installPaths[componentType]
	if installPath == "" {
		return fmt.Errorf("no install path configured for component %s", componentType)
	}

	// 检查安装目录是否存在
	info, err := os.Stat(installPath)
	if err != nil {
		return fmt.Errorf("installation directory not found: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("installation path is not a directory")
	}

	// 检查是否有文件
	entries, err := os.ReadDir(installPath)
	if err != nil {
		return fmt.Errorf("failed to read installation directory: %w", err)
	}
	if len(entries) == 0 {
		return fmt.Errorf("installation directory is empty")
	}

	return nil
}
