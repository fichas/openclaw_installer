package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

const (
	defaultVersionURL = "https://api.openclaw.io/versions/latest"
	updaterVersion    = "1.0.0"
)

func main() {
	var (
		checkOnly   = flag.Bool("check", false, "Only check for updates, don't install")
		forceUpdate = flag.Bool("force", false, "Force update even if versions match")
		dryRun      = flag.Bool("dry-run", false, "Simulate update without making changes")
		configPath  = flag.String("config", "", "Path to configuration file")
		installDir  = flag.String("install-dir", "", "Installation directory (auto-detect if not specified)")
		backupDir   = flag.String("backup-dir", "", "Backup directory (auto-detect if not specified)")
		versionURL  = flag.String("version-url", defaultVersionURL, "URL to check for version information")
		adapter     = flag.String("adapter", "all", "Adapter to update (wecom, dingtalk, feishu, or all)")
		yes         = flag.Bool("yes", false, "Auto-confirm update without prompting")
		rollback    = flag.Bool("rollback", false, "Rollback to previous version")
		listBackups = flag.Bool("list-backups", false, "List available backups")
		verbose     = flag.Bool("v", false, "Verbose output")
		help        = flag.Bool("help", false, "Show help")
	)

	flag.Parse()

	if *help {
		showHelp()
		os.Exit(0)
	}

	// Initialize logger
	logger := NewLogger(*verbose)
	logger.Info("OpenClaw Updater v%s starting...", updaterVersion)

	// Detect platform
	platform := DetectPlatform()
	logger.Info("Detected platform: %s/%s", platform.OS, platform.Arch)

	// Resolve paths
	if *installDir == "" {
		*installDir = platform.GetInstallDir()
	}
	if *backupDir == "" {
		*backupDir = platform.GetBackupDir()
	}
	if *configPath == "" {
		*configPath = platform.GetConfigPath()
	}

	logger.Info("Install directory: %s", *installDir)
	logger.Info("Backup directory: %s", *backupDir)
	logger.Info("Config file: %s", *configPath)

	// Ensure backup directory exists
	if err := os.MkdirAll(*backupDir, 0755); err != nil {
		logger.Fatal("Failed to create backup directory: %v", err)
	}

	// Load configuration
	config, err := LoadUpdateConfig(*configPath)
	if err != nil {
		logger.Warn("Could not load config: %v", err)
		config = DefaultUpdateConfig()
	}

	// Create updater instance
	updater := NewUpdater(UpdaterOptions{
		Platform:   platform,
		InstallDir: *installDir,
		BackupDir:  *backupDir,
		ConfigPath: *configPath,
		VersionURL: *versionURL,
		Adapter:    *adapter,
		DryRun:     *dryRun,
		Verbose:    *verbose,
		Logger:     logger,
	})

	// Handle rollback request
	if *rollback {
		logger.Info("Starting rollback...")
		if err := updater.Rollback(); err != nil {
			logger.Fatal("Rollback failed: %v", err)
		}
		logger.Success("Rollback completed successfully")
		os.Exit(0)
	}

	// Handle list backups request
	if *listBackups {
		backups, err := updater.ListBackups()
		if err != nil {
			logger.Fatal("Failed to list backups: %v", err)
		}
		if len(backups) == 0 {
			logger.Info("No backups found")
		} else {
			logger.Info("Available backups:")
			for _, b := range backups {
				fmt.Printf("  - %s (created: %s, size: %s)\n", b.Name, b.Created.Format("2006-01-02 15:04:05"), formatBytes(b.Size))
			}
		}
		os.Exit(0)
	}

	// Check for updates
	logger.Info("Checking for updates...")
	updateInfo, err := updater.CheckForUpdates()
	if err != nil {
		logger.Fatal("Failed to check for updates: %v", err)
	}

	// Display update information
	logger.Info("Current version: %s", updateInfo.CurrentVersion)
	logger.Info("Latest version: %s", updateInfo.LatestVersion)

	if len(updateInfo.AdapterUpdates) > 0 {
		logger.Info("Adapter updates available:")
		for _, adapter := range updateInfo.AdapterUpdates {
			fmt.Printf("  - %s: %s -> %s\n", adapter.Name, adapter.CurrentVersion, adapter.LatestVersion)
		}
	}

	// Check-only mode
	if *checkOnly {
		if updateInfo.HasUpdate() {
			logger.Info("Updates are available")
			os.Exit(0)
		} else {
			logger.Info("No updates available")
			os.Exit(1)
		}
	}

	// Check if update is needed
	if !updateInfo.HasUpdate() && !*forceUpdate {
		logger.Success("Already up to date (version %s)", updateInfo.CurrentVersion)
		os.Exit(0)
	}

	// Confirm update
	if !*yes && !*dryRun {
		if !confirmUpdate(updateInfo) {
			logger.Info("Update cancelled by user")
			os.Exit(0)
		}
	}

	// Perform update
	logger.Info("Starting update process...")
	if err := updater.Update(updateInfo); err != nil {
		logger.Error("Update failed: %v", err)
		logger.Info("Attempting rollback...")
		if rbErr := updater.Rollback(); rbErr != nil {
			logger.Error("Rollback also failed: %v", rbErr)
			logger.Fatal("System may be in an inconsistent state. Manual intervention required.")
		}
		logger.Success("Rollback completed. System restored to previous state.")
		os.Exit(1)
	}

	logger.Success("Update completed successfully!")
	logger.Info("Updated to version: %s", updateInfo.LatestVersion)

	// Cleanup old backups
	if err := updater.CleanupOldBackups(config.MaxBackups); err != nil {
		logger.Warn("Failed to cleanup old backups: %v", err)
	}
}

func showHelp() {
	fmt.Println(`OpenClaw Updater - Automatic update tool for OpenClaw

Usage: openclaw-updater [options]

Options:`)
	flag.PrintDefaults()
	fmt.Println(`
Examples:
  # Check for updates only
  openclaw-updater -check

  # Perform update with auto-confirm
  openclaw-updater -yes

  # Update specific adapter
  openclaw-updater -adapter=wecom -yes

  # Rollback to previous version
  openclaw-updater -rollback

  # List available backups
  openclaw-updater -list-backups

  # Dry run (simulate without changes)
  openclaw-updater -dry-run -yes
`)
}

func confirmUpdate(info *UpdateInfo) bool {
	fmt.Printf("\nUpdate available:\n")
	fmt.Printf("  Current: %s\n", info.CurrentVersion)
	fmt.Printf("  Latest:  %s\n", info.LatestVersion)
	if len(info.AdapterUpdates) > 0 {
		fmt.Printf("\nAdapter updates:\n")
		for _, a := range info.AdapterUpdates {
			fmt.Printf("  - %s: %s -> %s\n", a.Name, a.CurrentVersion, a.LatestVersion)
		}
	}
	fmt.Printf("\nProceed with update? [y/N]: ")
	var response string
	fmt.Scanln(&response)
	return response == "y" || response == "Y" || response == "yes"
}

func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)
	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// Logger provides logging functionality
type Logger struct {
	verbose bool
}

func NewLogger(verbose bool) *Logger {
	return &Logger{verbose: verbose}
}

func (l *Logger) Info(format string, v ...interface{}) {
	log.Printf("[INFO] "+format, v...)
}

func (l *Logger) Success(format string, v ...interface{}) {
	log.Printf("[OK] "+format, v...)
}

func (l *Logger) Warn(format string, v ...interface{}) {
	log.Printf("[WARN] "+format, v...)
}

func (l *Logger) Error(format string, v ...interface{}) {
	log.Printf("[ERROR] "+format, v...)
}

func (l *Logger) Fatal(format string, v ...interface{}) {
	log.Fatalf("[FATAL] "+format, v...)
}

func (l *Logger) Debug(format string, v ...interface{}) {
	if l.verbose {
		log.Printf("[DEBUG] "+format, v...)
	}
}
