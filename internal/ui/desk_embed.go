package ui

import "fyne.io/fyne/v2"

func (l *desktop) newDesktopWindowEmbed() fyne.Window {
	win := l.app.NewWindow("Embedded " + RootWindowName)
	win.SetPadded(false)
	win.Resize(fyne.NewSize(1024, 576)) // grow a little from the minimum for testing
	win.SetMaster()
	return win
}

func (l *desktop) runEmbed() {
	l.root.ShowAndRun()
}
