package wm

import (
	"fyne.io/fynedesk"
)

// PositionForNewWindow returns the suggested position for a new window of the given geometry.
// The screen list hints at available space, but normally list.Active() is the best.
func PositionForNewWindow(x, y int, w, h uint, list fynedesk.ScreenList) (int, int, uint, uint) {
	if x != 0 && y != 0 {
		return x, y, w, h
	}

	screenX := list.Active().X
	screenY := list.Active().Y
	screenW := list.Active().Width
	screenH := list.Active().Height

	offX := (screenW - int(w)) / 2
	offY := (screenH - int(h)) / 2
	return screenX + offX, screenY + offY, w, h
}
