package ui

import (
	"runtime/debug"

	"fyne.io/fyne"
)

func (l *desktop) newDesktopWindowFull() fyne.Window {
	desk := l.app.NewWindow(RootWindowName)
	desk.SetPadded(false)
	desk.SetFullScreen(true)

	desk.SetMaster()
	desk.SetOnClosed(func() {
		l.wm.Close()
	})

	return desk
}

func (l *desktop) runFull() {
	debug.SetPanicOnFault(true)

	defer func() {
		if r := recover(); r != nil {
			debug.PrintStack()
			l.wm.Close() // attempt to close cleanly to leave X server running
		}
	}()

	l.root.ShowAndRun()
}
