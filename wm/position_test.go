package wm

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"fyne.io/fynedesk"
	"fyne.io/fynedesk/test"
	"fyne.io/fynedesk/theme"
)

func TestPositionForNewWindow_Default(t *testing.T) {
	screen := &fynedesk.Screen{X: 50, Y: 0, Width: 500, Height: 500, Scale: 1}
	x, y, _, _ := PositionForNewWindow(0, 0, 100, 100, true, test.NewScreensProvider(screen))

	assert.Equal(t, 250-int(theme.BorderWidth), x)
	assert.Equal(t, 200-int(theme.TitleHeight), y)
}

func TestPositionForNewWindow_DefaultBorderless(t *testing.T) {
	screen := &fynedesk.Screen{X: 50, Y: 0, Width: 500, Height: 500, Scale: 1}
	x, y, _, _ := PositionForNewWindow(0, 0, 100, 100, false, test.NewScreensProvider(screen))

	assert.Equal(t, 250, x)
	assert.Equal(t, 200, y)
}
