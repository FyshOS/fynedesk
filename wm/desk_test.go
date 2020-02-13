package wm

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"fyne.io/desktop/internal/ui"
)

func TestScreenNameFromRootTitle(t *testing.T) {
	assert.Equal(t, "", screenNameFromRootTitle("NotARoot"))
	assert.Equal(t, "Primary", screenNameFromRootTitle(ui.RootWindowName+"Primary"))
	assert.Equal(t, "Screen 1", screenNameFromRootTitle(ui.RootWindowName+"Screen 1"))
}
