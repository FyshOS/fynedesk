package main

import (
	"fyne.io/fyne/app"
	"fyne.io/fyne/theme"

	_ "fyne.io/fynedesk/modules/builtin"
	_ "fyne.io/fynedesk/modules/christmas"
)

func main() {
	a := app.NewWithID("io.fyne.fynedesk")
	a.SetIcon(theme.FyneLogo())
	desk := setupDesktop(a)

	desk.Run()
}
