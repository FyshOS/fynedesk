package main

import (
	"log"
	"runtime"

	"fyne.io/fyne/v2"

	"fyne.io/fynedesk"
	"fyne.io/fynedesk/internal/icon"
	"fyne.io/fynedesk/internal/ui"
)

func setupDesktop(a fyne.App) fynedesk.Desktop {
	log.Println("Full desktop not possible on", runtime.GOOS)
	return ui.NewEmbeddedDesktop(a, icon.NewMacOSAppProvider())
}
