package wm

import (
	"fyne.io/fynedesk"
	"fyne.io/fynedesk/theme"
)

// PositionForNewWindow returns the suggested position for a new window of the given geometry.
// The screen list hints at available space, but normally list.Active() is the best.
func PositionForNewWindow(g fynedesk.Geometry, screens fynedesk.ScreenList) fynedesk.Geometry {
	if g.X != 0 && g.Y != 0 {
		return g
	}

	target := screens.Active()
	offX := ScaleToPixels(theme.BorderWidth, target)
	offY := ScaleToPixels(theme.TitleHeight, target)

	return fynedesk.NewGeometry(target.X+offX, target.Y+offY, g.Width, g.Height)
}
