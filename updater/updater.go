package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Updater handles the update process
type Updater struct {
	platform   *Platform
	installDir string
	backupDir  string
	configPath string
	versionURL string
	adapter    string
	dryRun     bool
	verbose    bool
	logger     *Logger
	httpClient *http.Client
}

// UpdaterOptions contains options for creating an updater
type UpdaterOptions struct {
	Platform   *Platform
	InstallDir string
	BackupDir  string
	ConfigPath string
	VersionURL string
	Adapter    string
	DryRun     bool
	Verbose    bool
	Logger     *Logger
}

// UpdateInfo contains information about available updates
type UpdateInfo struct {
	CurrentVersion   string           `json:"current_version"`
	LatestVersion    string           `json:"latest_version"`
	ReleaseNotes     string           `json:"release_notes"`
	DownloadURL      string           `json:"download_url"`
	Checksum         string           `json:"checksum"`
	Mandatory        bool             `json:"mandatory"`
	AdapterUpdates   []AdapterUpdate  `json:"adapter_updates"`
}

// AdapterUpdate contains update information for an adapter
type AdapterUpdate struct {
	Name           string `json:"name"`
	CurrentVersion string `json:"current_version"`
	LatestVersion  string `json:"latest_version"`
	DownloadURL    string `json:"download_url"`
	Checksum       string `json:"checksum"`
}

// BackupInfo contains information about a backup
type BackupInfo struct {
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	Created   time.Time `json:"created"`
	Size      int64     `json:"size"`
	Version   string    `json:"version"`
	Components []string `json:"components"`
}

// NewUpdater creates a new updater instance
func NewUpdater(opts UpdaterOptions) *Updater {
	return &Updater{
		platform:   opts.Platform,
		installDir: opts.InstallDir,
		backupDir:  opts.BackupDir,
		configPath: opts.ConfigPath,
		versionURL: opts.VersionURL,
		adapter:    opts.Adapter,
		dryRun:     opts.DryRun,
		verbose:    opts.Verbose,
		logger:     opts.Logger,
		httpClient: &http.Client{
			Timeout: 5 * time.Minute,
		},
	}
}

// CheckForUpdates checks if updates are available
func (u *Updater) CheckForUpdates() (*UpdateInfo, error) {
	// Get current version
	currentVersion, err := u.getCurrentVersion()
	if err != nil {
		u.logger.Warn("Could not determine current version: %v", err)
		currentVersion = "0.0.0"
	}

	// Fetch version info from remote
	versionInfo, err := u.fetchVersionInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch version info: %w", err)
	}

	versionInfo.CurrentVersion = currentVersion

	// Check adapter updates
	adapters := u.getAdaptersToCheck()
	for _, adapterName := range adapters {
		adapterUpdate, err := u.checkAdapterUpdate(adapterName)
		if err != nil {
			u.logger.Warn("Failed to check adapter %s: %v", adapterName, err)
			continue
		}
		if adapterUpdate != nil {
			versionInfo.AdapterUpdates = append(versionInfo.AdapterUpdates, *adapterUpdate)
		}
	}

	return versionInfo, nil
}

// HasUpdate returns true if an update is available
func (u *UpdateInfo) HasUpdate() bool {
	if u.LatestVersion != u.CurrentVersion {
		return true
	}
	return len(u.AdapterUpdates) > 0
}

// Update performs the update
func (u *Updater) Update(info *UpdateInfo) error {
	// Create backup before update
	if !u.dryRun {
		u.logger.Info("Creating backup...")
		backupPath, err := u.createBackup()
		if err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
		u.logger.Info("Backup created: %s", backupPath)
	}

	// Update core if needed
	if info.LatestVersion != info.CurrentVersion {
		u.logger.Info("Updating core to version %s...", info.LatestVersion)
		if err := u.updateCore(info); err != nil {
			return fmt.Errorf("failed to update core: %w", err)
		}
		u.logger.Success("Core updated successfully")
	}

	// Update adapters
	for _, adapterUpdate := range info.AdapterUpdates {
		u.logger.Info("Updating adapter %s to version %s...", adapterUpdate.Name, adapterUpdate.LatestVersion)
		if err := u.updateAdapter(adapterUpdate); err != nil {
			return fmt.Errorf("failed to update adapter %s: %w", adapterUpdate.Name, err)
		}
		u.logger.Success("Adapter %s updated successfully", adapterUpdate.Name)
	}

	// Update version file
	if !u.dryRun {
		if err := u.saveVersion(info.LatestVersion); err != nil {
			u.logger.Warn("Failed to save version file: %v", err)
		}
	}

	return nil
}

// Rollback restores the previous version from backup
func (u *Updater) Rollback() error {
	// Find the most recent backup
	backups, err := u.ListBackups()
	if err != nil {
		return fmt.Errorf("failed to list backups: %w", err)
	}

	if len(backups) == 0 {
		return fmt.Errorf("no backups available for rollback")
	}

	// Sort by creation time (newest first)
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].Created.After(backups[j].Created)
	})

	latestBackup := backups[0]
	u.logger.Info("Rolling back to backup: %s (version %s)", latestBackup.Name, latestBackup.Version)

	// Extract backup
	if err := u.extractBackup(latestBackup.Path); err != nil {
		return fmt.Errorf("failed to extract backup: %w", err)
	}

	return nil
}

// ListBackups returns a list of available backups
func (u *Updater) ListBackups() ([]BackupInfo, error) {
	entries, err := os.ReadDir(u.backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []BackupInfo{}, nil
		}
		return nil, err
	}

	var backups []BackupInfo
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".tar.gz") {
			continue
		}

		path := filepath.Join(u.backupDir, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue
		}

		// Try to read backup metadata
		version := "unknown"
		metaPath := path + ".meta"
		if metaData, err := os.ReadFile(metaPath); err == nil {
			var meta BackupInfo
			if json.Unmarshal(metaData, &meta) == nil {
				version = meta.Version
			}
		}

		backups = append(backups, BackupInfo{
			Name:    entry.Name(),
			Path:    path,
			Created: info.ModTime(),
			Size:    info.Size(),
			Version: version,
		})
	}

	return backups, nil
}

// CleanupOldBackups removes old backups keeping only the specified number
func (u *Updater) CleanupOldBackups(maxBackups int) error {
	if maxBackups <= 0 {
		return nil
	}

	backups, err := u.ListBackups()
	if err != nil {
		return err
	}

	if len(backups) <= maxBackups {
		return nil
	}

	// Sort by creation time (oldest first)
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].Created.Before(backups[j].Created)
	})

	// Remove oldest backups
	for i := 0; i < len(backups)-maxBackups; i++ {
		u.logger.Info("Removing old backup: %s", backups[i].Name)
		if err := os.Remove(backups[i].Path); err != nil {
			u.logger.Warn("Failed to remove backup %s: %v", backups[i].Name, err)
		}
		// Also remove metadata file
		metaPath := backups[i].Path + ".meta"
		os.Remove(metaPath)
	}

	return nil
}

// Private helper methods

func (u *Updater) getCurrentVersion() (string, error) {
	versionFile := filepath.Join(u.installDir, "version.txt")
	data, err := os.ReadFile(versionFile)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func (u *Updater) saveVersion(version string) error {
	versionFile := filepath.Join(u.installDir, "version.txt")
	return os.WriteFile(versionFile, []byte(version), 0644)
}

func (u *Updater) fetchVersionInfo() (*UpdateInfo, error) {
	// For now, simulate version info fetch
	// In production, this would make an HTTP request to versionURL

	// Simulate API response
	resp := &UpdateInfo{
		LatestVersion: "1.1.0",
		ReleaseNotes:  "Bug fixes and performance improvements",
		DownloadURL:   fmt.Sprintf("https://download.openclaw.io/core/openclaw-core-%s%s", u.platform.String(), u.platform.GetPackageExtension()),
		Checksum:      "sha256:abc123...",
		Mandatory:     false,
	}

	return resp, nil
}

func (u *Updater) getAdaptersToCheck() []string {
	if u.adapter == "all" {
		return []string{"wecom", "dingtalk", "feishu"}
	}
	return []string{u.adapter}
}

func (u *Updater) checkAdapterUpdate(name string) (*AdapterUpdate, error) {
	// Get current adapter version
	currentVersion, _ := u.getAdapterVersion(name)

	// Simulate fetching latest version
	// In production, this would query an API
	latestVersion := "1.0.1" // Simulated

	if latestVersion == currentVersion {
		return nil, nil // No update available
	}

	return &AdapterUpdate{
		Name:           name,
		CurrentVersion: currentVersion,
		LatestVersion:  latestVersion,
		DownloadURL:    fmt.Sprintf("https://download.openclaw.io/adapters/%s-adapter-%s%s", name, u.platform.String(), u.platform.GetPackageExtension()),
		Checksum:       "sha256:def456...",
	}, nil
}

func (u *Updater) getAdapterVersion(name string) (string, error) {
	versionFile := filepath.Join(u.installDir, fmt.Sprintf("%s-adapter.version", name))
	data, err := os.ReadFile(versionFile)
	if err != nil {
		return "0.0.0", err
	}
	return strings.TrimSpace(string(data)), nil
}

func (u *Updater) createBackup() (string, error) {
	if u.dryRun {
		return "dry-run-backup", nil
	}

	timestamp := time.Now().Format("20060102-150405")
	backupName := fmt.Sprintf("openclaw-backup-%s.tar.gz", timestamp)
	backupPath := filepath.Join(u.backupDir, backupName)

	// Create backup archive
	if err := u.createTarGz(backupPath, u.installDir); err != nil {
		return "", err
	}

	// Save metadata
	meta := BackupInfo{
		Name:      backupName,
		Path:      backupPath,
		Created:   time.Now(),
		Version:   "unknown", // Would be filled from actual version
		Components: []string{"core"},
	}

	metaData, _ := json.MarshalIndent(meta, "", "  ")
	metaPath := backupPath + ".meta"
	os.WriteFile(metaPath, metaData, 0644)

	return backupPath, nil
}

func (u *Updater) createTarGz(archivePath, sourceDir string) error {
	file, err := os.Create(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	gw := gzip.NewWriter(file)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip version.txt and other dynamic files
		if info.Name() == "version.txt" || info.Name() == "openclaw.pid" {
			return nil
		}

		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}

		relPath, _ := filepath.Rel(sourceDir, path)
		header.Name = relPath

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if !info.IsDir() {
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()

			_, err = io.Copy(tw, f)
			return err
		}

		return nil
	})
}

func (u *Updater) extractBackup(backupPath string) error {
	file, err := os.Open(backupPath)
	if err != nil {
		return err
	}
	defer file.Close()

	gr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gr.Close()

	tr := tar.NewReader(gr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		targetPath := filepath.Join(u.installDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			os.MkdirAll(targetPath, os.FileMode(header.Mode))
		case tar.TypeReg:
			dir := filepath.Dir(targetPath)
			os.MkdirAll(dir, 0755)

			outFile, err := os.Create(targetPath)
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
			os.Chmod(targetPath, os.FileMode(header.Mode))
		}
	}

	return nil
}

func (u *Updater) updateCore(info *UpdateInfo) error {
	if u.dryRun {
		u.logger.Info("[DRY RUN] Would download and install core %s", info.LatestVersion)
		return nil
	}

	// Download package
	packagePath, err := u.downloadPackage(info.DownloadURL, info.Checksum)
	if err != nil {
		return fmt.Errorf("failed to download package: %w", err)
	}
	defer os.Remove(packagePath)

	// Extract package
	if err := u.extractPackage(packagePath, u.installDir); err != nil {
		return fmt.Errorf("failed to extract package: %w", err)
	}

	return nil
}

func (u *Updater) updateAdapter(update AdapterUpdate) error {
	if u.dryRun {
		u.logger.Info("[DRY RUN] Would download and install adapter %s %s", update.Name, update.LatestVersion)
		return nil
	}

	// Download package
	packagePath, err := u.downloadPackage(update.DownloadURL, update.Checksum)
	if err != nil {
		return fmt.Errorf("failed to download package: %w", err)
	}
	defer os.Remove(packagePath)

	// Extract package
	adapterDir := filepath.Join(u.installDir, fmt.Sprintf("%s-adapter", update.Name))
	if err := u.extractPackage(packagePath, adapterDir); err != nil {
		return fmt.Errorf("failed to extract package: %w", err)
	}

	// Save adapter version
	versionFile := filepath.Join(u.installDir, fmt.Sprintf("%s-adapter.version", update.Name))
	os.WriteFile(versionFile, []byte(update.LatestVersion), 0644)

	return nil
}

func (u *Updater) downloadPackage(url, expectedChecksum string) (string, error) {
	// Create temp file
	tempFile, err := os.CreateTemp("", "openclaw-update-*.tmp")
	if err != nil {
		return "", err
	}
	tempFile.Close()

	// Download file
	resp, err := u.httpClient.Get(url)
	if err != nil {
		os.Remove(tempFile.Name())
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		os.Remove(tempFile.Name())
		return "", fmt.Errorf("download failed with status: %s", resp.Status)
	}

	file, err := os.Create(tempFile.Name())
	if err != nil {
		os.Remove(tempFile.Name())
		return "", err
	}

	hasher := sha256.New()
	writer := io.MultiWriter(file, hasher)

	if _, err := io.Copy(writer, resp.Body); err != nil {
		file.Close()
		os.Remove(tempFile.Name())
		return "", err
	}
	file.Close()

	// Verify checksum if provided
	if expectedChecksum != "" && strings.HasPrefix(expectedChecksum, "sha256:") {
		actualChecksum := hex.EncodeToString(hasher.Sum(nil))
		expected := strings.TrimPrefix(expectedChecksum, "sha256:")
		if actualChecksum != expected {
			os.Remove(tempFile.Name())
			return "", fmt.Errorf("checksum mismatch: expected %s, got %s", expected, actualChecksum)
		}
	}

	return tempFile.Name(), nil
}

func (u *Updater) extractPackage(packagePath, targetDir string) error {
	if u.platform.IsWindows() {
		return u.extractZip(packagePath, targetDir)
	}
	return u.extractTarGz(packagePath, targetDir)
}

func (u *Updater) extractZip(zipPath, targetDir string) error {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer reader.Close()

	for _, file := range reader.File {
		path := filepath.Join(targetDir, file.Name)

		// Security: prevent path traversal
		if !strings.HasPrefix(path, filepath.Clean(targetDir)+string(os.PathSeparator)) {
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
		rc.Close()
		outFile.Close()

		if err != nil {
			return err
		}
	}

	return nil
}

func (u *Updater) extractTarGz(tarGzPath, targetDir string) error {
	file, err := os.Open(tarGzPath)
	if err != nil {
		return err
	}
	defer file.Close()

	gr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gr.Close()

	tr := tar.NewReader(gr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		path := filepath.Join(targetDir, header.Name)

		// Security: prevent path traversal
		if !strings.HasPrefix(path, filepath.Clean(targetDir)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			os.MkdirAll(path, os.FileMode(header.Mode))
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				return err
			}

			outFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}

	return nil
}
