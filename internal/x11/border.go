package x11

import (
	"fyne.io/fynedesk"
	wmTheme "fyne.io/fynedesk/theme"
)

func BorderWidth(win XWin) uint16 {
	if !win.Properties().Decorated() {
		return 0
	}
	return scaleToPixels(wmTheme.BorderWidth, fynedesk.Instance().Screens().ScreenForWindow(win))
}

func ButtonWidth(win XWin) uint16 {
	return scaleToPixels(wmTheme.ButtonWidth, fynedesk.Instance().Screens().ScreenForWindow(win))
}

func TitleHeight(win XWin) uint16 {
	return scaleToPixels(wmTheme.TitleHeight, fynedesk.Instance().Screens().ScreenForWindow(win))
}

func scaleToPixels(i int, screen *fynedesk.Screen) uint16 {
	return uint16(float32(i) * screen.CanvasScale())
}
