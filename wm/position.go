package wm

import (
	"fyshos.com/fynedesk"
	"fyshos.com/fynedesk/theme"
)

// PositionForNewWindow returns the suggested position for a new window of the given geometry.
// The screen list hints at available space, but normally list.Active() is the best.
func PositionForNewWindow(win fynedesk.Window, x, y int, w, h uint, decorated bool,
	screens fynedesk.ScreenList) (int, int, uint, uint) {

	target := screens.Active()
	var offX, offY int
	parent := win.Parent()
	if parent != nil {
		wx, wy, ww, wh := parent.(interface{ Geometry() (int, int, uint, uint) }).Geometry() // avoid XWin import cycle
		offX, offY = positionInRect(w, h, wx, wy, ww, wh)
	} else {
		offX, offY = positionInRect(w, h, target.X, target.Y, uint(target.Width), uint(target.Height))
	}
	if decorated {
		offX -= ScaleToPixels(theme.BorderWidth, target)
		offY -= ScaleToPixels(theme.TitleHeight, target)
	}

	return offX, offY, w, h
}

func positionInRect(ww, hh uint, x, y int, w, h uint) (int, int) {
	return x + (int(w)-int(ww))/2, y + (int(h)-int(hh))/2
}
