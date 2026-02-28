// Package download 提供文件下载功能
package download

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/openclaw/updater/pkg/types"
)

// Downloader 提供下载功能
type Downloader struct {
	httpClient *http.Client
	proxyURL   *url.URL
	progressCh chan<- types.DownloadProgress
}

// Option 下载器配置选项
type Option func(*Downloader)

// WithHTTPClient 设置自定义 HTTP 客户端
func WithHTTPClient(client *http.Client) Option {
	return func(d *Downloader) {
		d.httpClient = client
	}
}

// WithProxy 设置代理
func WithProxy(proxyURL string) Option {
	return func(d *Downloader) {
		if proxyURL != "" {
			u, err := url.Parse(proxyURL)
			if err == nil {
				d.proxyURL = u
			}
		}
	}
}

// WithProgressChannel 设置进度通知通道
func WithProgressChannel(ch chan<- types.DownloadProgress) Option {
	return func(d *Downloader) {
		d.progressCh = ch
	}
}

// NewDownloader 创建新的下载器
func NewDownloader(opts ...Option) *Downloader {
	d := &Downloader{
		httpClient: &http.Client{
			Timeout: 5 * time.Minute,
		},
	}

	for _, opt := range opts {
		opt(d)
	}

	return d
}

// Download 下载文件
func (d *Downloader) Download(url, destPath string, component types.ComponentType) error {
	// 创建目标目录
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// 创建临时文件
	tempFile := destPath + ".tmp"
	file, err := os.Create(tempFile)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile) // 清理临时文件
	defer file.Close()

	// 发送请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// 设置代理
	if d.proxyURL != nil {
		d.httpClient.Transport = &http.Transport{
			Proxy: http.ProxyURL(d.proxyURL),
		}
	}

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// 获取文件大小
	totalSize := resp.ContentLength

	// 复制数据并报告进度
	writer := &progressWriter{
		writer:     file,
		total:      totalSize,
		component:  component,
		url:        url,
		progressCh: d.progressCh,
		startTime:  time.Now(),
	}

	_, err = io.Copy(writer, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	// 关闭文件
	file.Close()

	// 重命名为最终文件名
	if err := os.Rename(tempFile, destPath); err != nil {
		return fmt.Errorf("failed to rename file: %w", err)
	}

	return nil
}

// DownloadAndVerify 下载文件并验证校验和
func (d *Downloader) DownloadAndVerify(url, destPath, expectedChecksum string, component types.ComponentType) error {
	// 下载文件
	if err := d.Download(url, destPath, component); err != nil {
		return err
	}

	// 验证校验和
	if expectedChecksum != "" {
		if err := d.verifyChecksum(destPath, expectedChecksum); err != nil {
			os.Remove(destPath)
			return fmt.Errorf("checksum verification failed: %w", err)
		}
	}

	return nil
}

// verifyChecksum 验证文件校验和
func (d *Downloader) verifyChecksum(filePath, expectedChecksum string) error {
	// 支持 sha256: 前缀
	if len(expectedChecksum) > 7 && expectedChecksum[:7] == "sha256:" {
		expectedChecksum = expectedChecksum[7:]
	}

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return err
	}

	actualChecksum := hex.EncodeToString(hash.Sum(nil))
	if actualChecksum != expectedChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
	}

	return nil
}

// DownloadComponent 下载组件
func (d *Downloader) DownloadComponent(updateInfo types.UpdateInfo, destDir string) (string, error) {
	filename := filepath.Base(updateInfo.DownloadURL)
	if filename == "" {
		filename = string(updateInfo.Component) + ".zip"
	}

	destPath := filepath.Join(destDir, filename)

	if err := d.DownloadAndVerify(
		updateInfo.DownloadURL,
		destPath,
		updateInfo.Checksum,
		updateInfo.Component,
	); err != nil {
		return "", err
	}

	return destPath, nil
}

// progressWriter 带进度报告的写入器
type progressWriter struct {
	writer     io.Writer
	downloaded int64
	total      int64
	component  types.ComponentType
	url        string
	progressCh chan<- types.DownloadProgress
	startTime  time.Time
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	n, err := pw.writer.Write(p)
	pw.downloaded += int64(n)

	if pw.progressCh != nil {
		percentage := 0.0
		if pw.total > 0 {
			percentage = float64(pw.downloaded) / float64(pw.total) * 100
		}

		elapsed := time.Since(pw.startTime).Seconds()
		speed := 0.0
		if elapsed > 0 {
			speed = float64(pw.downloaded) / elapsed
		}

		pw.progressCh <- types.DownloadProgress{
			Component:  pw.component,
			URL:        pw.url,
			TotalBytes: pw.total,
			Downloaded: pw.downloaded,
			Percentage: percentage,
			Speed:      speed,
		}
	}

	return n, err
}

// DownloadMultiple 并发下载多个组件
func (d *Downloader) DownloadMultiple(updates []types.UpdateInfo, destDir string) (map[types.ComponentType]string, error) {
	results := make(map[types.ComponentType]string)
	var mu sync.Mutex
	var wg sync.WaitGroup
	errors := make(chan error, len(updates))

	for _, update := range updates {
		wg.Add(1)
		go func(u types.UpdateInfo) {
			defer wg.Done()
			path, err := d.DownloadComponent(u, destDir)
			if err != nil {
				errors <- fmt.Errorf("failed to download %s: %w", u.Component, err)
				return
			}
			mu.Lock()
			results[u.Component] = path
			mu.Unlock()
			errors <- nil
		}(update)
	}

	wg.Wait()
	close(errors)

	// 收集下载结果
	var errs []error
	for err := range errors {
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return results, fmt.Errorf("download errors: %v", errs)
	}

	return results, nil
}
