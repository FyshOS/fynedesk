package ui

import (
	"fmt"
	"runtime/debug"

	"fyne.io/fyne"
)

func (l *desktop) newDesktopWindowFull(outputName string) fyne.Window {
	desk := l.app.NewWindow(fmt.Sprintf("%s%s", RootWindowName, outputName))
	desk.SetPadded(false)

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

	l.controlWin.ShowAndRun()
}
