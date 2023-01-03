//go:build linux || openbsd || freebsd || netbsd
// +build linux openbsd freebsd netbsd

package wm

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"fyshos.com/fynedesk/internal/ui"
)

func TestScreenNameFromRootTitle(t *testing.T) {
	assert.Equal(t, "", screenNameFromRootTitle("NotARoot"))
	assert.Equal(t, "Primary", screenNameFromRootTitle(ui.RootWindowName+"Primary"))
	assert.Equal(t, "Screen 1", screenNameFromRootTitle(ui.RootWindowName+"Screen 1"))
}
