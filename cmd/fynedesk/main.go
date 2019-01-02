package main

import "fyne.io/fyne/app"

import "github.com/fyne-io/desktop"

func main() {
	app := app.New()
	desk := desktop.NewDesktop(app)

	desk.ShowAndRun()
}
