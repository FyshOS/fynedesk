package main

import (
	wmtheme "fyshos.com/fynedesk/theme"

	_ "fyshos.com/fynedesk/modules/composit"
	_ "fyshos.com/fynedesk/modules/desktops"
	_ "fyshos.com/fynedesk/modules/launcher"
	_ "fyshos.com/fynedesk/modules/status"
	_ "fyshos.com/fynedesk/modules/systray"

	"fyne.io/fyne/v2/app"
)

func main() {
	a := app.NewWithID("com.fyshos.fynedesk")
	a.SetIcon(wmtheme.AppIcon)
	desk := setupDesktop(a)

	desk.Run()
}
