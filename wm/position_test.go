package wm

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"fyne.io/fynedesk"
	"fyne.io/fynedesk/test"
	"fyne.io/fynedesk/theme"
)

func TestPositionForNewWindow_TopLeft(t *testing.T) {
	screen := &fynedesk.Screen{Geometry: fynedesk.NewGeometry(50, 0, 500, 500), Scale: 1}
	g := PositionForNewWindow(fynedesk.NewGeometry(0, 0, 100, 100), test.NewScreensProvider(screen))

	assert.Equal(t, 50+theme.BorderWidth, g.X)
	assert.Equal(t, theme.TitleHeight, g.Y)
}

func TestPositionForNewWindow_Set(t *testing.T) {
	xPos := 5
	yPos := 10

	g := PositionForNewWindow(fynedesk.NewGeometry(xPos, yPos, 100, 100), test.NewScreensProvider())
	assert.Equal(t, xPos, g.X)
	assert.Equal(t, yPos, g.Y)
}
