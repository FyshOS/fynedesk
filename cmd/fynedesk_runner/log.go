package main

import (
	"io"
	"os"
	"path/filepath"

	"fyne.io/fyne"
)

func logPath() string {
	return logPathRelativeTo(systemLogDir())
}

func logPathRelativeTo(parent string) string {
	cacheDir := filepath.Join(parent, "fyne", "io.fyne.fynedesk")
	err := os.MkdirAll(cacheDir, 0700)
	if err != nil {
		fyne.LogError("Could not create log directory", err)
	}

	return filepath.Join(cacheDir, "fynedesk.log")
}

// openLogWriter returns a stream for stdOut and stdErr of the process we run
func openLogWriter() (io.WriteCloser, io.WriteCloser) {
	f, err := os.Create(logPath())
	if err != nil {
		fyne.LogError("Unable to open log file", err)
		return os.Stdout, os.Stderr
	}

	return f, f
}
