package security

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestNewCryptoProvider(t *testing.T) {
	password := "test-password"
	salt := []byte("test-salt-123456")

	cp, err := NewCryptoProvider(password, salt)
	if err != nil {
		t.Fatalf("NewCryptoProvider failed: %v", err)
	}

	if cp == nil {
		t.Fatal("NewCryptoProvider returned nil")
	}

	if len(cp.masterKey) != 32 {
		t.Errorf("Expected key length 32, got %d", len(cp.masterKey))
	}
}

func TestNewCryptoProviderGenerateSalt(t *testing.T) {
	password := "test-password"

	cp, err := NewCryptoProvider(password, nil)
	if err != nil {
		t.Fatalf("NewCryptoProvider failed: %v", err)
	}

	if len(cp.salt) != 16 {
		t.Errorf("Expected salt length 16, got %d", len(cp.salt))
	}
}

func TestEncryptDecrypt(t *testing.T) {
	password := "test-password"
	salt := []byte("test-salt-123456")

	cp, err := NewCryptoProvider(password, salt)
	if err != nil {
		t.Fatalf("NewCryptoProvider failed: %v", err)
	}

	plaintext := []byte("Hello, World! This is a test message.")

	// 加密
	ciphertext, err := cp.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	// 密文应该与明文不同
	if bytes.Equal(ciphertext, plaintext) {
		t.Error("Ciphertext should be different from plaintext")
	}

	// 解密
	decrypted, err := cp.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	// 解密后的数据应该与原始数据相同
	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("Decrypted data doesn't match original: got %s, want %s", decrypted, plaintext)
	}
}

func TestEncryptDecryptString(t *testing.T) {
	password := "test-password"
	salt := []byte("test-salt-123456")

	cp, err := NewCryptoProvider(password, salt)
	if err != nil {
		t.Fatalf("NewCryptoProvider failed: %v", err)
	}

	plaintext := "Hello, World!"

	// 加密
	ciphertext, err := cp.EncryptString(plaintext)
	if err != nil {
		t.Fatalf("EncryptString failed: %v", err)
	}

	// 解密
	decrypted, err := cp.DecryptString(ciphertext)
	if err != nil {
		t.Fatalf("DecryptString failed: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("Decrypted string doesn't match: got %s, want %s", decrypted, plaintext)
	}
}

func TestPkcs7Pad(t *testing.T) {
	tests := []struct {
		data      []byte
		blockSize int
		expected  int
	}{
		{[]byte("12345678"), 16, 16},
		{[]byte("1234567890123456"), 16, 32},
		{[]byte("1"), 8, 8},
		{[]byte(""), 16, 16},
	}

	for _, tt := range tests {
		padded := pkcs7Pad(tt.data, tt.blockSize)
		if len(padded) != tt.expected {
			t.Errorf("pkcs7Pad(%s, %d) length = %d, want %d", tt.data, tt.blockSize, len(padded), tt.expected)
		}

		// 验证填充可以正确去除
		unpadded, err := pkcs7Unpad(padded)
		if err != nil {
			t.Errorf("pkcs7Unpad failed: %v", err)
		}

		if !bytes.Equal(unpadded, tt.data) {
			t.Errorf("pkcs7Unpad result = %s, want %s", unpadded, tt.data)
		}
	}
}

func TestPkcs7UnpadInvalid(t *testing.T) {
	// 测试空数据
	_, err := pkcs7Unpad([]byte{})
	if err == nil {
		t.Error("pkcs7Unpad should fail on empty data")
	}

	// 测试无效填充 - 填充值大于数据长度
	invalid := []byte{1, 2, 3, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20}
	_, err = pkcs7Unpad(invalid)
	if err == nil {
		t.Error("pkcs7Unpad should fail on invalid padding")
	}
}

func TestGenerateRandomKey(t *testing.T) {
	key, err := GenerateRandomKey(32)
	if err != nil {
		t.Fatalf("GenerateRandomKey failed: %v", err)
	}

	if len(key) != 32 {
		t.Errorf("Expected key length 32, got %d", len(key))
	}

	// 再次生成，应该不同
	key2, err := GenerateRandomKey(32)
	if err != nil {
		t.Fatalf("GenerateRandomKey failed: %v", err)
	}

	if bytes.Equal(key, key2) {
		t.Error("Random keys should be different")
	}
}

func TestGenerateSecureToken(t *testing.T) {
	token, err := GenerateSecureToken(32)
	if err != nil {
		t.Fatalf("GenerateSecureToken failed: %v", err)
	}

	if len(token) != 64 { // hex encoding doubles the length
		t.Errorf("Expected token length 64, got %d", len(token))
	}

	// 再次生成，应该不同
	token2, err := GenerateSecureToken(32)
	if err != nil {
		t.Fatalf("GenerateSecureToken failed: %v", err)
	}

	if token == token2 {
		t.Error("Random tokens should be different")
	}
}

func TestHashData(t *testing.T) {
	data := []byte("test data")
	hash := HashData(data)

	if len(hash) != 64 { // SHA256 hex string length
		t.Errorf("Expected hash length 64, got %d", len(hash))
	}

	// 相同数据应该产生相同哈希
	hash2 := HashData(data)
	if hash != hash2 {
		t.Error("Same data should produce same hash")
	}

	// 不同数据应该产生不同哈希
	hash3 := HashData([]byte("different data"))
	if hash == hash3 {
		t.Error("Different data should produce different hash")
	}
}

func TestHashString(t *testing.T) {
	s := "test string"
	hash := HashString(s)

	if len(hash) != 64 {
		t.Errorf("Expected hash length 64, got %d", len(hash))
	}
}

func TestSecureCompare(t *testing.T) {
	// 相同字符串
	if !SecureCompare("test", "test") {
		t.Error("SecureCompare should return true for equal strings")
	}

	// 不同字符串
	if SecureCompare("test", "different") {
		t.Error("SecureCompare should return false for different strings")
	}

	// 空字符串
	if !SecureCompare("", "") {
		t.Error("SecureCompare should return true for empty strings")
	}
}

func TestSecureCompareBytes(t *testing.T) {
	// 相同字节
	if !SecureCompareBytes([]byte("test"), []byte("test")) {
		t.Error("SecureCompareBytes should return true for equal bytes")
	}

	// 不同字节
	if SecureCompareBytes([]byte("test"), []byte("different")) {
		t.Error("SecureCompareBytes should return false for different bytes")
	}
}

func TestEncryptedConfig(t *testing.T) {
	tempDir := t.TempDir()
	password := "test-password"

	ec, err := NewEncryptedConfig(password, tempDir)
	if err != nil {
		t.Fatalf("NewEncryptedConfig failed: %v", err)
	}

	// 测试保存
	key := "test-key"
	value := []byte("test-value")
	if err := ec.Save(key, value); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// 测试存在
	if !ec.Exists(key) {
		t.Error("Exists should return true for existing key")
	}

	// 测试加载
	loaded, err := ec.Load(key)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if !bytes.Equal(loaded, value) {
		t.Errorf("Loaded value doesn't match: got %s, want %s", loaded, value)
	}

	// 测试删除
	if err := ec.Delete(key); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	if ec.Exists(key) {
		t.Error("Exists should return false after delete")
	}
}

func TestEncryptedConfigLoadNonExistent(t *testing.T) {
	tempDir := t.TempDir()
	password := "test-password"

	ec, err := NewEncryptedConfig(password, tempDir)
	if err != nil {
		t.Fatalf("NewEncryptedConfig failed: %v", err)
	}

	_, err = ec.Load("non-existent-key")
	if err == nil {
		t.Error("Load should fail for non-existent key")
	}
}

func TestIntegrityChecker(t *testing.T) {
	// 创建临时文件
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	testData := []byte("test data for integrity check")
	if err := os.WriteFile(testFile, testData, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	ic := NewIntegrityChecker()

	// 添加文件到基线
	if err := ic.AddFile(testFile); err != nil {
		t.Fatalf("AddFile failed: %v", err)
	}

	// 验证文件
	valid, err := ic.Verify(testFile)
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}

	if !valid {
		t.Error("File should be valid")
	}

	// 修改文件
	if err := os.WriteFile(testFile, []byte("modified data"), 0644); err != nil {
		t.Fatalf("Failed to modify test file: %v", err)
	}

	// 再次验证
	valid, err = ic.Verify(testFile)
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}

	if valid {
		t.Error("Modified file should be invalid")
	}
}

func TestIntegrityCheckerVerifyNonExistent(t *testing.T) {
	ic := NewIntegrityChecker()

	_, err := ic.Verify("/non/existent/file")
	if err == nil {
		t.Error("Verify should fail for non-existent file")
	}
}

func TestIntegrityCheckerSaveLoadBaseline(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	baselineFile := filepath.Join(tempDir, "baseline.txt")

	if err := os.WriteFile(testFile, []byte("test data"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	ic := NewIntegrityChecker()
	if err := ic.AddFile(testFile); err != nil {
		t.Fatalf("AddFile failed: %v", err)
	}

	// 保存基线
	if err := ic.SaveBaseline(baselineFile); err != nil {
		t.Fatalf("SaveBaseline failed: %v", err)
	}

	// 加载基线到新检查器
	ic2 := NewIntegrityChecker()
	if err := ic2.LoadBaseline(baselineFile); err != nil {
		t.Fatalf("LoadBaseline failed: %v", err)
	}

	// 验证
	valid, err := ic2.Verify(testFile)
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}

	if !valid {
		t.Error("File should be valid after loading baseline")
	}
}

func TestAntiDebug(t *testing.T) {
	ad := NewAntiDebug()

	// 启动检测
	ad.Start()

	// 停止检测
	ad.Stop()
}

func TestSecureMemory(t *testing.T) {
	sm, err := NewSecureMemory(64)
	if err != nil {
		t.Fatalf("NewSecureMemory failed: %v", err)
	}

	// 写入数据
	data := []byte("secret data that should be protected")
	sm.Write(data)

	// 读取数据
	read := sm.Read()
	if !bytes.Equal(read[:len(data)], data) {
		t.Error("Read data doesn't match written data")
	}

	// 清除
	sm.Clear()

	// 销毁
	sm.Destroy()

	if sm.data != nil {
		t.Error("data should be nil after Destroy")
	}
}

func BenchmarkEncrypt(b *testing.B) {
	cp, _ := NewCryptoProvider("password", nil)
	data := []byte("benchmark data for encryption")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cp.Encrypt(data)
	}
}

func BenchmarkDecrypt(b *testing.B) {
	cp, _ := NewCryptoProvider("password", nil)
	data := []byte("benchmark data for encryption")
	ciphertext, _ := cp.Encrypt(data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cp.Decrypt(ciphertext)
	}
}

func BenchmarkHashData(b *testing.B) {
	data := []byte("benchmark data for hashing")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		HashData(data)
	}
}
