// Windows 自动更新服务命令行工具
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/openclaw/updater/internal/updater"
)

const defaultManifestURL = "https://download.openclaw.io/manifest.json"

func main() {
	var (
		install   = flag.Bool("install", false, "Install the updater service")
		uninstall = flag.Bool("uninstall", false, "Uninstall the updater service")
		start     = flag.Bool("start", false, "Start the updater service")
		stop      = flag.Bool("stop", false, "Stop the updater service")
		status    = flag.Bool("status", false, "Get service status")
		check     = flag.Bool("check", false, "Check for updates once")
		update    = flag.Bool("update", false, "Check and install updates once")
		service   = flag.Bool("service", false, "Run as Windows service (internal use)")
		manifest  = flag.String("manifest", defaultManifestURL, "Manifest URL")
		interval  = flag.Duration("interval", 24*time.Hour, "Update check interval")
	)
	flag.Parse()

	// 确保数据目录存在
	if err := updater.EnsureServiceDataDir(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// 处理服务安装/卸载命令
	if *install {
		if err := updater.InstallService(*manifest); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to install service: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Service installed successfully")
		return
	}

	if *uninstall {
		if err := updater.UninstallService(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to uninstall service: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Service uninstalled successfully")
		return
	}

	if *start {
		if err := updater.StartService(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to start service: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Service started successfully")
		return
	}

	if *stop {
		if err := updater.StopService(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to stop service: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Service stopped successfully")
		return
	}

	if *status {
		state, err := updater.GetServiceStatus()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get service status: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Service status: %s\n", stateToString(state))
		return
	}

	// 创建更新器实例
	opts := updater.WindowsUpdaterOptions{
		ManifestURL: *manifest,
	}

	// 如果在服务上下文中运行
	if *service || updater.IsWindowsService() {
		svcLogger, err := updater.NewServiceLogger()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create service logger: %v\n", err)
			os.Exit(1)
		}
		defer svcLogger.Close()

		opts.Logger = svcLogger
		u := updater.NewWindowsUpdater(opts)
		u.ScheduleUpdate(*interval)

		if err := updater.RunService(u); err != nil {
			svcLogger.Error("Service error: %v", err)
			os.Exit(1)
		}
		return
	}

	// 创建带控制台的更新器
	u := updater.NewWindowsUpdater(opts)

	// 检查更新
	if *check {
		updates, err := u.CheckForUpdates()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to check for updates: %v\n", err)
			os.Exit(1)
		}

		if len(updates) == 0 {
			fmt.Println("No updates available")
			return
		}

		fmt.Printf("Found %d update(s):\n", len(updates))
		for _, update := range updates {
			fmt.Printf("  - %s: %s -> %s\n", update.Name, update.CurrentVer, update.AvailableVer)
			if update.ReleaseNotes != "" {
				fmt.Printf("    Release notes: %s\n", update.ReleaseNotes)
			}
		}
		return
	}

	// 执行更新
	if *update {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()

		if err := u.RunAutoUpdate(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "Update failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Update check completed")
		return
	}

	// 默认行为：显示帮助
	flag.Usage()
}

// stateToString 将服务状态转换为字符串
func stateToString(state int) string {
	states := map[int]string{
		1:  "Stopped",
		2:  "Start Pending",
		3:  "Stop Pending",
		4:  "Running",
		5:  "Continue Pending",
		6:  "Pause Pending",
		7:  "Paused",
	}
	if s, ok := states[state]; ok {
		return s
	}
	return fmt.Sprintf("Unknown (%d)", state)
}
