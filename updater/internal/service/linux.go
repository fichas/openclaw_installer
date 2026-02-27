// Package service 提供 Linux systemd 服务管理功能
package service

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

// SystemdService 提供 systemd 服务管理
type SystemdService struct {
	name        string
	displayName string
	description string
	execPath    string
	configPath  string
	workDir     string
	user        string
	group       string
	logger      Logger
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

// ServiceOptions 包含服务配置选项
type ServiceOptions struct {
	Name        string
	DisplayName string
	Description string
	ExecPath    string
	ConfigPath  string
	WorkDir     string
	User        string
	Group       string
	Logger      Logger
}

// NewSystemdService 创建新的 systemd 服务管理器
func NewSystemdService(opts ServiceOptions) *SystemdService {
	if opts.Name == "" {
		opts.Name = "openclaw"
	}
	if opts.DisplayName == "" {
		opts.DisplayName = "OpenClaw"
	}
	if opts.Description == "" {
		opts.Description = "OpenClaw - Enterprise IM Integration Platform"
	}
	if opts.ExecPath == "" {
		opts.ExecPath = "/usr/local/bin/openclaw"
	}
	if opts.ConfigPath == "" {
		opts.ConfigPath = "/etc/openclaw/openclaw.json"
	}
	if opts.WorkDir == "" {
		opts.WorkDir = "/var/lib/openclaw"
	}
	if opts.User == "" {
		opts.User = "openclaw"
	}
	if opts.Group == "" {
		opts.Group = "openclaw"
	}
	if opts.Logger == nil {
		opts.Logger = &defaultLogger{}
	}

	return &SystemdService{
		name:        opts.Name,
		displayName: opts.DisplayName,
		description: opts.Description,
		execPath:    opts.ExecPath,
		configPath:  opts.ConfigPath,
		workDir:     opts.WorkDir,
		user:        opts.User,
		group:       opts.Group,
		logger:      opts.Logger,
	}
}

// IsInstalled 检查服务是否已安装
func (s *SystemdService) IsInstalled() bool {
	servicePath := filepath.Join("/etc/systemd/system", s.name+".service")
	_, err := os.Stat(servicePath)
	return err == nil
}

// IsRunning 检查服务是否正在运行
func (s *SystemdService) IsRunning() bool {
	cmd := exec.Command("systemctl", "is-active", s.name)
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) == "active"
}

// Install 安装服务
func (s *SystemdService) Install() error {
	s.logger.Info("Installing systemd service: %s", s.name)

	// 检查是否具有 root 权限
	if os.Getuid() != 0 {
		return fmt.Errorf("service installation requires root privileges")
	}

	// 检查 systemd 是否可用
	if !s.isSystemdAvailable() {
		return fmt.Errorf("systemd is not available on this system")
	}

	// 创建用户和组
	if err := s.createUserAndGroup(); err != nil {
		s.logger.Warn("Failed to create user/group: %v", err)
	}

	// 创建工作目录
	if err := s.createWorkDir(); err != nil {
		return fmt.Errorf("failed to create work directory: %w", err)
	}

	// 生成服务文件
	serviceContent, err := s.generateServiceFile()
	if err != nil {
		return fmt.Errorf("failed to generate service file: %w", err)
	}

	// 写入服务文件
	servicePath := filepath.Join("/etc/systemd/system", s.name+".service")
	if err := os.WriteFile(servicePath, []byte(serviceContent), 0644); err != nil {
		return fmt.Errorf("failed to write service file: %w", err)
	}

	s.logger.Info("Service file written to: %s", servicePath)

	// 重新加载 systemd
	if err := s.reloadDaemon(); err != nil {
		return fmt.Errorf("failed to reload systemd: %w", err)
	}

	// 启用服务（开机自启）
	if err := s.Enable(); err != nil {
		s.logger.Warn("Failed to enable service: %v", err)
	}

	s.logger.Info("Service installed successfully")
	return nil
}

// Uninstall 卸载服务
func (s *SystemdService) Uninstall() error {
	s.logger.Info("Uninstalling systemd service: %s", s.name)

	if os.Getuid() != 0 {
		return fmt.Errorf("service uninstallation requires root privileges")
	}

	// 停止服务
	if err := s.Stop(); err != nil {
		s.logger.Warn("Failed to stop service: %v", err)
	}

	// 禁用服务
	if err := s.Disable(); err != nil {
		s.logger.Warn("Failed to disable service: %v", err)
	}

	// 删除服务文件
	servicePath := filepath.Join("/etc/systemd/system", s.name+".service")
	if err := os.Remove(servicePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove service file: %w", err)
	}

	// 重新加载 systemd
	if err := s.reloadDaemon(); err != nil {
		s.logger.Warn("Failed to reload systemd: %v", err)
	}

	s.logger.Info("Service uninstalled successfully")
	return nil
}

// Start 启动服务
func (s *SystemdService) Start() error {
	s.logger.Info("Starting service: %s", s.name)

	cmd := exec.Command("systemctl", "start", s.name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to start service: %w, output: %s", err, string(output))
	}

	// 等待服务启动
	time.Sleep(500 * time.Millisecond)

	if !s.IsRunning() {
		return fmt.Errorf("service failed to start")
	}

	s.logger.Info("Service started successfully")
	return nil
}

// Stop 停止服务
func (s *SystemdService) Stop() error {
	s.logger.Info("Stopping service: %s", s.name)

	cmd := exec.Command("systemctl", "stop", s.name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to stop service: %w, output: %s", err, string(output))
	}

	// 等待服务停止
	time.Sleep(500 * time.Millisecond)

	s.logger.Info("Service stopped successfully")
	return nil
}

// Restart 重启服务
func (s *SystemdService) Restart() error {
	s.logger.Info("Restarting service: %s", s.name)

	cmd := exec.Command("systemctl", "restart", s.name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to restart service: %w, output: %s", err, string(output))
	}

	// 等待服务启动
	time.Sleep(500 * time.Millisecond)

	if !s.IsRunning() {
		return fmt.Errorf("service failed to restart")
	}

	s.logger.Info("Service restarted successfully")
	return nil
}

// Reload 重新加载服务配置（平滑重启）
func (s *SystemdService) Reload() error {
	s.logger.Info("Reloading service: %s", s.name)

	cmd := exec.Command("systemctl", "reload", s.name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// 如果服务不支持 reload，尝试 restart
		if strings.Contains(string(output), "not loaded") {
			return s.Restart()
		}
		return fmt.Errorf("failed to reload service: %w, output: %s", err, string(output))
	}

	s.logger.Info("Service reloaded successfully")
	return nil
}

// Enable 启用服务（开机自启）
func (s *SystemdService) Enable() error {
	s.logger.Info("Enabling service: %s", s.name)

	cmd := exec.Command("systemctl", "enable", s.name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to enable service: %w, output: %s", err, string(output))
	}

	s.logger.Info("Service enabled successfully")
	return nil
}

// Disable 禁用服务
func (s *SystemdService) Disable() error {
	s.logger.Info("Disabling service: %s", s.name)

	cmd := exec.Command("systemctl", "disable", s.name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to disable service: %w, output: %s", err, string(output))
	}

	s.logger.Info("Service disabled successfully")
	return nil
}

// Status 获取服务状态
func (s *SystemdService) Status() (*ServiceStatus, error) {
	cmd := exec.Command("systemctl", "status", s.name, "--no-pager")
	output, err := cmd.CombinedOutput()

	status := &ServiceStatus{
		Name: s.name,
	}

	// 解析状态输出
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "Active:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				status.Active = strings.TrimSpace(parts[1])
				status.Running = strings.Contains(status.Active, "active (running)")
			}
		}

		if strings.HasPrefix(line, "Loaded:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				status.Loaded = strings.TrimSpace(parts[1])
				status.Enabled = strings.Contains(status.Loaded, "enabled")
			}
		}

		if strings.HasPrefix(line, "Main PID:") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				status.PID = parts[2]
			}
		}

		if strings.Contains(line, "since") && strings.Contains(line, "ago") {
			status.Uptime = extractUptime(line)
		}
	}

	// 获取内存使用
	if status.Running && status.PID != "" {
		memCmd := exec.Command("ps", "-p", status.PID, "-o", "%mem=,rss=")
		memOutput, _ := memCmd.Output()
		status.Memory = strings.TrimSpace(string(memOutput))
	}

	return status, err
}

// ServiceStatus 表示服务状态
type ServiceStatus struct {
	Name    string
	Active  string
	Loaded  string
	Running bool
	Enabled bool
	PID     string
	Uptime  string
	Memory  string
}

// String 返回格式化的状态字符串
func (s *ServiceStatus) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Service: %s\n", s.Name))
	sb.WriteString(fmt.Sprintf("Status: %s\n", s.Active))
	sb.WriteString(fmt.Sprintf("Loaded: %s\n", s.Loaded))

	if s.Running {
		sb.WriteString(fmt.Sprintf("PID: %s\n", s.PID))
		if s.Uptime != "" {
			sb.WriteString(fmt.Sprintf("Uptime: %s\n", s.Uptime))
		}
		if s.Memory != "" {
			sb.WriteString(fmt.Sprintf("Memory: %s\n", s.Memory))
		}
	}

	return sb.String()
}

// ViewLogs 查看服务日志
func (s *SystemdService) ViewLogs(lines int, follow bool) error {
	args := []string{"-u", s.name, "--no-pager"}

	if lines > 0 {
		args = append(args, "-n", fmt.Sprintf("%d", lines))
	}

	if follow {
		args = append(args, "-f")
	}

	cmd := exec.Command("journalctl", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

// GetLogs 获取日志内容
func (s *SystemdService) GetLogs(lines int) (string, error) {
	args := []string{"-u", s.name, "--no-pager", "-n", fmt.Sprintf("%d", lines)}

	cmd := exec.Command("journalctl", args...)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(output), nil
}

// ClearLogs 清除日志
func (s *SystemdService) ClearLogs() error {
	s.logger.Info("Clearing logs for service: %s", s.name)

	cmd := exec.Command("journalctl", "--rotate", "-u", s.name)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to rotate logs: %w, output: %s", err, string(output))
	}

	cmd = exec.Command("journalctl", "--vacuum-time=1s", "-u", s.name)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to vacuum logs: %w, output: %s", err, string(output))
	}

	s.logger.Info("Logs cleared successfully")
	return nil
}

// 内部辅助方法

func (s *SystemdService) isSystemdAvailable() bool {
	_, err := exec.LookPath("systemctl")
	return err == nil
}

func (s *SystemdService) reloadDaemon() error {
	cmd := exec.Command("systemctl", "daemon-reload")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("daemon-reload failed: %w, output: %s", err, string(output))
	}
	return nil
}

func (s *SystemdService) createUserAndGroup() error {
	// 检查组是否存在
	groupCmd := exec.Command("getent", "group", s.group)
	if err := groupCmd.Run(); err != nil {
		// 创建组
		s.logger.Info("Creating group: %s", s.group)
		cmd := exec.Command("groupadd", "-r", s.group)
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to create group: %w, output: %s", err, string(output))
		}
	}

	// 检查用户是否存在
	userCmd := exec.Command("id", s.user)
	if err := userCmd.Run(); err != nil {
		// 创建用户
		s.logger.Info("Creating user: %s", s.user)
		cmd := exec.Command("useradd", "-r", "-g", s.group,
			"-d", s.workDir,
			"-s", "/bin/false",
			"-c", s.displayName,
			s.user)
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to create user: %w, output: %s", err, string(output))
		}
	}

	return nil
}

func (s *SystemdService) createWorkDir() error {
	dirs := []string{
		s.workDir,
		"/var/log/openclaw",
		"/etc/openclaw",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}

		// 设置所有权
		chownCmd := exec.Command("chown", fmt.Sprintf("%s:%s", s.user, s.group), dir)
		chownCmd.Run() // 忽略错误
	}

	return nil
}

func (s *SystemdService) generateServiceFile() (string, error) {
	tmpl := `[Unit]
Description={{.Description}}
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
Environment="OPENCLAW_HOME={{.WorkDir}}"
Environment="OPENCLAW_CONFIG={{.ConfigPath}}"
Environment="OPENCLAW_LOG_LEVEL=info"
EnvironmentFile=-/etc/default/{{.Name}}
WorkingDirectory={{.WorkDir}}
ExecStart={{.ExecPath}} --config {{.ConfigPath}}
Restart=on-failure
RestartSec=5
StartLimitInterval=60s
StartLimitBurst=3
TimeoutStopSec=30
User={{.User}}
Group={{.Group}}
LimitNOFILE=65535
LimitNPROC=4096
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths={{.WorkDir}} /var/log/{{.Name}} /tmp
StandardOutput=journal
StandardError=journal
SyslogIdentifier={{.Name}}

[Install]
WantedBy=multi-user.target
`

	type templateData struct {
		Name        string
		Description string
		ExecPath    string
		ConfigPath  string
		WorkDir     string
		User        string
		Group       string
	}

	data := templateData{
		Name:        s.name,
		Description: s.description,
		ExecPath:    s.execPath,
		ConfigPath:  s.configPath,
		WorkDir:     s.workDir,
		User:        s.user,
		Group:       s.group,
	}

	t := template.Must(template.New("service").Parse(tmpl))
	var buf strings.Builder
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func extractUptime(line string) string {
	// 提取 "since Mon 2024-01-01 00:00:00 CST; 1h 30min ago" 中的时间部分
	if idx := strings.Index(line, ";"); idx != -1 {
		return strings.TrimSpace(line[idx+1:])
	}
	return ""
}

// InstallFromTemplate 从模板文件安装服务
func (s *SystemdService) InstallFromTemplate(templatePath string) error {
	s.logger.Info("Installing service from template: %s", templatePath)

	if os.Getuid() != 0 {
		return fmt.Errorf("service installation requires root privileges")
	}

	// 读取模板
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template: %w", err)
	}

	// 替换变量
	serviceContent := string(content)
	serviceContent = strings.ReplaceAll(serviceContent, "{{NAME}}", s.name)
	serviceContent = strings.ReplaceAll(serviceContent, "{{DISPLAY_NAME}}", s.displayName)
	serviceContent = strings.ReplaceAll(serviceContent, "{{DESCRIPTION}}", s.description)
	serviceContent = strings.ReplaceAll(serviceContent, "{{EXEC_PATH}}", s.execPath)
	serviceContent = strings.ReplaceAll(serviceContent, "{{CONFIG_PATH}}", s.configPath)
	serviceContent = strings.ReplaceAll(serviceContent, "{{WORK_DIR}}", s.workDir)
	serviceContent = strings.ReplaceAll(serviceContent, "{{USER}}", s.user)
	serviceContent = strings.ReplaceAll(serviceContent, "{{GROUP}}", s.group)

	// 创建用户和组
	if err := s.createUserAndGroup(); err != nil {
		s.logger.Warn("Failed to create user/group: %v", err)
	}

	// 创建工作目录
	if err := s.createWorkDir(); err != nil {
		return fmt.Errorf("failed to create work directory: %w", err)
	}

	// 写入服务文件
	servicePath := filepath.Join("/etc/systemd/system", s.name+".service")
	if err := os.WriteFile(servicePath, []byte(serviceContent), 0644); err != nil {
		return fmt.Errorf("failed to write service file: %w", err)
	}

	s.logger.Info("Service file written to: %s", servicePath)

	// 重新加载 systemd
	if err := s.reloadDaemon(); err != nil {
		return fmt.Errorf("failed to reload systemd: %w", err)
	}

	// 启用服务
	if err := s.Enable(); err != nil {
		s.logger.Warn("Failed to enable service: %v", err)
	}

	s.logger.Info("Service installed successfully")
	return nil
}

// ListServices 列出所有 OpenClaw 相关服务
func ListServices() ([]string, error) {
	cmd := exec.Command("systemctl", "list-unit-files", "--type=service", "--no-pager")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var services []string
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "openclaw") {
			fields := strings.Fields(line)
			if len(fields) > 0 {
				services = append(services, fields[0])
			}
		}
	}

	return services, scanner.Err()
}
