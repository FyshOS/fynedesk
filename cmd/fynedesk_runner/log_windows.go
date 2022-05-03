//go:build windows
// +build windows

package main

import (
	"os"
	"path/filepath"
)

func systemLogDir() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, "AppData", "Local")
}
