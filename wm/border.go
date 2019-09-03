package wm

import (
	"fyne.io/desktop"
	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"

	wmTheme "fyne.io/desktop/theme"
)

var iconSize = 32

func makeFiller() fyne.CanvasObject {
	filler := canvas.NewRectangle(theme.BackgroundColor()) // make a border on the X axis only
	filler.SetMinSize(fyne.NewSize(0, 2))                  // 0 wide forced

	return filler
}

func newBorder(win desktop.Window) fyne.CanvasObject {
	desk := desktop.Instance()
	iconTheme := desk.Settings().IconTheme()
	app := desk.IconProvider().FindAppFromWinInfo(win)
	icon := app.Icon(iconTheme, iconSize)
	titleBar := widget.NewHBox(makeFiller(),
		widget.NewButtonWithIcon("", theme.CancelIcon(), func() {}),
		widget.NewButtonWithIcon("", wmTheme.MaximizeIcon, func() {}),
		widget.NewButtonWithIcon("", wmTheme.IconifyIcon, func() {}),
		widget.NewLabel(win.Title()),
		layout.NewSpacer())

	if icon != nil {
		titleBar.Append(widget.NewIcon(icon))
		titleBar.Append(makeFiller())
	}

	return fyne.NewContainerWithLayout(layout.NewBorderLayout(titleBar, nil, nil, nil),
		titleBar)
}
