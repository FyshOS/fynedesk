package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
)

func crashLogPath() string {
	return crashLogPathRelativeTo(systemLogDir())
}

func crashLogPathRelativeTo(parent string) string {
	path := filepath.Join(logDir(parent), "fynedesk")
	now := time.Now().Format(time.RFC3339)
	return fmt.Sprintf("%s-crash-%s.log", path, now)
}

func logDir(parent string) string {
	cacheDir := filepath.Join(parent, "fyne", "com.fyshos.fynedesk")
	err := os.MkdirAll(cacheDir, 0700)
	if err != nil {
		fyne.LogError("Could not create log directory", err)
	}

	return cacheDir
}

func logPath() string {
	return logPathRelativeTo(systemLogDir())
}

func logPathRelativeTo(parent string) string {
	return filepath.Join(logDir(parent), "fynedesk.log")
}

// openLogWriter returns the log file that can be used to write stdOut and
// stdErr of the process we run
func openLogWriter() *os.File {
	f, err := os.Create(logPath())
	if err != nil {
		fyne.LogError("Unable to open log file", err)
		return os.Stderr
	}

	return f
}

func openRunnerLogWriter() *os.File {
	f, err := os.Create(runnerLogPath())
	if err != nil {
		fyne.LogError("Unable to open log file", err)
		return os.Stderr
	}

	return f
}

func runnerLogPath() string {
	return filepath.Join(logDir(systemLogDir()), "fynedesk_runner.log")
}
