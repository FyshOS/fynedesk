// +build !ci

package main

import (
	"log"

	"fyne.io/fyne"

	"github.com/fyne-io/desktop"
	"github.com/fyne-io/desktop/wm"
)

func setupDesktop(a fyne.App) desktop.Desktop {
	mgr, err := wm.NewX11WindowManager(a)
	if err != nil {
		log.Println("Could not create window manager:", err)
		return desktop.NewEmbeddedDesktop(a)
	}

	return desktop.NewDesktop(a, mgr)
}
