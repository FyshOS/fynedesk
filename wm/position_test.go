package wm

import (
	"testing"

	"fyne.io/fynedesk"
	"fyne.io/fynedesk/test"
	"github.com/stretchr/testify/assert"
)

func TestPositionForNewWindow_Center(t *testing.T) {
	screen := &fynedesk.Screen{X: 50, Y: 0, Width: 200, Height: 200}
	x, y, _, _ := PositionForNewWindow(0, 0, 100, 100, test.NewScreensProvider(screen))

	assert.Equal(t, 100, x)
	assert.Equal(t, 50, y)
}

func TestPositionForNewWindow_Set(t *testing.T) {
	xPos := 5
	yPos := 10

	x, y, _, _ := PositionForNewWindow(xPos, yPos, 100, 100, test.NewScreensProvider())
	assert.Equal(t, xPos, x)
	assert.Equal(t, yPos, y)
}
