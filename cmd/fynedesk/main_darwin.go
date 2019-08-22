package main

import (
	"log"
	"runtime"

	"fyne.io/fyne"

	"fyne.io/desktop"
	"fyne.io/desktop/internal"
)

func setupDesktop(a fyne.App) desktop.Desktop {
	log.Println("Full desktop not possible on", runtime.GOOS)
	return desktop.NewEmbeddedDesktop(a, internal.NewMacOSAppProvider())
}
