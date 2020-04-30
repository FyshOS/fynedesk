// +build linux

package x11

import (
	"fyne.io/fynedesk"
	wmTheme "fyne.io/fynedesk/theme"
)

// BorderWidth is the number of pixels required for a border
func BorderWidth(win XWin) uint16 {
	if !win.Properties().Decorated() {
		return 0
	}
	return ScaleToPixels(wmTheme.BorderWidth, fynedesk.Instance().Screens().ScreenForWindow(win))
}

// ButtonWidth is the number of pixels required for a border button
func ButtonWidth(win XWin) uint16 {
	return ScaleToPixels(wmTheme.ButtonWidth, fynedesk.Instance().Screens().ScreenForWindow(win))
}

// TitleHeight is the number of pixels required for a title bar
func TitleHeight(win XWin) uint16 {
	return ScaleToPixels(wmTheme.TitleHeight, fynedesk.Instance().Screens().ScreenForWindow(win))
}

// ScaleToPixels calculates the pixels required to show a specified Fyne dimension on the given screen
func ScaleToPixels(i int, screen *fynedesk.Screen) uint16 {
	return uint16(float32(i) * screen.CanvasScale())
}
