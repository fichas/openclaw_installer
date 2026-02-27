// Package updater 提供差异更新（增量更新）功能
package updater

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/openclaw/updater/pkg/types"
)

// DeltaUpdateInfo 差异更新信息
type DeltaUpdateInfo struct {
	BaseVersion   string            `json:"base_version"`
	TargetVersion string            `json:"target_version"`
	Component     types.ComponentType `json:"component"`
	Patches       []FilePatch       `json:"patches"`
	NewFiles      []FileInfo        `json:"new_files"`
	DeletedFiles  []string          `json:"deleted_files"`
	TotalSize     int64             `json:"total_size"`
}

// FilePatch 文件补丁信息
type FilePatch struct {
	Path          string `json:"path"`
	OldHash       string `json:"old_hash"`
	NewHash       string `json:"new_hash"`
	PatchURL      string `json:"patch_url"`
	PatchSize     int64  `json:"patch_size"`
	OldSize       int64  `json:"old_size"`
	NewSize       int64  `json:"new_size"`
	PatchHash     string `json:"patch_hash"`
}

// FileInfo 文件信息
type FileInfo struct {
	Path     string `json:"path"`
	Hash     string `json:"hash"`
	Size     int64  `json:"size"`
	URL      string `json:"url"`
	FileHash string `json:"file_hash"`
}

// DeltaPatcher 差异更新补丁应用器
type DeltaPatcher struct {
	tempDir string
	logger  Logger
}

// NewDeltaPatcher 创建新的差异更新补丁应用器
func NewDeltaPatcher(tempDir string, logger Logger) *DeltaPatcher {
	if logger == nil {
		logger = &defaultLogger{}
	}
	return &DeltaPatcher{
		tempDir: tempDir,
		logger:  logger,
	}
}

// CalculateFileHash 计算文件哈希
func (dp *DeltaPatcher) CalculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// GenerateManifest 生成目录清单
func (dp *DeltaPatcher) GenerateManifest(dirPath string) (*DirectoryManifest, error) {
	manifest := &DirectoryManifest{
		Files: make(map[string]FileEntry),
	}

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(dirPath, path)
		if err != nil {
			return err
		}

		hash, err := dp.CalculateFileHash(path)
		if err != nil {
			return fmt.Errorf("failed to hash file %s: %w", path, err)
		}

		manifest.Files[relPath] = FileEntry{
			Path: relPath,
			Hash: hash,
			Size: info.Size(),
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return manifest, nil
}

// DirectoryManifest 目录清单
type DirectoryManifest struct {
	Files map[string]FileEntry `json:"files"`
}

// FileEntry 文件条目
type FileEntry struct {
	Path string `json:"path"`
	Hash string `json:"hash"`
	Size int64  `json:"size"`
}

// CalculateDelta 计算两个版本之间的差异
func (dp *DeltaPatcher) CalculateDelta(oldManifest, newManifest *DirectoryManifest) *DeltaUpdateInfo {
	delta := &DeltaUpdateInfo{
		Patches:      []FilePatch{},
		NewFiles:     []FileInfo{},
		DeletedFiles: []string{},
	}

	// 查找修改和新增的文件
	for path, newEntry := range newManifest.Files {
		oldEntry, exists := oldManifest.Files[path]
		if !exists {
			// 新增文件
			delta.NewFiles = append(delta.NewFiles, FileInfo{
				Path: path,
				Hash: newEntry.Hash,
				Size: newEntry.Size,
			})
			delta.TotalSize += newEntry.Size
		} else if oldEntry.Hash != newEntry.Hash {
			// 修改的文件
			delta.Patches = append(delta.Patches, FilePatch{
				Path:    path,
				OldHash: oldEntry.Hash,
				NewHash: newEntry.Hash,
				OldSize: oldEntry.Size,
				NewSize: newEntry.Size,
			})
			// 估算补丁大小（通常比完整文件小得多）
			estimatedPatchSize := estimatePatchSize(oldEntry.Size, newEntry.Size)
			delta.TotalSize += estimatedPatchSize
		}
	}

	// 查找删除的文件
	for path := range oldManifest.Files {
		if _, exists := newManifest.Files[path]; !exists {
			delta.DeletedFiles = append(delta.DeletedFiles, path)
		}
	}

	return delta
}

// estimatePatchSize 估算补丁大小
func estimatePatchSize(oldSize, newSize int64) int64 {
	// 简单的估算：假设二进制差异约为文件大小的 10-30%
	avgSize := (oldSize + newSize) / 2
	return avgSize / 5 // 20% 估算
}

// ApplyDelta 应用差异更新
func (dp *DeltaPatcher) ApplyDelta(delta *DeltaUpdateInfo, sourceDir, targetDir string) error {
	dp.logger.Info("Applying delta update from %s to %s", delta.BaseVersion, delta.TargetVersion)

	// 确保目标目录存在
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// 1. 复制未修改的文件
	if err := dp.copyUnchangedFiles(delta, sourceDir, targetDir); err != nil {
		return fmt.Errorf("failed to copy unchanged files: %w", err)
	}

	// 2. 应用补丁
	for _, patch := range delta.Patches {
		if err := dp.applyPatch(patch, sourceDir, targetDir); err != nil {
			return fmt.Errorf("failed to apply patch for %s: %w", patch.Path, err)
		}
	}

	// 3. 下载新文件
	for _, fileInfo := range delta.NewFiles {
		if err := dp.downloadNewFile(fileInfo, targetDir); err != nil {
			return fmt.Errorf("failed to download new file %s: %w", fileInfo.Path, err)
		}
	}

	// 4. 验证结果
	if err := dp.verifyResult(delta, targetDir); err != nil {
		return fmt.Errorf("verification failed: %w", err)
	}

	dp.logger.Info("Delta update applied successfully")
	return nil
}

// copyUnchangedFiles 复制未修改的文件
func (dp *DeltaPatcher) copyUnchangedFiles(delta *DeltaUpdateInfo, sourceDir, targetDir string) error {
	// 获取所有需要处理的文件
	changedFiles := make(map[string]bool)
	for _, patch := range delta.Patches {
		changedFiles[patch.Path] = true
	}
	for _, fileInfo := range delta.NewFiles {
		changedFiles[fileInfo.Path] = true
	}
	for _, path := range delta.DeletedFiles {
		changedFiles[path] = true
	}

	// 复制未修改的文件
	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		// 跳过已修改的文件
		if changedFiles[relPath] {
			return nil
		}

		targetPath := filepath.Join(targetDir, relPath)
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}

		return dp.copyFile(path, targetPath)
	})
}

// applyPatch 应用文件补丁
func (dp *DeltaPatcher) applyPatch(patch FilePatch, sourceDir, targetDir string) error {
	sourcePath := filepath.Join(sourceDir, patch.Path)
	targetPath := filepath.Join(targetDir, patch.Path)

	// 确保目标目录存在
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return err
	}

	// 验证源文件哈希
	actualHash, err := dp.CalculateFileHash(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to hash source file: %w", err)
	}
	if actualHash != patch.OldHash {
		return fmt.Errorf("source file hash mismatch: expected %s, got %s", patch.OldHash, actualHash)
	}

	// 如果有补丁文件，应用补丁
	if patch.PatchURL != "" {
		// 下载补丁
		patchPath := filepath.Join(dp.tempDir, filepath.Base(patch.Path)+".patch")
		if err := dp.downloadPatch(patch.PatchURL, patchPath); err != nil {
			return fmt.Errorf("failed to download patch: %w", err)
		}
		defer os.Remove(patchPath)

		// 应用二进制补丁
		if err := dp.applyBinaryPatch(sourcePath, patchPath, targetPath); err != nil {
			return fmt.Errorf("failed to apply binary patch: %w", err)
		}
	} else {
		// 没有补丁，直接复制（小文件或文本文件）
		if err := dp.copyFile(sourcePath, targetPath); err != nil {
			return err
		}
	}

	// 验证结果哈希
	resultHash, err := dp.CalculateFileHash(targetPath)
	if err != nil {
		return fmt.Errorf("failed to hash result file: %w", err)
	}
	if resultHash != patch.NewHash {
		return fmt.Errorf("result file hash mismatch: expected %s, got %s", patch.NewHash, resultHash)
	}

	return nil
}

// downloadPatch 下载补丁文件
func (dp *DeltaPatcher) downloadPatch(url, destPath string) error {
	// 这里应该实现 HTTP 下载
	// 简化实现，实际应该使用 download 包
	return fmt.Errorf("patch download not implemented")
}

// applyBinaryPatch 应用二进制补丁
func (dp *DeltaPatcher) applyBinaryPatch(sourcePath, patchPath, targetPath string) error {
	// 这里应该实现二进制补丁算法（如 bsdiff）
	// 简化实现：直接复制源文件
	// 实际应该使用 bspatch 或类似算法
	return dp.copyFile(sourcePath, targetPath)
}

// downloadNewFile 下载新文件
func (dp *DeltaPatcher) downloadNewFile(fileInfo FileInfo, targetDir string) error {
	targetPath := filepath.Join(targetDir, fileInfo.Path)

	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return err
	}

	// 这里应该实现 HTTP 下载
	// 简化实现
	return fmt.Errorf("file download not implemented")
}

// verifyResult 验证更新结果
func (dp *DeltaPatcher) verifyResult(delta *DeltaUpdateInfo, targetDir string) error {
	// 验证所有补丁文件
	for _, patch := range delta.Patches {
		targetPath := filepath.Join(targetDir, patch.Path)
		hash, err := dp.CalculateFileHash(targetPath)
		if err != nil {
			return fmt.Errorf("failed to verify file %s: %w", patch.Path, err)
		}
		if hash != patch.NewHash {
			return fmt.Errorf("file %s hash mismatch: expected %s, got %s", patch.Path, patch.NewHash, hash)
		}
	}

	// 验证所有新文件
	for _, fileInfo := range delta.NewFiles {
		targetPath := filepath.Join(targetDir, fileInfo.Path)
		hash, err := dp.CalculateFileHash(targetPath)
		if err != nil {
			return fmt.Errorf("failed to verify new file %s: %w", fileInfo.Path, err)
		}
		if hash != fileInfo.Hash {
			return fmt.Errorf("new file %s hash mismatch: expected %s, got %s", fileInfo.Path, fileInfo.Hash, hash)
		}
	}

	// 验证删除的文件确实不存在
	for _, path := range delta.DeletedFiles {
		targetPath := filepath.Join(targetDir, path)
		if _, err := os.Stat(targetPath); err == nil {
			return fmt.Errorf("deleted file %s still exists", path)
		}
	}

	return nil
}

// copyFile 复制文件
func (dp *DeltaPatcher) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// SaveDeltaManifest 保存差异清单到文件
func (dp *DeltaPatcher) SaveDeltaManifest(delta *DeltaUpdateInfo, path string) error {
	data, err := json.MarshalIndent(delta, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// LoadDeltaManifest 从文件加载差异清单
func (dp *DeltaPatcher) LoadDeltaManifest(path string) (*DeltaUpdateInfo, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var delta DeltaUpdateInfo
	if err := json.Unmarshal(data, &delta); err != nil {
		return nil, err
	}

	return &delta, nil
}

// CalculateSavings 计算差异更新节省的空间
func (dp *DeltaPatcher) CalculateSavings(delta *DeltaUpdateInfo, fullUpdateSize int64) (saved int64, percentage float64) {
	saved = fullUpdateSize - delta.TotalSize
	if fullUpdateSize > 0 {
		percentage = float64(saved) / float64(fullUpdateSize) * 100
	}
	return
}

// BlockHash 块哈希计算（用于大文件差异）
type BlockHash struct {
	Size  int
	Hash  string
	Index int
}

// RollingHash 滚动哈希实现（Rabin-Karp 算法）
type RollingHash struct {
	windowSize int
	hash       uint64
	base       uint64
	mod        uint64
	data       []byte
}

// NewRollingHash 创建新的滚动哈希
func NewRollingHash(windowSize int) *RollingHash {
	return &RollingHash{
		windowSize: windowSize,
		base:       256,
		mod:        1<<61 - 1, // 大素数
		data:       make([]byte, 0, windowSize),
	}
}

// Update 更新哈希
func (rh *RollingHash) Update(b byte) uint64 {
	if len(rh.data) < rh.windowSize {
		rh.data = append(rh.data, b)
		rh.hash = (rh.hash*rh.base + uint64(b)) % rh.mod
	} else {
		old := uint64(rh.data[0])
		rh.data = append(rh.data[1:], b)
		rh.hash = (rh.hash*rh.base + uint64(b) - old*rh.pow(rh.base, rh.windowSize)) % rh.mod
	}
	return rh.hash
}

// pow 计算幂
func (rh *RollingHash) pow(base uint64, exp int) uint64 {
	result := uint64(1)
	for i := 0; i < exp; i++ {
		result = (result * base) % rh.mod
	}
	return result
}

// ChunkedFile 分块文件表示
type ChunkedFile struct {
	Path   string      `json:"path"`
	Size   int64       `json:"size"`
	Chunks []FileChunk `json:"chunks"`
}

// FileChunk 文件块
type FileChunk struct {
	Offset int64  `json:"offset"`
	Size   int64  `json:"size"`
	Hash   string `json:"hash"`
}

// ChunkedDeltaPatcher 基于分块的差异更新
type ChunkedDeltaPatcher struct {
	*DeltaPatcher
	chunkSize int
}

// NewChunkedDeltaPatcher 创建新的分块差异更新器
func NewChunkedDeltaPatcher(tempDir string, chunkSize int, logger Logger) *ChunkedDeltaPatcher {
	if chunkSize <= 0 {
		chunkSize = 64 * 1024 // 默认 64KB 块
	}
	return &ChunkedDeltaPatcher{
		DeltaPatcher: NewDeltaPatcher(tempDir, logger),
		chunkSize:    chunkSize,
	}
}

// CalculateChunkedManifest 计算分块清单
func (cdp *ChunkedDeltaPatcher) CalculateChunkedManifest(filePath string) (*ChunkedFile, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	chunkedFile := &ChunkedFile{
		Path:   filePath,
		Size:   stat.Size(),
		Chunks: []FileChunk{},
	}

	buffer := make([]byte, cdp.chunkSize)
	offset := int64(0)

	for {
		n, err := file.Read(buffer)
		if n > 0 {
			hash := sha256.Sum256(buffer[:n])
			chunkedFile.Chunks = append(chunkedFile.Chunks, FileChunk{
				Offset: offset,
				Size:   int64(n),
				Hash:   hex.EncodeToString(hash[:]),
			})
			offset += int64(n)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
	}

	return chunkedFile, nil
}

// FindMatchingChunks 查找匹配的块
func (cdp *ChunkedDeltaPatcher) FindMatchingChunks(oldChunks, newChunks []FileChunk) ([]int, []int) {
	// 创建旧块哈希映射
	oldHashMap := make(map[string][]int)
	for i, chunk := range oldChunks {
		oldHashMap[chunk.Hash] = append(oldHashMap[chunk.Hash], i)
	}

	// 查找匹配
	matchedOld := make([]int, 0)
	matchedNew := make([]int, 0)

	for newIdx, newChunk := range newChunks {
		if oldIndices, exists := oldHashMap[newChunk.Hash]; exists {
			// 找到匹配，取第一个未匹配的
			for _, oldIdx := range oldIndices {
				if !contains(matchedOld, oldIdx) {
					matchedOld = append(matchedOld, oldIdx)
					matchedNew = append(matchedNew, newIdx)
					break
				}
			}
		}
	}

	return matchedOld, matchedNew
}

// contains 检查切片是否包含元素
func contains(slice []int, item int) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// SortDeltaPatches 按路径排序补丁
func (dp *DeltaPatcher) SortDeltaPatches(delta *DeltaUpdateInfo) {
	sort.Slice(delta.Patches, func(i, j int) bool {
		return delta.Patches[i].Path < delta.Patches[j].Path
	})
	sort.Slice(delta.NewFiles, func(i, j int) bool {
		return delta.NewFiles[i].Path < delta.NewFiles[j].Path
	})
	sort.Strings(delta.DeletedFiles)
}
