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
	x, y, _, _ := PositionForNewWindow(0, 0, 100, 100, test.NewScreensProvider(screen))

	assert.Equal(t, 50+theme.BorderWidth, x)
	assert.Equal(t, theme.TitleHeight, y)
}

func TestPositionForNewWindow_Set(t *testing.T) {
	xPos := 5
	yPos := 10

	x, y, _, _ := PositionForNewWindow(xPos, yPos, 100, 100, test.NewScreensProvider())
	assert.Equal(t, xPos, x)
	assert.Equal(t, yPos, y)
}
