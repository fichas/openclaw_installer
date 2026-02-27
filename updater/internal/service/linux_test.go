package service

import (
	"os"
	"testing"
)

// mockLogger 用于测试的 mock 日志记录器
type mockLogger struct {
	infos  []string
	warns  []string
	errors []string
}

func (m *mockLogger) Info(format string, args ...interface{}) {
	m.infos = append(m.infos, format)
}

func (m *mockLogger) Warn(format string, args ...interface{}) {
	m.warns = append(m.warns, format)
}

func (m *mockLogger) Error(format string, args ...interface{}) {
	m.errors = append(m.errors, format)
}

func TestNewSystemdService(t *testing.T) {
	logger := &mockLogger{}
	opts := ServiceOptions{
		Name:        "test-service",
		DisplayName: "Test Service",
		Description: "A test service",
		ExecPath:    "/usr/bin/test",
		ConfigPath:  "/etc/test/config.json",
		WorkDir:     "/var/lib/test",
		User:        "testuser",
		Group:       "testgroup",
		Logger:      logger,
	}

	svc := NewSystemdService(opts)

	if svc == nil {
		t.Fatal("NewSystemdService returned nil")
	}

	if svc.name != opts.Name {
		t.Errorf("Name = %s, want %s", svc.name, opts.Name)
	}

	if svc.displayName != opts.DisplayName {
		t.Errorf("DisplayName = %s, want %s", svc.displayName, opts.DisplayName)
	}

	if svc.execPath != opts.ExecPath {
		t.Errorf("ExecPath = %s, want %s", svc.execPath, opts.ExecPath)
	}
}

func TestNewSystemdServiceDefaults(t *testing.T) {
	svc := NewSystemdService(ServiceOptions{})

	if svc == nil {
		t.Fatal("NewSystemdService returned nil")
	}

	if svc.name != "openclaw" {
		t.Errorf("Default Name = %s, want openclaw", svc.name)
	}

	if svc.execPath != "/usr/local/bin/openclaw" {
		t.Errorf("Default ExecPath = %s, want /usr/local/bin/openclaw", svc.execPath)
	}

	if svc.user != "openclaw" {
		t.Errorf("Default User = %s, want openclaw", svc.user)
	}

	if svc.logger == nil {
		t.Error("Default Logger should not be nil")
	}
}

func TestSystemdServiceIsInstalled(t *testing.T) {
	// 此测试需要 root 权限，跳过
	if os.Getuid() != 0 {
		t.Skip("Skipping test that requires root privileges")
	}

	svc := NewSystemdService(ServiceOptions{})

	// 测试未安装的服务
	installed := svc.IsInstalled()
	// 结果取决于系统状态，不断言具体值
	t.Logf("IsInstalled = %v", installed)
}

func TestSystemdServiceIsRunning(t *testing.T) {
	// 此测试需要 systemd，跳过
	if _, err := os.Stat("/run/systemd/system"); os.IsNotExist(err) {
		t.Skip("Skipping test - systemd not available")
	}

	svc := NewSystemdService(ServiceOptions{})

	// 测试未安装的服务状态
	running := svc.IsRunning()
	// 结果取决于系统状态
	t.Logf("IsRunning = %v", running)
}

func TestExtractUptime(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			"Active: active (running) since Mon 2024-01-01 10:00:00 CST; 2h 30min ago",
			"2h 30min ago",
		},
		{
			"Active: active (running) since Mon 2024-01-01 10:00:00 CST; 5min ago",
			"5min ago",
		},
		{
			"Active: inactive (dead) since Mon 2024-01-01 10:00:00 CST",
			"",
		},
	}

	for _, tt := range tests {
		result := extractUptime(tt.input)
		if result != tt.expected {
			t.Errorf("extractUptime(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestServiceStatusString(t *testing.T) {
	status := &ServiceStatus{
		Name:    "openclaw",
		Active:  "active (running) since Mon 2024-01-01 10:00:00 CST",
		Loaded:  "loaded (/etc/systemd/system/openclaw.service; enabled)",
		Running: true,
		Enabled: true,
		PID:     "1234",
		Uptime:  "2h 30min ago",
		Memory:  "1.5% 24560",
	}

	str := status.String()

	if str == "" {
		t.Error("ServiceStatus.String() returned empty string")
	}

	if !contains(str, "openclaw") {
		t.Error("String should contain service name")
	}

	if !contains(str, "1234") {
		t.Error("String should contain PID")
	}
}

func TestGenerateServiceFile(t *testing.T) {
	svc := NewSystemdService(ServiceOptions{
		Name:        "test",
		Description: "Test Service",
		ExecPath:    "/usr/bin/test",
		ConfigPath:  "/etc/test.json",
		WorkDir:     "/var/lib/test",
		User:        "testuser",
		Group:       "testgroup",
	})

	content, err := svc.generateServiceFile()
	if err != nil {
		t.Fatalf("generateServiceFile failed: %v", err)
	}

	// 验证内容包含关键配置
	checks := []string{
		"[Unit]",
		"Description=Test Service",
		"[Service]",
		"ExecStart=/usr/bin/test",
		"User=testuser",
		"Group=testgroup",
		"WorkingDirectory=/var/lib/test",
		"[Install]",
	}

	for _, check := range checks {
		if !contains(content, check) {
			t.Errorf("Generated service file missing: %s", check)
		}
	}
}

func TestIsSystemdAvailable(t *testing.T) {
	svc := NewSystemdService(ServiceOptions{})

	available := svc.isSystemdAvailable()

	// 在 Linux 系统上应该可用
	if os.Getenv("CI") == "" {
		t.Logf("systemd available: %v", available)
	}
}

func TestCreateWorkDir(t *testing.T) {
	// 此测试需要 root 权限
	if os.Getuid() != 0 {
		t.Skip("Skipping test that requires root privileges")
	}

	tempDir := t.TempDir()
	svc := NewSystemdService(ServiceOptions{
		WorkDir: tempDir,
	})

	err := svc.createWorkDir()
	if err != nil {
		t.Errorf("createWorkDir failed: %v", err)
	}

	// 验证目录创建
	dirs := []string{
		tempDir,
		"/var/log/openclaw",
		"/etc/openclaw",
	}

	for _, dir := range dirs {
		if dir == tempDir {
			if _, err := os.Stat(dir); err != nil {
				t.Errorf("Directory not created: %s", dir)
			}
		}
	}
}

// 辅助函数
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
