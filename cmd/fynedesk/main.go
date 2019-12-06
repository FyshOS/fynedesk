package main

import (
	"os"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
)

func startSettingsListener(app fyne.App, listener chan fyne.Settings) {
	for {
		_ = <-listener
		if os.Getenv("FYNE_DESK_RUNNER") != "" {
			os.Exit(1)
		}
	}
}

func main() {
	a := app.NewWithID("io.fyne.desktop")
	desk := setupDesktop(a)
	listener := make(chan fyne.Settings)
	a.Settings().AddChangeListener(listener)
	go startSettingsListener(a, listener)

	desk.Run()
}
