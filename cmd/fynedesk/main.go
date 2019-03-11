package main

import (
	"fyne.io/fyne/app"
)

func main() {
	a := app.New()
	desk := setupDesktop(a)

	desk.Run()
}
