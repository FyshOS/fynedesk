// +build !ci
// +build !efl

package desktop

import "log"

import "fyne.io/fyne"

// newDesktopWindow will return a new window based the current environment.
// When running in an existing desktop then load a window.
// Otherwise let's return a desktop root!
func newDesktopWindow(app fyne.App) fyne.Window {
	if isEmbedded() {
		desk = app.NewWindow("Fyne Desktop")
		desk.SetPadded(false)
		return desk
	}

	log.Fatalln("Cannot run a complete desktop environemnt without -tags efl")
	return nil
}

func initInput() {
}
