package main

import (
	"log"

	"fyne.io/fyne"

	"fyne.io/desktop"
	"fyne.io/desktop/internal/icon"
	"fyne.io/desktop/internal/ui"
	"fyne.io/desktop/wm"
)

func setupDesktop(a fyne.App) desktop.Desktop {
	icons := icon.NewFDOIconProvider()
	mgr, err := wm.NewX11WindowManager(a)
	if err != nil {
		log.Println("Could not create window manager:", err)
		return ui.NewEmbeddedDesktop(a, icons)
	}
	desk := ui.NewDesktop(a, mgr, icons, wm.NewX11ScreensProvider(mgr))
	mgr.SetRoot(desk.Root())
	return desk
}
