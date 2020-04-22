package ui

import "fyne.io/fyne"

func (l *desktop) newDesktopWindowEmbed(_ string) fyne.Window {
	win := l.app.NewWindow("Embedded " + RootWindowName)
	win.SetPadded(false)
	return win
}

func (l *desktop) runEmbed() {
	l.roots[0].ShowAndRun()
}
