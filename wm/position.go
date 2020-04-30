package wm

import (
	"fyne.io/fynedesk"
)

// PositionForNewWindow returns the suggested position for a new window of the given geometry.
// The screen list hints at available space, but normally list.Active() is the best.
func PositionForNewWindow(g fynedesk.Geometry, list fynedesk.ScreenList) fynedesk.Geometry {
	if g.X != 0 && g.Y != 0 {
		return g
	}

	screenGeom := list.Active().Geometry
	offX := (int(screenGeom.Width) - int(g.Width)) / 2
	offY := (int(screenGeom.Height) - int(g.Height)) / 2
	return fynedesk.NewGeometry(screenGeom.X+offX, screenGeom.Y+offY, g.Width, g.Height)
}
