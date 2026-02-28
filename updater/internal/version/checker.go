// Package version 提供版本检查和比较功能
package version

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/openclaw/updater/pkg/types"
)

// Checker 提供版本检查功能
type Checker struct {
	manifestURL string
	httpClient  *http.Client
}

// NewChecker 创建新的版本检查器
func NewChecker(manifestURL string) *Checker {
	return &Checker{
		manifestURL: manifestURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
	}

// SetHTTPClient 设置自定义 HTTP 客户端
func (c *Checker) SetHTTPClient(client *http.Client) {
	c.httpClient = client
}

// FetchRemoteManifest 获取远程发布清单
func (c *Checker) FetchRemoteManifest() (*types.ReleaseManifest, error) {
	resp, err := c.httpClient.Get(c.manifestURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch manifest: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest body: %w", err)
	}

	var manifest types.ReleaseManifest
	if err := json.Unmarshal(body, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	return &manifest, nil
}

// FetchLocalManifest 从本地路径获取发布清单
func (c *Checker) FetchLocalManifest(localPath string) (*types.ReleaseManifest, error) {
	manifestPath := filepath.Join(localPath, "manifest.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read local manifest: %w", err)
	}

	var manifest types.ReleaseManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse local manifest: %w", err)
	}

	return &manifest, nil
}

// GetInstalledVersion 获取已安装组件的版本
func (c *Checker) GetInstalledVersion(component types.ComponentType, installPath string) (string, error) {
	versionFile := filepath.Join(installPath, ".version")

	// 尝试读取版本文件
	data, err := os.ReadFile(versionFile)
	if err == nil {
		var info types.InstalledComponent
		if err := json.Unmarshal(data, &info); err == nil {
			return info.Version, nil
		}
	}

	// 回退：尝试执行二进制文件获取版本
	binaryName := c.getBinaryName(component)
	_ = filepath.Join(installPath, binaryName) // 保留以备后续使用

	// 这里可以执行 binaryPath --version 来获取版本
	// 简化处理：返回未知版本
	return "0.0.0", nil
}

// CheckForUpdates 检查可用更新
func (c *Checker) CheckForUpdates(manifest *types.ReleaseManifest, installed map[types.ComponentType]string) ([]types.UpdateInfo, error) {
	var updates []types.UpdateInfo

	// 检查核心更新
	if coreVer, ok := installed[types.ComponentCore]; ok {
		if CompareVersions(coreVer, manifest.Core.Version) < 0 {
			updates = append(updates, types.UpdateInfo{
				Component:    types.ComponentCore,
				Name:         "OpenClaw Core",
				CurrentVer:   coreVer,
				AvailableVer: manifest.Core.Version,
				DownloadURL:  c.expandURL(manifest.Core.DownloadURL),
				Checksum:     manifest.Core.Checksum,
				ReleaseNotes: manifest.Core.ReleaseNotes,
			})
		}
	}

	// 检查适配器更新
	for id, adapter := range manifest.Adapters {
		componentType := c.adapterIDToComponent(id)
		if installedVer, ok := installed[componentType]; ok {
			if CompareVersions(installedVer, adapter.Version) < 0 {
				// 检查核心版本兼容性
				if adapter.MinCoreVersion != "" {
					if coreVer, ok := installed[types.ComponentCore]; ok {
						if CompareVersions(coreVer, adapter.MinCoreVersion) < 0 {
							// 跳过不兼容的适配器更新
							continue
						}
					}
				}

				updates = append(updates, types.UpdateInfo{
					Component:    componentType,
					Name:         c.getAdapterName(id),
					CurrentVer:   installedVer,
					AvailableVer: adapter.Version,
					DownloadURL:  c.expandURL(adapter.DownloadURL),
					Checksum:     adapter.Checksum,
					ReleaseNotes: adapter.ReleaseNotes,
				})
			}
		}
	}

	return updates, nil
}

// expandURL 扩展 URL 模板变量
func (c *Checker) expandURL(url string) string {
	platform := runtime.GOOS
	arch := runtime.GOARCH
	ext := "tar.gz"
	if platform == "windows" {
		ext = "zip"
	}

	url = strings.ReplaceAll(url, "{os}", platform)
	url = strings.ReplaceAll(url, "{arch}", arch)
	url = strings.ReplaceAll(url, "{ext}", ext)

	return url
}

// getBinaryName 获取组件二进制文件名
func (c *Checker) getBinaryName(component types.ComponentType) string {
	baseName := string(component)
	if component == types.ComponentCore {
		baseName = "openclaw"
	}

	if runtime.GOOS == "windows" {
		return baseName + ".exe"
	}
	return baseName
}

// adapterIDToComponent 将适配器 ID 转换为组件类型
func (c *Checker) adapterIDToComponent(id string) types.ComponentType {
	switch id {
	case "wecom":
		return types.ComponentWecom
	case "dingtalk":
		return types.ComponentDingtalk
	case "feishu":
		return types.ComponentFeishu
	default:
		return types.ComponentType(id)
	}
}

// getAdapterName 获取适配器显示名称
func (c *Checker) getAdapterName(id string) string {
	names := map[string]string{
		"wecom":    "企业微信适配器",
		"dingtalk": "钉钉适配器",
		"feishu":   "飞书适配器",
	}
	if name, ok := names[id]; ok {
		return name
	}
	return id
}
