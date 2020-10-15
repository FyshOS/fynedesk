package wm

import (
	"log"

	"fyne.io/fyne"
)

func (x *x11WM) screenshot() {
	log.Println("Trying to print screen")

}

func (x *x11WM) screenshotWindow() {
	win := x.stack.TopWindow()
	if win == nil {
		fyne.LogError("Unable to print window with no window visible", nil)
		return
	}

	log.Println("Trying to print window:", win.Properties().Title())
}
