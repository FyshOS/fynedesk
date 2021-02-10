package ui

import (
	"runtime/debug"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

func (l *desktop) newDesktopWindowFull() fyne.Window {
	desk := l.app.NewWindow(RootWindowName)
	desk.SetPadded(false)

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

func (l *desktop) showMenuFull(menu *fyne.Menu, pos fyne.Position) {
	height := widget.NewMenu(menu).MinSize().Height
	size := fyne.NewSize(widgetPanelWidth, height)
	l.wm.ShowMenuOverlay(menu, size, pos)
}
