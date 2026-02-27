//go:build windows
// +build windows

package updater

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"
)

const (
	serviceName        = "OpenClawUpdater"
	serviceDisplayName = "OpenClaw Auto Updater"
	serviceDescription = "Automatically checks for and installs OpenClaw updates"
	defaultCheckInterval = 24 * time.Hour
)

// WindowsService 实现 Windows 服务接口
type WindowsService struct {
	updater  *WindowsUpdater
	interval time.Duration
	elog     debug.Log
}

// NewWindowsService 创建新的 Windows 服务实例
func NewWindowsService(updater *WindowsUpdater) *WindowsService {
	return &WindowsService{
		updater:  updater,
		interval: defaultCheckInterval,
	}
}

// Execute 实现 svc.Handler 接口
func (s *WindowsService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown
	changes <- svc.Status{State: svc.StartPending}

	// 创建上下文和取消函数
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 启动更新检查循环
	go s.runUpdateLoop(ctx)

	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}

	if s.elog != nil {
		s.elog.Info(1, "OpenClaw Updater service started")
	}

loop:
	for {
		select {
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				if s.elog != nil {
					s.elog.Info(1, "OpenClaw Updater service stopping")
				}
				break loop
			default:
				if s.elog != nil {
					s.elog.Error(1, fmt.Sprintf("unexpected control request #%d", c))
				}
			}
		}
	}

	changes <- svc.Status{State: svc.StopPending}
	cancel()
	time.Sleep(100 * time.Millisecond)
	changes <- svc.Status{State: svc.Stopped}

	return
}

// runUpdateLoop 运行更新检查循环
func (s *WindowsService) runUpdateLoop(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	// 立即执行一次检查
	s.checkAndUpdate(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.checkAndUpdate(ctx)
		}
	}
}

// checkAndUpdate 检查并执行更新
func (s *WindowsService) checkAndUpdate(ctx context.Context) {
	if s.elog != nil {
		s.elog.Info(1, "Checking for updates...")
	}

	updateCtx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()

	if err := s.updater.RunAutoUpdate(updateCtx); err != nil {
		if s.elog != nil {
			s.elog.Error(1, fmt.Sprintf("Auto update failed: %v", err))
		}
	} else {
		if s.elog != nil {
			s.elog.Info(1, "Auto update check completed")
		}
	}
}

// RunService 运行服务
func RunService(updater *WindowsUpdater) error {
	s := NewWindowsService(updater)

	// 尝试打开事件日志
	elog, err := eventlog.Open(serviceName)
	if err == nil {
		s.elog = elog
	}

	return svc.Run(serviceName, s)
}

// IsWindowsService 检查当前是否在 Windows 服务上下文中运行
func IsWindowsService() bool {
	isService, err := svc.IsWindowsService()
	if err != nil {
		return false
	}
	return isService
}

// InstallService 安装 Windows 服务
func InstallService(manifestURL string) error {
	// 获取当前可执行文件路径
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// 连接到服务管理器
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer m.Disconnect()

	// 检查服务是否已存在
	service, err := m.OpenService(serviceName)
	if err == nil {
		service.Close()
		return fmt.Errorf("service %s already exists", serviceName)
	}

	// 创建服务配置
	config := mgr.Config{
		DisplayName: serviceDisplayName,
		Description: serviceDescription,
		StartType:   mgr.StartAutomatic,
	}

	// 创建服务
	service, err = m.CreateService(
		serviceName,
		exePath,
		config,
		"service",
		manifestURL,
	)
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}
	defer service.Close()

	// 设置失败恢复选项
	recoveryActions := []mgr.RecoveryAction{
		{Type: mgr.ServiceRestart, Delay: 60 * time.Second},
		{Type: mgr.ServiceRestart, Delay: 120 * time.Second},
		{Type: mgr.NoAction},
	}
	service.SetRecoveryActions(recoveryActions, 86400) // 24小时内重置计数

	// 创建事件日志源
	err = eventlog.InstallAsEventCreate(serviceName, eventlog.Error|eventlog.Warning|eventlog.Info)
	if err != nil {
		// 非致命错误，继续
		log.Printf("Warning: failed to install event log source: %v", err)
	}

	return nil
}

// UninstallService 卸载 Windows 服务
func UninstallService() error {
	// 连接到服务管理器
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer m.Disconnect()

	// 打开服务
	service, err := m.OpenService(serviceName)
	if err != nil {
		return fmt.Errorf("service %s not found: %w", serviceName, err)
	}
	defer service.Close()

	// 停止服务（如果正在运行）
	status, err := service.Control(svc.Stop)
	if err == nil {
		// 等待服务停止
		timeout := time.After(30 * time.Second)
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

	waitLoop:
		for {
			select {
			case <-timeout:
				break waitLoop
			case <-ticker.C:
				status, _ = service.Query()
				if status.State == svc.Stopped {
					break waitLoop
				}
			}
		}
	}

	// 删除服务
	if err := service.Delete(); err != nil {
		return fmt.Errorf("failed to delete service: %w", err)
	}

	// 删除事件日志源
	eventlog.Remove(serviceName)

	return nil
}

// StartService 启动服务
func StartService() error {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer m.Disconnect()

	service, err := m.OpenService(serviceName)
	if err != nil {
		return fmt.Errorf("service %s not found: %w", serviceName, err)
	}
	defer service.Close()

	if err := service.Start(); err != nil {
		return fmt.Errorf("failed to start service: %w", err)
	}

	return nil
}

// StopService 停止服务
func StopService() error {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer m.Disconnect()

	service, err := m.OpenService(serviceName)
	if err != nil {
		return fmt.Errorf("service %s not found: %w", serviceName, err)
	}
	defer service.Close()

	_, err = service.Control(svc.Stop)
	if err != nil {
		return fmt.Errorf("failed to stop service: %w", err)
	}

	return nil
}

// GetServiceStatus 获取服务状态
func GetServiceStatus() (svc.State, error) {
	m, err := mgr.Connect()
	if err != nil {
		return 0, fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer m.Disconnect()

	service, err := m.OpenService(serviceName)
	if err != nil {
		return 0, fmt.Errorf("service %s not found: %w", serviceName, err)
	}
	defer service.Close()

	status, err := service.Query()
	if err != nil {
		return 0, fmt.Errorf("failed to query service status: %w", err)
	}

	return status.State, nil
}

// UpdateServiceConfig 更新服务配置
func UpdateServiceConfig(interval time.Duration) error {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer m.Disconnect()

	service, err := m.OpenService(serviceName)
	if err != nil {
		return fmt.Errorf("service %s not found: %w", serviceName, err)
	}
	defer service.Close()

	// 更新服务配置
	config := mgr.Config{
		DisplayName: serviceDisplayName,
		Description: fmt.Sprintf("%s (Check interval: %v)", serviceDescription, interval),
		StartType:   mgr.StartAutomatic,
	}

	if err := service.UpdateConfig(config); err != nil {
		return fmt.Errorf("failed to update service config: %w", err)
	}

	return nil
}

// ServiceLogger 实现 Windows 事件日志记录器
type ServiceLogger struct {
	elog *eventlog.Log
}

// NewServiceLogger 创建新的服务日志记录器
func NewServiceLogger() (*ServiceLogger, error) {
	elog, err := eventlog.Open(serviceName)
	if err != nil {
		return nil, err
	}
	return &ServiceLogger{elog: elog}, nil
}

// Info 记录信息日志
func (l *ServiceLogger) Info(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.elog.Info(1, msg)
}

// Warn 记录警告日志
func (l *ServiceLogger) Warn(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.elog.Warning(1, msg)
}

// Error 记录错误日志
func (l *ServiceLogger) Error(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.elog.Error(1, msg)
}

// Close 关闭日志记录器
func (l *ServiceLogger) Close() error {
	return l.elog.Close()
}

// EnsureServiceDataDir 确保服务数据目录存在
func EnsureServiceDataDir() error {
	dataDir := `C:\ProgramData\OpenClaw`
	dirs := []string{
		dataDir,
		filepath.Join(dataDir, "logs"),
		filepath.Join(dataDir, "backups"),
		filepath.Join(dataDir, "updates"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}
