// +build linux

package x11

import (
	"fyne.io/fynedesk"
	wmTheme "fyne.io/fynedesk/theme"
	"fyne.io/fynedesk/wm"
)

// BorderWidth is the number of pixels required for a border
func BorderWidth(win XWin) uint {
	if !win.Properties().Decorated() {
		return 0
	}
	return uint(wm.ScaleToPixels(wmTheme.BorderWidth, fynedesk.Instance().Screens().ScreenForWindow(win)))
}

// ButtonWidth is the number of pixels required for a border button
func ButtonWidth(win XWin) uint {
	return uint(wm.ScaleToPixels(wmTheme.ButtonWidth, fynedesk.Instance().Screens().ScreenForWindow(win)))
}

// TitleHeight is the number of pixels required for a title bar
func TitleHeight(win XWin) uint {
	return uint(wm.ScaleToPixels(wmTheme.TitleHeight, fynedesk.Instance().Screens().ScreenForWindow(win)))
}
