//go:build !windows && !darwin && !linux
// +build !windows,!darwin,!linux

package security

import "errors"

// checkDebugger 检查调试器（通用实现）
func checkDebugger() bool {
	return false
}

// lockMemory 锁定内存（通用实现）
func lockMemory(data []byte) error {
	return errors.New("memory locking not supported on this platform")
}

// unlockMemory 解锁内存（通用实现）
func unlockMemory(data []byte) {
	// 无操作
}
