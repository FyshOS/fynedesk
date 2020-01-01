package main

import (
	"log"
	"runtime"

	"fyne.io/fyne"

	"fyne.io/desktop"
	"fyne.io/desktop/internal/icon"
	"fyne.io/desktop/internal/ui"
)

func setupDesktop(a fyne.App) desktop.Desktop {
	log.Println("Full desktop not possible on", runtime.GOOS)
	return ui.NewEmbeddedDesktop(a, icon.NewMacOSAppProvider())
}
