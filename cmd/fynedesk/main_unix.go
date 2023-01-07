//go:build linux || openbsd || freebsd || netbsd
// +build linux openbsd freebsd netbsd

package main

import (
	"log"

	"fyne.io/fyne/v2"

	"fyshos.com/fynedesk"
	"fyshos.com/fynedesk/internal/icon"
	"fyshos.com/fynedesk/internal/ui"
	"fyshos.com/fynedesk/internal/x11/wm"
)

func setupDesktop(a fyne.App) fynedesk.Desktop {
	icons := icon.NewFDOIconProvider()
	mgr, err := wm.NewX11WindowManager(a)
	if err != nil {
		log.Println("Could not create window manager:", err)
		return ui.NewEmbeddedDesktop(a, icons)
	}
	desk := ui.NewDesktop(a, mgr, icons, wm.NewX11ScreensProvider(mgr))
	return desk
}
