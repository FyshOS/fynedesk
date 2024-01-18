//go:build !darwin && !windows
// +build !darwin,!windows

package main

import (
	"os"
	"path/filepath"
)

func systemLogDir() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".cache")
}
