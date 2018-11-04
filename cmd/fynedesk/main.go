package main

import "github.com/fyne-io/desktop"
import "github.com/fyne-io/desktop/cmd/fynedesk/driver"

func main() {
	app := driver.NewApp()
	desk := desktop.NewDesktop(app)

	desk.ShowAndRun()
}
