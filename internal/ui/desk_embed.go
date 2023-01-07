package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	wmtheme "fyshos.com/fynedesk/theme"
)

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

func (l *desktop) showMenuEmbed(menu *fyne.Menu, pos fyne.Position) {
	wid := widget.NewPopUpMenu(menu, l.root.Canvas())
	wid.Resize(fyne.NewSize(wmtheme.WidgetPanelWidth, wid.MinSize().Height))
	wid.ShowAtPosition(pos)
}
