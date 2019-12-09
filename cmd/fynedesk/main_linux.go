// +build !ci

package main

import (
	"log"

	"fyne.io/fyne"

	"fyne.io/desktop"
	"fyne.io/desktop/internal"
	"fyne.io/desktop/wm"
)

func setupDesktop(a fyne.App) desktop.Desktop {
	icons := internal.NewFDOIconProvider()
	mgr, screensProvider, err := wm.NewX11WindowManager(a)
	if err != nil {
		log.Println("Could not create window manager:", err)
		return desktop.NewEmbeddedDesktop(a, icons)
	}
	desk := desktop.NewDesktop(a, mgr, icons, screensProvider)
	mgr.SetRoot(desk.Root())
	return desk
}
