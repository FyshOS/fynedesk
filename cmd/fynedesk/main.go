package main

import (
	"fyne.io/fyne/app"
)

func main() {
	a := app.NewWithID("io.fyne.desktop")
	desk := setupDesktop(a)

	desk.Run()
}
