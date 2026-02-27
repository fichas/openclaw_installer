package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// setupLogging initializes logging to file (and optionally console)
func setupLogging() (*os.File, error) {
	// Determine log directory based on platform
	logDir := getLogDir()

	// Create log directory if it doesn't exist
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Create log file with timestamp
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	logFileName := fmt.Sprintf("openclaw-installer_%s.log", timestamp)
	logPath := filepath.Join(logDir, logFileName)

	// Open log file
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	// Set up multi-writer for both file and console (if not Windows GUI mode)
	if runtime.GOOS == "windows" && isWindowsGUI() {
		// Windows GUI mode: log only to file
		log.SetOutput(logFile)
	} else {
		// Console mode: log to both file and stdout
		log.SetOutput(io.MultiWriter(os.Stdout, logFile))
	}

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	log.Printf("Logging initialized. Log file: %s", logPath)
	return logFile, nil
}

// getLogDir returns the appropriate log directory for the current platform
func getLogDir() string {
	switch runtime.GOOS {
	case "windows":
		// Use %LOCALAPPDATA% or fallback to temp directory
		if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
			return filepath.Join(localAppData, "OpenClaw", "Logs")
		}
		return filepath.Join(os.TempDir(), "OpenClaw", "Logs")
	case "darwin":
		// macOS: ~/Library/Logs/OpenClaw
		home, err := os.UserHomeDir()
		if err != nil {
			return filepath.Join(os.TempDir(), "OpenClaw", "Logs")
		}
		return filepath.Join(home, "Library", "Logs", "OpenClaw")
	default:
		// Linux: ~/.local/share/OpenClaw/logs or /tmp
		home, err := os.UserHomeDir()
		if err != nil {
			return filepath.Join(os.TempDir(), "OpenClaw", "Logs")
		}
		return filepath.Join(home, ".local", "share", "OpenClaw", "logs")
	}
}

// cleanupOldLogs removes log files older than 7 days
func cleanupOldLogs() {
	logDir := getLogDir()
	cutoff := time.Now().AddDate(0, 0, -7)

	entries, err := os.ReadDir(logDir)
	if err != nil {
		return // Silently ignore errors
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			os.Remove(filepath.Join(logDir, entry.Name()))
		}
	}
}
