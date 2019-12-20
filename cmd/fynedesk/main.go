package main

import (
	"fyne.io/fyne/app"
	"fyne.io/fyne/theme"
)

func main() {
	a := app.NewWithID("io.fyne.desktop")
	a.SetIcon(theme.FyneLogo())
	desk := setupDesktop(a)

	desk.Run()
}
