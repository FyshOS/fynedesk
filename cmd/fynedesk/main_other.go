//go:build !linux && !darwin && !freebsd && !openbsd && !netbsd
// +build !linux,!darwin,!freebsd,!openbsd,!netbsd

package main

import (
	"log"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fynedesk"

	"fyne.io/fynedesk"
	"fyne.io/fynedesk/internal"
	"fyne.io/fynedesk/internal/ui"
)

func setupDesktop(a fyne.App) fynedesk.Desktop {
	log.Println("Full desktop not possible on", runtime.GOOS)
	return ui.NewEmbeddedDesktop(a, internal.NewFDOIconProvider())
}
