//go:build darwin
// +build darwin

package main

import (
	"os"
	"path/filepath"
)

func systemLogDir() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, "Library", "Logs")
}
