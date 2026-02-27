// Package security 提供加密和安全功能
package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"bufio"
	"strings"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/pbkdf2"
)

const (
	// AES 密钥长度
	aesKeySize = 32 // 256-bit

	// Argon2 参数
	argon2Time    = 3
	argon2Memory  = 64 * 1024 // 64 MB
	argon2Threads = 4
	argon2KeyLen  = 32

	// PBKDF2 参数（备选）
	pbkdf2Iterations = 100000
)

// SecureStorage 提供安全的存储接口
type SecureStorage interface {
	// Store 安全存储数据
	Store(key string, data []byte) error
	// Retrieve 检索存储的数据
	Retrieve(key string) ([]byte, error)
	// Delete 删除存储的数据
	Delete(key string) error
	// Exists 检查键是否存在
	Exists(key string) bool
}

// CryptoProvider 提供加密功能
type CryptoProvider struct {
	masterKey []byte
	salt      []byte
}

// NewCryptoProvider 创建新的加密提供者
func NewCryptoProvider(password string, salt []byte) (*CryptoProvider, error) {
	if len(salt) == 0 {
		// 生成随机 salt
		salt = make([]byte, 16)
		if _, err := rand.Read(salt); err != nil {
			return nil, fmt.Errorf("failed to generate salt: %w", err)
		}
	}

	// 使用 Argon2 派生密钥
	key := deriveKey(password, salt)

	return &CryptoProvider{
		masterKey: key,
		salt:      salt,
	}, nil
}

// deriveKey 使用 Argon2 派生密钥
func deriveKey(password string, salt []byte) []byte {
	return argon2.IDKey(
		[]byte(password),
		salt,
		argon2Time,
		argon2Memory,
		argon2Threads,
		argon2KeyLen,
	)
}

// deriveKeyPBKDF2 使用 PBKDF2 派生密钥（备选）
func deriveKeyPBKDF2(password string, salt []byte) []byte {
	return pbkdf2.Key([]byte(password), salt, pbkdf2Iterations, aesKeySize, sha256.New)
}

// Encrypt 加密数据
func (cp *CryptoProvider) Encrypt(plaintext []byte) ([]byte, error) {
	// 生成随机 IV
	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		return nil, fmt.Errorf("failed to generate IV: %w", err)
	}

	// 创建 AES 密码块
	block, err := aes.NewCipher(cp.masterKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// 填充数据
	paddedData := pkcs7Pad(plaintext, aes.BlockSize)

	// 加密
	ciphertext := make([]byte, len(paddedData))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, paddedData)

	// 组合 salt + iv + ciphertext
	result := make([]byte, 0, len(cp.salt)+len(iv)+len(ciphertext))
	result = append(result, cp.salt...)
	result = append(result, iv...)
	result = append(result, ciphertext...)

	return result, nil
}

// Decrypt 解密数据
func (cp *CryptoProvider) Decrypt(ciphertext []byte) ([]byte, error) {
	if len(ciphertext) < 16+aes.BlockSize {
		return nil, errors.New("ciphertext too short")
	}

	// 分离 salt, iv 和密文
	iv := ciphertext[16 : 16+aes.BlockSize]
	encrypted := ciphertext[16+aes.BlockSize:]

	// 创建 AES 密码块（使用 masterKey）
	block, err := aes.NewCipher(cp.masterKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// 解密
	mode := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(encrypted))
	mode.CryptBlocks(plaintext, encrypted)

	// 去除填充
	return pkcs7Unpad(plaintext)
}

// EncryptString 加密字符串
func (cp *CryptoProvider) EncryptString(plaintext string) (string, error) {
	encrypted, err := cp.Encrypt([]byte(plaintext))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// DecryptString 解密字符串
func (cp *CryptoProvider) DecryptString(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode: %w", err)
	}

	decrypted, err := cp.Decrypt(data)
	if err != nil {
		return "", err
	}

	return string(decrypted), nil
}

// pkcs7Pad 添加 PKCS7 填充
func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := make([]byte, padding)
	for i := range padtext {
		padtext[i] = byte(padding)
	}
	return append(data, padtext...)
}

// pkcs7Unpad 去除 PKCS7 填充
func pkcs7Unpad(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("empty data")
	}

	padding := int(data[len(data)-1])
	if padding > len(data) || padding == 0 {
		return nil, errors.New("invalid padding")
	}

	// 验证填充
	for i := 0; i < padding; i++ {
		if data[len(data)-1-i] != byte(padding) {
			return nil, errors.New("invalid padding")
		}
	}

	return data[:len(data)-padding], nil
}

// GenerateRandomKey 生成随机密钥
func GenerateRandomKey(size int) ([]byte, error) {
	key := make([]byte, size)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	return key, nil
}

// GenerateSecureToken 生成安全令牌
func GenerateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// HashData 计算 SHA256 哈希
func HashData(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// HashString 计算字符串的 SHA256 哈希
func HashString(s string) string {
	return HashData([]byte(s))
}

// SecureCompare 安全比较（防止时序攻击）
func SecureCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// SecureCompareBytes 安全比较字节（防止时序攻击）
func SecureCompareBytes(a, b []byte) bool {
	return subtle.ConstantTimeCompare(a, b) == 1
}

// EncryptedConfig 加密配置存储
type EncryptedConfig struct {
	provider *CryptoProvider
	dataDir  string
}

// NewEncryptedConfig 创建新的加密配置存储
func NewEncryptedConfig(password string, dataDir string) (*EncryptedConfig, error) {
	if dataDir == "" {
		dataDir = getDefaultDataDir()
	}

	// 确保目录存在
	if err := os.MkdirAll(dataDir, 0750); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	// 加载或生成 salt
	saltPath := filepath.Join(dataDir, ".salt")
	var salt []byte

	if _, err := os.Stat(saltPath); err == nil {
		// 加载现有 salt
		salt, err = os.ReadFile(saltPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read salt: %w", err)
		}
	} else {
		// 生成新 salt
		salt = make([]byte, 16)
		if _, err := rand.Read(salt); err != nil {
			return nil, fmt.Errorf("failed to generate salt: %w", err)
		}
		// 保存 salt
		if err := os.WriteFile(saltPath, salt, 0600); err != nil {
			return nil, fmt.Errorf("failed to write salt: %w", err)
		}
	}

	provider, err := NewCryptoProvider(password, salt)
	if err != nil {
		return nil, err
	}

	return &EncryptedConfig{
		provider: provider,
		dataDir:  dataDir,
	}, nil
}

// Save 保存加密配置
func (ec *EncryptedConfig) Save(key string, value []byte) error {
	encrypted, err := ec.provider.Encrypt(value)
	if err != nil {
		return err
	}

	filePath := filepath.Join(ec.dataDir, key+".enc")
	return os.WriteFile(filePath, encrypted, 0600)
}

// Load 加载加密配置
func (ec *EncryptedConfig) Load(key string) ([]byte, error) {
	filePath := filepath.Join(ec.dataDir, key+".enc")
	encrypted, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return ec.provider.Decrypt(encrypted)
}

// Delete 删除配置
func (ec *EncryptedConfig) Delete(key string) error {
	filePath := filepath.Join(ec.dataDir, key+".enc")
	return os.Remove(filePath)
}

// Exists 检查配置是否存在
func (ec *EncryptedConfig) Exists(key string) bool {
	filePath := filepath.Join(ec.dataDir, key+".enc")
	_, err := os.Stat(filePath)
	return err == nil
}

// getDefaultDataDir 获取默认数据目录
func getDefaultDataDir() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("ProgramData"), "OpenClaw", "secure")
	case "darwin":
		return filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "OpenClaw", "secure")
	default: // linux
		return filepath.Join(os.Getenv("HOME"), ".config", "openclaw", "secure")
	}
}

// IntegrityChecker 完整性检查器
type IntegrityChecker struct {
	baseline map[string]string // 文件路径 -> 哈希
}

// NewIntegrityChecker 创建新的完整性检查器
func NewIntegrityChecker() *IntegrityChecker {
	return &IntegrityChecker{
		baseline: make(map[string]string),
	}
}

// AddFile 添加文件到基线
func (ic *IntegrityChecker) AddFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	ic.baseline[path] = HashData(data)
	return nil
}

// Verify 验证文件完整性
func (ic *IntegrityChecker) Verify(path string) (bool, error) {
	expectedHash, exists := ic.baseline[path]
	if !exists {
		return false, fmt.Errorf("file not in baseline: %s", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}

	actualHash := HashData(data)
	return SecureCompare(expectedHash, actualHash), nil
}

// VerifyAll 验证所有文件
func (ic *IntegrityChecker) VerifyAll() (map[string]bool, error) {
	results := make(map[string]bool)

	for path := range ic.baseline {
		valid, err := ic.Verify(path)
		if err != nil {
			results[path] = false
		} else {
			results[path] = valid
		}
	}

	return results, nil
}

// SaveBaseline 保存基线到文件
func (ic *IntegrityChecker) SaveBaseline(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	for filePath, hash := range ic.baseline {
		fmt.Fprintf(file, "%s %s\n", hash, filePath)
	}

	return nil
}

// LoadBaseline 从文件加载基线
func (ic *IntegrityChecker) LoadBaseline(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, " ", 2)
		if len(parts) == 2 {
			ic.baseline[parts[1]] = parts[0]
		}
	}

	return scanner.Err()
}

// AntiDebug 反调试检测
type AntiDebug struct {
	detected chan bool
	stop     chan bool
}

// NewAntiDebug 创建新的反调试检测器
func NewAntiDebug() *AntiDebug {
	return &AntiDebug{
		detected: make(chan bool, 1),
		stop:     make(chan bool),
	}
}

// Start 启动反调试检测
func (ad *AntiDebug) Start() {
	go ad.detect()
}

// Stop 停止反调试检测
func (ad *AntiDebug) Stop() {
	close(ad.stop)
}

// Detected 返回调试检测通道
func (ad *AntiDebug) Detected() <-chan bool {
	return ad.detected
}

// detect 执行检测（平台特定实现）
func (ad *AntiDebug) detect() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ad.stop:
			return
		case <-ticker.C:
			if ad.checkDebugger() {
				select {
				case ad.detected <- true:
				default:
				}
			}
		}
	}
}

// checkDebugger 检查是否连接调试器（平台特定）
func (ad *AntiDebug) checkDebugger() bool {
	return checkDebugger()
}

// SecureMemory 安全内存（防止交换到磁盘）
type SecureMemory struct {
	data []byte
}

// NewSecureMemory 创建新的安全内存区域
func NewSecureMemory(size int) (*SecureMemory, error) {
	data := make([]byte, size)

	// 锁定内存（平台特定）
	if err := lockMemory(data); err != nil {
		return nil, err
	}

	return &SecureMemory{data: data}, nil
}

// Write 写入数据到安全内存
func (sm *SecureMemory) Write(data []byte) {
	copy(sm.data, data)
}

// Read 从安全内存读取数据
func (sm *SecureMemory) Read() []byte {
	result := make([]byte, len(sm.data))
	copy(result, sm.data)
	return result
}

// Clear 清除安全内存
func (sm *SecureMemory) Clear() {
	// 安全擦除
	for i := range sm.data {
		sm.data[i] = 0
	}
}

// Destroy 销毁安全内存
func (sm *SecureMemory) Destroy() {
	sm.Clear()
	// 解锁内存（平台特定）
	unlockMemory(sm.data)
	sm.data = nil
}
