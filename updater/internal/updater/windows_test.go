package updater

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/openclaw/updater/pkg/types"
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

func TestNewWindowsUpdater(t *testing.T) {
	logger := &mockLogger{}
	opts := WindowsUpdaterOptions{
		InstallDir:  `C:\Test\OpenClaw`,
		BackupDir:   `C:\Test\Backups`,
		ManifestURL: "https://test.openclaw.io/manifest.json",
		Logger:      logger,
	}

	updater := NewWindowsUpdater(opts)

	if updater == nil {
		t.Fatal("NewWindowsUpdater returned nil")
	}

	if updater.installDir != opts.InstallDir {
		t.Errorf("InstallDir = %s, want %s", updater.installDir, opts.InstallDir)
	}

	if updater.backupDir != opts.BackupDir {
		t.Errorf("BackupDir = %s, want %s", updater.backupDir, opts.BackupDir)
	}

	if updater.manifestURL != opts.ManifestURL {
		t.Errorf("ManifestURL = %s, want %s", updater.manifestURL, opts.ManifestURL)
	}
}

func TestNewWindowsUpdaterDefaults(t *testing.T) {
	updater := NewWindowsUpdater(WindowsUpdaterOptions{})

	if updater == nil {
		t.Fatal("NewWindowsUpdater returned nil")
	}

	if updater.installDir != `C:\Program Files\OpenClaw` {
		t.Errorf("Default InstallDir = %s, want C:\\Program Files\\OpenClaw", updater.installDir)
	}

	if updater.backupDir != `C:\ProgramData\OpenClaw\backups` {
		t.Errorf("Default BackupDir = %s, want C:\\ProgramData\\OpenClaw\\backups", updater.backupDir)
	}

	if updater.logger == nil {
		t.Error("Default Logger should not be nil")
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		v1       string
		v2       string
		expected int
	}{
		{"1.0.0", "1.0.0", 0},
		{"1.0.0", "1.0.1", -1},
		{"1.0.1", "1.0.0", 1},
		{"1.0.0", "1.1.0", -1},
		{"1.1.0", "1.0.0", 1},
		{"1.0.0", "2.0.0", -1},
		{"2.0.0", "1.0.0", 1},
		{"1.0.0", "1.0.0.1", -1},
		{"1.0.10", "1.0.2", 1},
	}

	for _, tt := range tests {
		result := compareVersions(tt.v1, tt.v2)
		if result != tt.expected {
			t.Errorf("compareVersions(%s, %s) = %d, want %d", tt.v1, tt.v2, result, tt.expected)
		}
	}
}

func TestAdapterIDToComponent(t *testing.T) {
	tests := []struct {
		id       string
		expected types.ComponentType
	}{
		{"wecom", types.ComponentWecom},
		{"dingtalk", types.ComponentDingtalk},
		{"feishu", types.ComponentFeishu},
		{"unknown", types.ComponentType("unknown")},
	}

	for _, tt := range tests {
		result := adapterIDToComponent(tt.id)
		if result != tt.expected {
			t.Errorf("adapterIDToComponent(%s) = %s, want %s", tt.id, result, tt.expected)
		}
	}
}

func TestGetAdapterName(t *testing.T) {
	tests := []struct {
		id       string
		expected string
	}{
		{"wecom", "企业微信适配器"},
		{"dingtalk", "钉钉适配器"},
		{"feishu", "飞书适配器"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		result := getAdapterName(tt.id)
		if result != tt.expected {
			t.Errorf("getAdapterName(%s) = %s, want %s", tt.id, result, tt.expected)
		}
	}
}

func TestWindowsUpdaterGetComponentPath(t *testing.T) {
	updater := NewWindowsUpdater(WindowsUpdaterOptions{
		InstallDir: `C:\Program Files\OpenClaw`,
	})

	tests := []struct {
		component types.ComponentType
		expected  string
	}{
		{types.ComponentCore, `C:\Program Files\OpenClaw`},
		{types.ComponentWecom, `C:\Program Files\OpenClaw\adapters\wecom`},
		{types.ComponentDingtalk, `C:\Program Files\OpenClaw\adapters\dingtalk`},
		{types.ComponentFeishu, `C:\Program Files\OpenClaw\adapters\feishu`},
		{types.ComponentType("custom"), `C:\Program Files\OpenClaw\custom`},
	}

	for _, tt := range tests {
		result := updater.getComponentPath(tt.component)
		normalizedResult := strings.ReplaceAll(result, "\\", "/")
		normalizedExpected := strings.ReplaceAll(tt.expected, "\\", "/")
		if normalizedResult != normalizedExpected {
			t.Errorf("getComponentPath(%s) = %s, want %s", tt.component, result, tt.expected)
		}
	}
}

func TestWindowsUpdaterExpandURL(t *testing.T) {
	updater := NewWindowsUpdater(WindowsUpdaterOptions{})

	tests := []struct {
		url      string
		expected string
	}{
		{
			"https://download.openclaw.io/core-{os}-{arch}.{ext}",
			"https://download.openclaw.io/core-windows-amd64.zip",
		},
		{
			"https://download.openclaw.io/{os}/{arch}/package.{ext}",
			"https://download.openclaw.io/windows/amd64/package.zip",
		},
	}

	for _, tt := range tests {
		result := updater.expandURL(tt.url)
		if result != tt.expected {
			t.Errorf("expandURL(%s) = %s, want %s", tt.url, result, tt.expected)
		}
	}
}

func TestDeltaPatcherCalculateFileHash(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()

	// 创建测试文件
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := []byte("Hello, World!")
	if err := os.WriteFile(testFile, testContent, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	dp := NewDeltaPatcher(tempDir, nil)

	hash, err := dp.CalculateFileHash(testFile)
	if err != nil {
		t.Fatalf("CalculateFileHash failed: %v", err)
	}

	// 预期的 SHA256 哈希
	expectedHash := "dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f"
	if hash != expectedHash {
		t.Errorf("CalculateFileHash = %s, want %s", hash, expectedHash)
	}
}

func TestDeltaPatcherGenerateManifest(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()

	// 创建测试文件结构
	files := map[string]string{
		"file1.txt": "content1",
		"dir/file2.txt": "content2",
		"dir/subdir/file3.txt": "content3",
	}

	for path, content := range files {
		fullPath := filepath.Join(tempDir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}
	}

	dp := NewDeltaPatcher(tempDir, nil)

	manifest, err := dp.GenerateManifest(tempDir)
	if err != nil {
		t.Fatalf("GenerateManifest failed: %v", err)
	}

	if len(manifest.Files) != 3 {
		t.Errorf("Expected 3 files, got %d", len(manifest.Files))
	}

	for path := range files {
		if _, exists := manifest.Files[path]; !exists {
			t.Errorf("Expected file %s in manifest", path)
		}
	}
}

func TestDeltaPatcherCalculateDelta(t *testing.T) {
	dp := NewDeltaPatcher("", nil)

	oldManifest := &DirectoryManifest{
		Files: map[string]FileEntry{
			"unchanged.txt": {Path: "unchanged.txt", Hash: "hash1", Size: 100},
			"modified.txt":  {Path: "modified.txt", Hash: "hash2", Size: 200},
			"deleted.txt":   {Path: "deleted.txt", Hash: "hash3", Size: 300},
		},
	}

	newManifest := &DirectoryManifest{
		Files: map[string]FileEntry{
			"unchanged.txt": {Path: "unchanged.txt", Hash: "hash1", Size: 100},
			"modified.txt":  {Path: "modified.txt", Hash: "hash4", Size: 250},
			"new.txt":       {Path: "new.txt", Hash: "hash5", Size: 400},
		},
	}

	delta := dp.CalculateDelta(oldManifest, newManifest)

	// 应该有一个修改的文件
	if len(delta.Patches) != 1 {
		t.Errorf("Expected 1 patch, got %d", len(delta.Patches))
	}
	if len(delta.Patches) > 0 && delta.Patches[0].Path != "modified.txt" {
		t.Errorf("Expected patch for modified.txt, got %s", delta.Patches[0].Path)
	}

	// 应该有一个新增的文件
	if len(delta.NewFiles) != 1 {
		t.Errorf("Expected 1 new file, got %d", len(delta.NewFiles))
	}
	if len(delta.NewFiles) > 0 && delta.NewFiles[0].Path != "new.txt" {
		t.Errorf("Expected new file new.txt, got %s", delta.NewFiles[0].Path)
	}

	// 应该有一个删除的文件
	if len(delta.DeletedFiles) != 1 {
		t.Errorf("Expected 1 deleted file, got %d", len(delta.DeletedFiles))
	}
	if len(delta.DeletedFiles) > 0 && delta.DeletedFiles[0] != "deleted.txt" {
		t.Errorf("Expected deleted file deleted.txt, got %s", delta.DeletedFiles[0])
	}
}

func TestRollingHash(t *testing.T) {
	rh := NewRollingHash(4)

	data := []byte("abcdefgh")

	// 计算第一个窗口的哈希
	hash1 := rh.Update(data[0])
	hash1 = rh.Update(data[1])
	hash1 = rh.Update(data[2])
	hash1 = rh.Update(data[3])

	// 滚动到下一个窗口
	hash2 := rh.Update(data[4])

	// 哈希应该不同
	if hash1 == hash2 {
		t.Error("Rolling hash should produce different values for different windows")
	}
}

func TestCalculateSavings(t *testing.T) {
	dp := NewDeltaPatcher("", nil)

	delta := &DeltaUpdateInfo{
		TotalSize: 1000,
	}

	saved, percentage := dp.CalculateSavings(delta, 5000)

	if saved != 4000 {
		t.Errorf("Expected saved = 4000, got %d", saved)
	}

	expectedPercentage := 80.0
	if percentage != expectedPercentage {
		t.Errorf("Expected percentage = %.2f, got %.2f", expectedPercentage, percentage)
	}
}

func TestDeltaPatcherCopyFile(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()

	srcFile := filepath.Join(tempDir, "source.txt")
	dstFile := filepath.Join(tempDir, "dest.txt")
	testContent := []byte("Test content for copy")

	if err := os.WriteFile(srcFile, testContent, 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	dp := NewDeltaPatcher(tempDir, nil)

	if err := dp.copyFile(srcFile, dstFile); err != nil {
		t.Fatalf("copyFile failed: %v", err)
	}

	// 验证目标文件
	content, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf("Failed to read destination file: %v", err)
	}

	if string(content) != string(testContent) {
		t.Errorf("Destination content = %s, want %s", string(content), string(testContent))
	}
}

func TestChunkedDeltaPatcherFindMatchingChunks(t *testing.T) {
	cdp := NewChunkedDeltaPatcher("", 1024, nil)

	oldChunks := []FileChunk{
		{Offset: 0, Size: 1024, Hash: "hash1"},
		{Offset: 1024, Size: 1024, Hash: "hash2"},
		{Offset: 2048, Size: 1024, Hash: "hash3"},
	}

	newChunks := []FileChunk{
		{Offset: 0, Size: 1024, Hash: "hash1"},    // 匹配
		{Offset: 1024, Size: 1024, Hash: "hash4"}, // 不匹配
		{Offset: 2048, Size: 1024, Hash: "hash2"}, // 匹配（但位置不同）
	}

	matchedOld, matchedNew := cdp.FindMatchingChunks(oldChunks, newChunks)

	// 应该找到 2 个匹配
	if len(matchedOld) != 2 {
		t.Errorf("Expected 2 matched old chunks, got %d", len(matchedOld))
	}
	if len(matchedNew) != 2 {
		t.Errorf("Expected 2 matched new chunks, got %d", len(matchedNew))
	}
}

func TestWindowsUpdaterReadSaveVersion(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()

	logger := &mockLogger{}
	updater := NewWindowsUpdater(WindowsUpdaterOptions{
		InstallDir: tempDir,
		Logger:     logger,
	})

	// 测试读取不存在的版本文件
	version := updater.readVersion(types.ComponentCore)
	if version != "0.0.0" {
		t.Errorf("Expected version 0.0.0 for non-existent file, got %s", version)
	}

	// 保存版本
	testVersion := "1.2.3"
	if err := updater.saveVersion(types.ComponentCore, testVersion); err != nil {
		t.Fatalf("saveVersion failed: %v", err)
	}

	// 读取版本
	version = updater.readVersion(types.ComponentCore)
	if version != testVersion {
		t.Errorf("Expected version %s, got %s", testVersion, version)
	}
}

func TestContains(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}

	if !contains(slice, 3) {
		t.Error("Expected contains(slice, 3) = true")
	}

	if contains(slice, 10) {
		t.Error("Expected contains(slice, 10) = false")
	}
}

func TestSortDeltaPatches(t *testing.T) {
	dp := NewDeltaPatcher("", nil)

	delta := &DeltaUpdateInfo{
		Patches: []FilePatch{
			{Path: "z/file.txt"},
			{Path: "a/file.txt"},
			{Path: "m/file.txt"},
		},
		NewFiles: []FileInfo{
			{Path: "z/new.txt"},
			{Path: "a/new.txt"},
		},
		DeletedFiles: []string{"z/old.txt", "a/old.txt"},
	}

	dp.SortDeltaPatches(delta)

	// 验证排序结果
	if delta.Patches[0].Path != "a/file.txt" {
		t.Errorf("Expected first patch path = a/file.txt, got %s", delta.Patches[0].Path)
	}
	if delta.NewFiles[0].Path != "a/new.txt" {
		t.Errorf("Expected first new file path = a/new.txt, got %s", delta.NewFiles[0].Path)
	}
	if delta.DeletedFiles[0] != "a/old.txt" {
		t.Errorf("Expected first deleted file = a/old.txt, got %s", delta.DeletedFiles[0])
	}
}

func BenchmarkCompareVersions(b *testing.B) {
	for i := 0; i < b.N; i++ {
		compareVersions("1.2.3", "1.2.4")
	}
}

func BenchmarkCalculateFileHash(b *testing.B) {
	// 创建临时目录
	tempDir := b.TempDir()

	// 创建测试文件
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := make([]byte, 1024*1024) // 1MB
	os.WriteFile(testFile, testContent, 0644)

	dp := NewDeltaPatcher(tempDir, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dp.CalculateFileHash(testFile)
	}
}

func TestWindowsUpdaterScheduleUpdate(t *testing.T) {
	logger := &mockLogger{}
	updater := NewWindowsUpdater(WindowsUpdaterOptions{
		Logger: logger,
	})

	// 测试调度功能不会 panic
	updater.ScheduleUpdate(100 * time.Millisecond)

	// 等待一段时间让调度器运行
	time.Sleep(200 * time.Millisecond)
}
