package main

import efl "github.com/fyne-io/fyne/desktop"

import "github.com/fyne-io/desktop"

func main() {
	app := efl.NewApp()
	desk := desktop.NewDesktop(app)

	desk.Show()
}
