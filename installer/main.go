package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

const (
	defaultPort = "18080"
	configFile  = "openclaw.json"
)

func main() {
	// Setup logging
	logFile, err := setupLogging()
	if err != nil {
		log.Printf("Warning: Failed to setup file logging: %v", err)
	} else {
		defer logFile.Close()
	}

	// Clean up old log files
	cleanupOldLogs()

	log.Println("OpenClaw Installer Starting...")

	// Detect platform
	platform := DetectPlatform()
	log.Printf("Detected platform: %s/%s", platform.OS, platform.Arch)

	// Create installer instance
	installer := NewInstaller(platform)

	// Define command-line flags
	var (
		installDir   = flag.String("install-dir", "", "Installation directory (default: platform default)")
		configDir    = flag.String("config-dir", "", "Configuration directory (default: platform default)")
		sourceDir    = flag.String("source", "", "Source directory (USB path)")
		adapterType  = flag.String("adapter", "usb", "Adapter type (usb/bluetooth/network)")
		serverHost   = flag.String("host", "0.0.0.0", "Server host")
		serverPort   = flag.Int("port", 8080, "Server port")
		enableTLS    = flag.Bool("tls", false, "Enable TLS")
		uninstall    = flag.Bool("uninstall", false, "Uninstall OpenClaw")
		verify       = flag.Bool("verify", false, "Verify installation")
		version      = flag.Bool("version", false, "Show version")
		help         = flag.Bool("help", false, "Show help")
	)

	flag.Parse()

	if *help {
		showHelp()
		return
	}

	if *version {
		fmt.Println("OpenClaw Installer v1.0.0")
		return
	}

	// Use platform defaults if not specified
	if *installDir == "" {
		*installDir = platform.GetInstallDir()
	}
	if *configDir == "" {
		*configDir = platform.GetConfigDir()
	}

	log.Printf("Install directory: %s", *installDir)
	log.Printf("Config directory: %s", *configDir)

	// Handle different modes
	if *uninstall {
		files := []string{platform.GetBinaryName("openclaw")}
		if err := installer.Uninstall(*installDir, files); err != nil {
			log.Fatalf("Uninstall failed: %v", err)
		}
		fmt.Println("OpenClaw uninstalled successfully!")
		return
	}

	if *verify {
		files := []string{platform.GetBinaryName("openclaw")}
		if err := installer.VerifyInstallation(*installDir, files); err != nil {
			log.Fatalf("Verification failed: %v", err)
		}
		fmt.Println("Installation verified successfully!")
		return
	}

	// Installation mode
	fmt.Println("╔════════════════════════════════════════════════════════╗")
	fmt.Println("║         OpenClaw Installer v1.0.0                      ║")
	fmt.Println("╚════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Printf("Platform: %s/%s\n", platform.OS, platform.Arch)
	fmt.Printf("Install Directory: %s\n", *installDir)
	fmt.Printf("Config Directory: %s\n", *configDir)
	fmt.Println()

	// Check if source directory is provided
	if *sourceDir == "" {
		// Try to auto-detect USB/source directory
		*sourceDir = autoDetectSourceDir()
		if *sourceDir == "" {
			fmt.Println("Warning: No source directory specified.")
			fmt.Println("Usage: openclaw-installer -source <path>")
			fmt.Println()
			fmt.Println("Attempting installation without source files...")
		}
	}

	// Install binaries
	installOpts := InstallOptions{
		SourceDir:  *sourceDir,
		InstallDir: *installDir,
		ConfigDir:  *configDir,
		BinaryName: "openclaw",
	}

	fmt.Println("Installing OpenClaw...")
	if err := installer.InstallWithOptions(installOpts); err != nil {
		log.Fatalf("Installation failed: %v", err)
	}

	// Generate configuration
	configOpts := ConfigOptions{
		Version:     "1.0.0",
		ServerHost:  *serverHost,
		ServerPort:  *serverPort,
		EnableTLS:   *enableTLS,
		AdapterType: *adapterType,
		AdapterName: "default",
	}

	config, err := GenerateConfig(configOpts)
	if err != nil {
		log.Fatalf("Config generation failed: %v", err)
	}

	// Save configuration
	configPath := filepath.Join(*configDir, configFile)
	if err := config.Save(configPath); err != nil {
		log.Fatalf("Failed to save config: %v", err)
	}

	fmt.Println()
	fmt.Println("✓ OpenClaw installed successfully!")
	fmt.Printf("✓ Configuration saved to: %s\n", configPath)
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("1. Edit the configuration file to add your API key and IM adapter settings")
	fmt.Println("2. Run 'openclaw' to start the application")
	fmt.Println()
	fmt.Println("Configuration guide:")
	fmt.Println("  - API Key: Set your AI service API key")
	fmt.Println("  - IM Adapter: Configure WeCom/DingTalk/Feishu settings")
	fmt.Println("  - Server: Configure host and port for the web config interface")
}

func showHelp() {
	fmt.Println("OpenClaw Installer - Cross-platform installation tool")
	fmt.Println()
	fmt.Println("Usage: openclaw-installer [options]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -source string     Source directory path (USB path)")
	fmt.Println("  -install-dir string Installation directory (default: platform default)")
	fmt.Println("  -config-dir string  Configuration directory (default: platform default)")
	fmt.Println("  -adapter string    Adapter type: usb, bluetooth, network (default: usb)")
	fmt.Println("  -host string       Server host (default: 0.0.0.0)")
	fmt.Println("  -port int          Server port (default: 8080)")
	fmt.Println("  -tls               Enable TLS")
	fmt.Println("  -uninstall         Uninstall OpenClaw")
	fmt.Println("  -verify            Verify installation")
	fmt.Println("  -version           Show version")
	fmt.Println("  -help              Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Install from USB drive")
	fmt.Println("  openclaw-installer -source /media/usb/OpenClaw")
	fmt.Println()
	fmt.Println("  # Install with custom directory")
	fmt.Println("  openclaw-installer -install-dir /opt/openclaw -source ./OpenClaw")
	fmt.Println()
	fmt.Println("  # Uninstall")
	fmt.Println("  openclaw-installer -uninstall")
	fmt.Println()
	fmt.Println("Platform defaults:")
	fmt.Println("  Windows: C:\\Program Files\\OpenClaw")
	fmt.Println("  macOS:   /usr/local/bin")
	fmt.Println("  Linux:   /usr/local/bin")
}

func autoDetectSourceDir() string {
	// Try common USB mount points
	possiblePaths := []string{
		"/media/OpenClaw",
		"/mnt/OpenClaw",
		"/Volumes/OpenClaw",
		"D:\\OpenClaw",
		"E:\\OpenClaw",
		"F:\\OpenClaw",
		"./OpenClaw",
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}
