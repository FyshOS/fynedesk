package main

import (
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/theme"

	_ "fyshos.com/fynedesk/modules/composit"
	_ "fyshos.com/fynedesk/modules/launcher"
	_ "fyshos.com/fynedesk/modules/status"
	_ "fyshos.com/fynedesk/modules/systray"
)

func main() {
	a := app.NewWithID("com.fyshos.fynedesk")
	a.SetIcon(theme.FyneLogo())
	desk := setupDesktop(a)

	desk.Run()
}
