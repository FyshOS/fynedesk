package main

import "github.com/fyne-io/desktop"
import "github.com/fyne-io/fyne/app"

func main() {
	app := app.New()
	desk := desktop.NewDesktop(app)

	desk.ShowAndRun()
}
