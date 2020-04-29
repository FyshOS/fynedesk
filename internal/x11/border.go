// +build linux

package x11

import (
	"fyne.io/fynedesk"
	wmTheme "fyne.io/fynedesk/theme"
)

// BorderWidth is the number of pixels required for a border
func BorderWidth(win XWin) uint {
	if !win.Properties().Decorated() {
		return 0
	}
	return uint(ScaleToPixels(wmTheme.BorderWidth, fynedesk.Instance().Screens().ScreenForWindow(win)))
}

// ButtonWidth is the number of pixels required for a border button
func ButtonWidth(win XWin) uint {
	return uint(ScaleToPixels(wmTheme.ButtonWidth, fynedesk.Instance().Screens().ScreenForWindow(win)))
}

// TitleHeight is the number of pixels required for a title bar
func TitleHeight(win XWin) uint {
	return uint(ScaleToPixels(wmTheme.TitleHeight, fynedesk.Instance().Screens().ScreenForWindow(win)))
}

// ScaleToPixels calculates the pixels required to show a specified Fyne dimension on the given screen
func ScaleToPixels(i int, screen *fynedesk.Screen) int {
	return int(float32(i) * screen.CanvasScale())
}
