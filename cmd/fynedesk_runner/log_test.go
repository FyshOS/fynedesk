package main

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testLogDir() string {
	return filepath.Join("testdata", "cache")
}

func TestLogPath(t *testing.T) {
	path := logPathRelativeTo(testLogDir())

	assert.Equal(t, "testdata/cache/fyne/io.fyne.fynedesk/fynedesk.log", path)
}
