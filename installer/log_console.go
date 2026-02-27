//go:build !windows
// +build !windows

package main

// isWindowsGUI always returns false on non-Windows platforms
func isWindowsGUI() bool {
	return false
}
