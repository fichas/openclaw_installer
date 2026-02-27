//go:build windows
// +build windows

package main

import (
	"os"
)

func init() {
	// Redirect stderr to null on Windows to prevent console window from showing
	// This works in conjunction with -H=windowsgui ldflag
	if isWindowsGUI() {
		// Suppress any remaining console output
		null, _ := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
		os.Stderr = null
		os.Stdout = null
	}
}

// isWindowsGUI returns true on Windows when built with -H=windowsgui
func isWindowsGUI() bool {
	// Check if we have a console window
	// If not, we're running as a GUI application
	return true // On Windows, assume GUI mode by default
}
