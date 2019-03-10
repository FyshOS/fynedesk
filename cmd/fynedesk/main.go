package main

import (
	"log"

	"fyne.io/fyne/app"

	"github.com/fyne-io/desktop"
)

func main() {
	a := app.New()
	desk := desktop.NewDesktop(a)
	if desk == nil {
		log.Println("Could not create window, exiting")
		return
	}

	desk.ShowAndRun()
}
