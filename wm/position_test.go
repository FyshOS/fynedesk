package wm

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"fyne.io/fynedesk"
	"fyne.io/fynedesk/test"
	"fyne.io/fynedesk/theme"
)

func TestPositionForNewWindow_TopLeft(t *testing.T) {
	screen := &fynedesk.Screen{X: 50, Y: 0, Width: 500, Height: 500, Scale: 1}
	x, y, _, _ := PositionForNewWindow(0, 0, 100, 100, true, test.NewScreensProvider(screen))

	assert.Equal(t, 50+theme.BorderWidth, x)
	assert.Equal(t, theme.TitleHeight, y)
}

func TestPositionForNewWindow_TopLeftBorderless(t *testing.T) {
	screen := &fynedesk.Screen{X: 50, Y: 0, Width: 500, Height: 500, Scale: 1}
	x, y, _, _ := PositionForNewWindow(0, 0, 100, 100, false, test.NewScreensProvider(screen))

	assert.Equal(t, 50, x)
	assert.Equal(t, 0, y)
}
