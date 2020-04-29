package wm

import (
	"testing"

	"fyne.io/fynedesk"
	"fyne.io/fynedesk/test"
	"github.com/stretchr/testify/assert"
)

func TestPositionForNewWindow_Center(t *testing.T) {
	screen := &fynedesk.Screen{Geometry: fynedesk.NewGeometry(50, 0, 200, 200)}
	g := PositionForNewWindow(fynedesk.NewGeometry(0, 0, 100, 100), test.NewScreensProvider(screen))

	assert.Equal(t, 100, g.X)
	assert.Equal(t, 50, g.Y)
}

func TestPositionForNewWindow_Set(t *testing.T) {
	xPos := 5
	yPos := 10

	g := PositionForNewWindow(fynedesk.NewGeometry(xPos, yPos, 100, 100), test.NewScreensProvider())
	assert.Equal(t, xPos, g.X)
	assert.Equal(t, yPos, g.Y)
}
