package ui

import "fyne.io/fyne"

func (l *desktop) newDesktopWindowEmbed(_ string) fyne.Window {
	win := l.app.NewWindow("Embedded " + RootWindowName)
	win.SetPadded(false)
	win.Resize(fyne.NewSize(1024, 576)) // grow a little from the minimum for testing
	return win
}

func (l *desktop) runEmbed() {
	l.roots[0].ShowAndRun()
}
