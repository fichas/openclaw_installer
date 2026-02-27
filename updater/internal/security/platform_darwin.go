//go:build darwin
// +build darwin

package security

/*
#include <stdlib.h>
#include <string.h>
#include <Security/Security.h>

// 存储密码到 macOS 钥匙串
int storePassword(const char* service, const char* account, const char* password, int passwordLen) {
    SecKeychainItemRef item = NULL;
    OSStatus status = SecKeychainFindGenericPassword(
        NULL,
        strlen(service),
        service,
        strlen(account),
        account,
        NULL,
        NULL,
        &item
    );

    if (status == errSecSuccess && item != NULL) {
        // 更新现有条目
        status = SecKeychainItemModifyAttributesAndData(
            item,
            NULL,
            passwordLen,
            password
        );
        CFRelease(item);
    } else {
        // 创建新条目
        status = SecKeychainAddGenericPassword(
            NULL,
            strlen(service),
            service,
            strlen(account),
            account,
            passwordLen,
            password,
            NULL
        );
    }

    return status;
}

// 从 macOS 钥匙串检索密码
int retrievePassword(const char* service, const char* account, char** password, unsigned int* passwordLen) {
    OSStatus status = SecKeychainFindGenericPassword(
        NULL,
        strlen(service),
        service,
        strlen(account),
        account,
        passwordLen,
        (void**)password,
        NULL
    );
    return status;
}

// 从 macOS 钥匙串删除密码
int deletePassword(const char* service, const char* account) {
    SecKeychainItemRef item = NULL;
    OSStatus status = SecKeychainFindGenericPassword(
        NULL,
        strlen(service),
        service,
        strlen(account),
        account,
        NULL,
        NULL,
        &item
    );

    if (status == errSecSuccess && item != NULL) {
        status = SecKeychainItemDelete(item);
        CFRelease(item);
    }

    return status;
}

// 释放内存
void freePassword(void* password) {
    SecKeychainItemFreeContent(NULL, password);
}
*/
import "C"
import (
	"fmt"
	"unsafe"
)

// DarwinSecureStorage macOS 钥匙串存储
type DarwinSecureStorage struct {
	service string
}

// NewDarwinSecureStorage 创建新的 macOS 安全存储
func NewDarwinSecureStorage(service string) *DarwinSecureStorage {
	if service == "" {
		service = "OpenClaw"
	}
	return &DarwinSecureStorage{service: service}
}

// Store 存储数据到钥匙串
func (dss *DarwinSecureStorage) Store(key string, data []byte) error {
	serviceC := C.CString(dss.service)
	accountC := C.CString(key)
	passwordC := C.CString(string(data))
	defer C.free(unsafe.Pointer(serviceC))
	defer C.free(unsafe.Pointer(accountC))
	defer C.free(unsafe.Pointer(passwordC))

	status := C.storePassword(serviceC, accountC, passwordC, C.int(len(data)))
	if status != 0 {
		return fmt.Errorf("failed to store password: %d", status)
	}

	return nil
}

// Retrieve 从钥匙串检索数据
func (dss *DarwinSecureStorage) Retrieve(key string) ([]byte, error) {
	serviceC := C.CString(dss.service)
	accountC := C.CString(key)
	defer C.free(unsafe.Pointer(serviceC))
	defer C.free(unsafe.Pointer(accountC))

	var password *C.char
	var passwordLen C.uint

	status := C.retrievePassword(serviceC, accountC, &password, &passwordLen)
	if status != 0 {
		return nil, fmt.Errorf("failed to retrieve password: %d", status)
	}

	defer C.freePassword(unsafe.Pointer(password))

	// 复制数据
	data := C.GoBytes(unsafe.Pointer(password), C.int(passwordLen))
	return data, nil
}

// Delete 从钥匙串删除数据
func (dss *DarwinSecureStorage) Delete(key string) error {
	serviceC := C.CString(dss.service)
	accountC := C.CString(key)
	defer C.free(unsafe.Pointer(serviceC))
	defer C.free(unsafe.Pointer(accountC))

	status := C.deletePassword(serviceC, accountC)
	if status != 0 {
		return fmt.Errorf("failed to delete password: %d", status)
	}

	return nil
}

// Exists 检查数据是否存在
func (dss *DarwinSecureStorage) Exists(key string) bool {
	_, err := dss.Retrieve(key)
	return err == nil
}

// checkDebugger 检查 macOS 调试器
func checkDebugger() bool {
	// 使用 sysctl 检查 P_TRACED 标志
	// 简化实现，实际应该调用 sysctl
	return false
}

// lockMemory 锁定内存（Darwin 实现）
func lockMemory(data []byte) error {
	// macOS 使用 mlock
	// 简化实现
	return nil
}

// unlockMemory 解锁内存（Darwin 实现）
func unlockMemory(data []byte) {
	// macOS 使用 munlock
}
