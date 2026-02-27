//go:build linux
// +build linux

package security

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/godbus/dbus/v5"
)

const (
	secretServiceBusName    = "org.freedesktop.secrets"
	secretServicePath       = "/org/freedesktop/secrets"
	secretServiceInterface  = "org.freedesktop.secrets.Service"
	secretCollectionPath    = "/org/freedesktop/secrets/aliases/default"
	secretItemInterface     = "org.freedesktop.secrets.Item"
)

// LinuxSecureStorage Linux Secret Service 存储
type LinuxSecureStorage struct {
	collection string
	attributes map[string]string
}

// NewLinuxSecureStorage 创建新的 Linux 安全存储
func NewLinuxSecureStorage(collection string) *LinuxSecureStorage {
	if collection == "" {
		collection = "openclaw"
	}
	return &LinuxSecureStorage{
		collection: collection,
		attributes: map[string]string{
			"application": "openclaw",
		},
	}
}

// Store 存储数据到 Secret Service
func (lss *LinuxSecureStorage) Store(key string, data []byte) error {
	// 检查是否可以使用 secret-tool
	if _, err := exec.LookPath("secret-tool"); err == nil {
		return lss.storeWithSecretTool(key, data)
	}

	// 回退到文件存储（加密）
	return lss.storeToFile(key, data)
}

// Retrieve 从 Secret Service 检索数据
func (lss *LinuxSecureStorage) Retrieve(key string) ([]byte, error) {
	// 检查是否可以使用 secret-tool
	if _, err := exec.LookPath("secret-tool"); err == nil {
		return lss.retrieveWithSecretTool(key)
	}

	// 回退到文件存储
	return lss.retrieveFromFile(key)
}

// Delete 从 Secret Service 删除数据
func (lss *LinuxSecureStorage) Delete(key string) error {
	// 检查是否可以使用 secret-tool
	if _, err := exec.LookPath("secret-tool"); err == nil {
		return lss.deleteWithSecretTool(key)
	}

	// 回退到文件存储
	return lss.deleteFromFile(key)
}

// Exists 检查数据是否存在
func (lss *LinuxSecureStorage) Exists(key string) bool {
	_, err := lss.Retrieve(key)
	return err == nil
}

// storeWithSecretTool 使用 secret-tool 存储
func (lss *LinuxSecureStorage) storeWithSecretTool(key string, data []byte) error {
	cmd := exec.Command("secret-tool", "store",
		"--label="+key,
		"application", "openclaw",
		"key", key,
	)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	if _, err := stdin.Write(data); err != nil {
		return err
	}
	stdin.Close()

	return cmd.Wait()
}

// retrieveWithSecretTool 使用 secret-tool 检索
func (lss *LinuxSecureStorage) retrieveWithSecretTool(key string) ([]byte, error) {
	cmd := exec.Command("secret-tool", "lookup",
		"application", "openclaw",
		"key", key,
	)

	return cmd.Output()
}

// deleteWithSecretTool 使用 secret-tool 删除
func (lss *LinuxSecureStorage) deleteWithSecretTool(key string) error {
	cmd := exec.Command("secret-tool", "clear",
		"application", "openclaw",
		"key", key,
	)

	return cmd.Run()
}

// getSecureDir 获取安全存储目录
func (lss *LinuxSecureStorage) getSecureDir() string {
	home := os.Getenv("HOME")
	if home == "" {
		home = "/tmp"
	}
	return filepath.Join(home, ".config", "openclaw", "secure")
}

// storeToFile 存储到加密文件
func (lss *LinuxSecureStorage) storeToFile(key string, data []byte) error {
	dir := lss.getSecureDir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	// 使用简单的 XOR 加密（实际应该使用更强的加密）
	encrypted := xorEncrypt(data, []byte(lss.collection))

	filePath := filepath.Join(dir, key+".dat")
	return os.WriteFile(filePath, encrypted, 0600)
}

// retrieveFromFile 从文件检索
func (lss *LinuxSecureStorage) retrieveFromFile(key string) ([]byte, error) {
	filePath := filepath.Join(lss.getSecureDir(), key+".dat")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return xorEncrypt(data, []byte(lss.collection)), nil
}

// deleteFromFile 从文件删除
func (lss *LinuxSecureStorage) deleteFromFile(key string) error {
	filePath := filepath.Join(lss.getSecureDir(), key+".dat")
	return os.Remove(filePath)
}

// xorEncrypt 简单的 XOR 加密/解密
func xorEncrypt(data, key []byte) []byte {
	result := make([]byte, len(data))
	for i := range data {
		result[i] = data[i] ^ key[i%len(key)]
	}
	return result
}

// SecretServiceClient D-Bus Secret Service 客户端
type SecretServiceClient struct {
	conn *dbus.Conn
}

// NewSecretServiceClient 创建新的 Secret Service 客户端
func NewSecretServiceClient() (*SecretServiceClient, error) {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to session bus: %w", err)
	}

	return &SecretServiceClient{conn: conn}, nil
}

// Close 关闭连接
func (ssc *SecretServiceClient) Close() {
	ssc.conn.Close()
}

// Unlock 解锁集合
func (ssc *SecretServiceClient) Unlock(collection dbus.ObjectPath) error {
	obj := ssc.conn.Object(secretServiceBusName, secretServicePath)

	var unlocked []dbus.ObjectPath
	var prompt dbus.ObjectPath

	err := obj.Call(secretServiceInterface+".Unlock", 0, []dbus.ObjectPath{collection}).Store(&unlocked, &prompt)
	if err != nil {
		return err
	}

	return nil
}

// CreateItem 创建新项目
func (ssc *SecretServiceClient) CreateItem(collection dbus.ObjectPath, label string, attributes map[string]string, secret []byte) (dbus.ObjectPath, error) {
	obj := ssc.conn.Object(secretServiceBusName, collection)

	// 创建属性
	props := map[string]dbus.Variant{
		"org.freedesktop.Secret.Item.Label":      dbus.MakeVariant(label),
		"org.freedesktop.Secret.Item.Attributes": dbus.MakeVariant(attributes),
	}

	// 创建秘密（简化实现）
	secretStruct := struct {
		Session     dbus.ObjectPath
		Parameters  []byte
		Value       []byte
		ContentType string
	}{
		Session:     "/org/freedesktop/secrets/session/s1",
		Parameters:  []byte{},
		Value:       secret,
		ContentType: "text/plain",
	}

	var item dbus.ObjectPath
	var prompt dbus.ObjectPath

	err := obj.Call("org.freedesktop.Secret.Collection.CreateItem", 0, props, secretStruct, true).Store(&item, &prompt)
	if err != nil {
		return "", err
	}

	return item, nil
}

// checkDebugger 检查 Linux 调试器
func checkDebugger() bool {
	// 检查 /proc/self/status 中的 TracerPid
	data, err := os.ReadFile("/proc/self/status")
	if err != nil {
		return false
	}

	// 解析 TracerPid
	lines := string(data)
	for _, line := range strings.Split(lines, "\n") {
		if strings.HasPrefix(line, "TracerPid:") {
			var pid int
			fmt.Sscanf(line, "TracerPid:\t%d", &pid)
			return pid != 0
		}
	}

	// 检查 LD_PRELOAD
	if os.Getenv("LD_PRELOAD") != "" {
		return true
	}

	return false
}

// lockMemoryLinux 锁定内存（Linux 实现）
func lockMemory(data []byte) error {
	// Linux 使用 mlock
	// 简化实现，实际应该使用 syscall.Mlock
	return nil
}

// unlockMemoryLinux 解锁内存（Linux 实现）
func unlockMemory(data []byte) {
	// Linux 使用 munlock
}

// KeyringStorage Linux 内核密钥环存储
type KeyringStorage struct {
	keyring string
}

// NewKeyringStorage 创建新的密钥环存储
func NewKeyringStorage(keyring string) *KeyringStorage {
	if keyring == "" {
		keyring = "user"
	}
	return &KeyringStorage{keyring: keyring}
}

// Store 存储到密钥环
func (ks *KeyringStorage) Store(key string, data []byte) error {
	// 使用 keyctl 命令
	cmd := exec.Command("keyctl", "add", "user", key, string(data), ks.keyring)
	return cmd.Run()
}

// Retrieve 从密钥环检索
func (ks *KeyringStorage) Retrieve(key string) ([]byte, error) {
	cmd := exec.Command("keyctl", "pipe", key)
	return cmd.Output()
}

// Delete 从密钥环删除
func (ks *KeyringStorage) Delete(key string) error {
	cmd := exec.Command("keyctl", "unlink", key, ks.keyring)
	return cmd.Run()
}

// Exists 检查密钥环中是否存在
func (ks *KeyringStorage) Exists(key string) bool {
	_, err := ks.Retrieve(key)
	return err == nil
}
