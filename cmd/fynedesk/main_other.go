//go:build !linux && !darwin && !freebsd && !openbsd && !netbsd
// +build !linux,!darwin,!freebsd,!openbsd,!netbsd

package main

import (
	"log"
	"runtime"

	"fyne.io/fyne/v2"
	"fyshos.com/fynedesk"

	"fyshos.com/fynedesk"
	"fyshos.com/fynedesk/internal"
	"fyshos.com/fynedesk/internal/ui"
)

func setupDesktop(a fyne.App) fynedesk.Desktop {
	log.Println("Full desktop not possible on", runtime.GOOS)
	return ui.NewEmbeddedDesktop(a, internal.NewFDOIconProvider())
}
