//go:build linux || openbsd || freebsd || netbsd
// +build linux openbsd freebsd netbsd

package x11

import (
	"fyshos.com/fynedesk"
	wmTheme "fyshos.com/fynedesk/theme"
	"fyshos.com/fynedesk/wm"
)

// BorderWidth is the number of pixels required for a border
func BorderWidth(win XWin) uint16 {
	if !win.Properties().Decorated() {
		return 0
	}
	return uint16(wm.ScaleToPixels(wmTheme.BorderWidth, fynedesk.Instance().Screens().ScreenForWindow(win)))
}

// ButtonWidth is the number of pixels required for a border button
func ButtonWidth(win XWin) uint16 {
	return uint16(wm.ScaleToPixels(wmTheme.ButtonWidth, fynedesk.Instance().Screens().ScreenForWindow(win)))
}

// TitleHeight is the number of pixels required for a title bar
func TitleHeight(win XWin) uint16 {
	return uint16(wm.ScaleToPixels(wmTheme.TitleHeight, fynedesk.Instance().Screens().ScreenForWindow(win)))
}
