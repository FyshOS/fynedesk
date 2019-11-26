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
	desk, err := wm.NewX11WindowManager(a, icons)
	if err != nil {
		log.Println("Could not create window manager:", err)
		return desktop.NewEmbeddedDesktop(a, icons)
	}
	return desk
}
