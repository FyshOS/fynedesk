package main

import (
	"fyne.io/fyne/app"
	"fyne.io/fyne/theme"

	_ "fyne.io/fynedesk/modules/status"
)

func main() {
	a := app.NewWithID("io.fyne.fynedesk")
	a.SetIcon(theme.FyneLogo())
	desk := setupDesktop(a)

	desk.Run()
}
