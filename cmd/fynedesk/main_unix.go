// +build linux openbsd freebsd netbsd

package main

import (
	"log"

	"fyne.io/fyne/v2"

	"fyne.io/fynedesk"
	"fyne.io/fynedesk/internal/icon"
	"fyne.io/fynedesk/internal/ui"
	"fyne.io/fynedesk/internal/x11/wm"
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
