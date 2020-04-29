package wm

import (
	"fyne.io/fynedesk"
	"fyne.io/fynedesk/theme"
)

// PositionForNewWindow returns the suggested position for a new window of the given geometry.
// The screen list hints at available space, but normally list.Active() is the best.
func PositionForNewWindow(x, y int, w, h uint, screens fynedesk.ScreenList) (int, int, uint, uint) {
	if x != 0 && y != 0 {
		return x, y, w, h
	}

	target := screens.Active()
	offX := ScaleToPixels(theme.BorderWidth, target)
	offY := ScaleToPixels(theme.TitleHeight, target)

	return target.X + offX, target.Y + offY, w, h
}
