//go:build windows
// +build windows

package security

import (
	"fmt"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Windows 凭据管理器常量
const (
	CRED_TYPE_GENERIC               = 1
	CRED_PERSIST_LOCAL_MACHINE      = 2
	CRED_PERSIST_ENTERPRISE         = 3
	CRED_FLAGS_PROMPT_NOW           = 0x2
	CRED_FLAGS_USERNAME_TARGET      = 0x4
)

// CREDENTIAL 结构体
type CREDENTIAL struct {
	Flags              uint32
	Type               uint32
	TargetName         *uint16
	Comment            *uint16
	LastWritten        windows.Filetime
	CredentialBlobSize uint32
	CredentialBlob     *byte
	Persist            uint32
	AttributeCount     uint32
	Attributes         uintptr
	TargetAlias        *uint16
	UserName           *uint16
}

var (
	advapi32                       = windows.NewLazySystemDLL("advapi32.dll")
	procCredWriteW                 = advapi32.NewProc("CredWriteW")
	procCredReadW                  = advapi32.NewProc("CredReadW")
	procCredDeleteW                = advapi32.NewProc("CredDeleteW")
	procCredFree                   = advapi32.NewProc("CredFree")
)

// WindowsSecureStorage Windows 凭据管理器存储
type WindowsSecureStorage struct {
	targetPrefix string
}

// NewWindowsSecureStorage 创建新的 Windows 安全存储
func NewWindowsSecureStorage(targetPrefix string) *WindowsSecureStorage {
	if targetPrefix == "" {
		targetPrefix = "OpenClaw"
	}
	return &WindowsSecureStorage{targetPrefix: targetPrefix}
}

// Store 存储凭据到 Windows 凭据管理器
func (wss *WindowsSecureStorage) Store(key string, data []byte) error {
	targetName := wss.targetPrefix + "/" + key

	targetNamePtr, err := syscall.UTF16PtrFromString(targetName)
	if err != nil {
		return err
	}

	userNamePtr, err := syscall.UTF16PtrFromString("OpenClawUser")
	if err != nil {
		return err
	}

	// 创建凭据结构
	cred := CREDENTIAL{
		Type:               CRED_TYPE_GENERIC,
		TargetName:         targetNamePtr,
		UserName:           userNamePtr,
		CredentialBlobSize: uint32(len(data)),
		CredentialBlob:     &data[0],
		Persist:            CRED_PERSIST_LOCAL_MACHINE,
	}

	ret, _, err := procCredWriteW.Call(
		uintptr(unsafe.Pointer(&cred)),
		0,
	)

	if ret == 0 {
		return fmt.Errorf("CredWrite failed: %v", err)
	}

	return nil
}

// Retrieve 从 Windows 凭据管理器检索凭据
func (wss *WindowsSecureStorage) Retrieve(key string) ([]byte, error) {
	targetName := wss.targetPrefix + "/" + key

	targetNamePtr, err := syscall.UTF16PtrFromString(targetName)
	if err != nil {
		return nil, err
	}

	var credPtr *CREDENTIAL

	ret, _, err := procCredReadW.Call(
		uintptr(unsafe.Pointer(targetNamePtr)),
		CRED_TYPE_GENERIC,
		0,
		uintptr(unsafe.Pointer(&credPtr)),
	)

	if ret == 0 {
		return nil, fmt.Errorf("CredRead failed: %v", err)
	}

	defer procCredFree.Call(uintptr(unsafe.Pointer(credPtr)))

	// 复制凭据数据
	data := make([]byte, credPtr.CredentialBlobSize)
	copy(data, unsafe.Slice(credPtr.CredentialBlob, credPtr.CredentialBlobSize))

	return data, nil
}

// Delete 从 Windows 凭据管理器删除凭据
func (wss *WindowsSecureStorage) Delete(key string) error {
	targetName := wss.targetPrefix + "/" + key

	targetNamePtr, err := syscall.UTF16PtrFromString(targetName)
	if err != nil {
		return err
	}

	ret, _, err := procCredDeleteW.Call(
		uintptr(unsafe.Pointer(targetNamePtr)),
		CRED_TYPE_GENERIC,
		0,
	)

	if ret == 0 {
		return fmt.Errorf("CredDelete failed: %v", err)
	}

	return nil
}

// Exists 检查凭据是否存在
func (wss *WindowsSecureStorage) Exists(key string) bool {
	_, err := wss.Retrieve(key)
	return err == nil
}

// checkDebugger 检查 Windows 调试器
func checkDebugger() bool {
	// 使用 IsDebuggerPresent API
	kernel32 := windows.NewLazySystemDLL("kernel32.dll")
	isDebuggerPresent := kernel32.NewProc("IsDebuggerPresent")

	ret, _, _ := isDebuggerPresent.Call()
	return ret != 0
}

// lockMemoryWindows 锁定内存（Windows 实现）
func lockMemory(data []byte) error {
	kernel32 := windows.NewLazySystemDLL("kernel32.dll")
	virtualLock := kernel32.NewProc("VirtualLock")

	ret, _, err := virtualLock.Call(
		uintptr(unsafe.Pointer(&data[0])),
		uintptr(len(data)),
	)

	if ret == 0 {
		return fmt.Errorf("VirtualLock failed: %v", err)
	}

	return nil
}

// unlockMemoryWindows 解锁内存（Windows 实现）
func unlockMemory(data []byte) {
	kernel32 := windows.NewLazySystemDLL("kernel32.dll")
	virtualUnlock := kernel32.NewProc("VirtualUnlock")

	virtualUnlock.Call(
		uintptr(unsafe.Pointer(&data[0])),
		uintptr(len(data)),
	)
}

// DataProtection 使用 Windows DPAPI 加密数据
func DataProtection(data []byte) ([]byte, error) {
	// 使用 CryptProtectData API
	// 简化实现，实际应该使用完整的 DPAPI 调用
	return data, nil
}

// DataUnprotection 使用 Windows DPAPI 解密数据
func DataUnprotection(data []byte) ([]byte, error) {
	// 使用 CryptUnprotectData API
	// 简化实现
	return data, nil
}
