package wm

import (
	"fyne.io/desktop"
	wmTheme "fyne.io/desktop/theme"
	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
)

var iconSize = 32

func makeFiller() fyne.CanvasObject {
	filler := canvas.NewRectangle(theme.BackgroundColor()) // make a border on the X axis only
	filler.SetMinSize(fyne.NewSize(0, 2))                  // 0 wide forced

	return filler
}

func newBorder(win desktop.Window, icon fyne.Resource) fyne.CanvasObject {
	desk := desktop.Instance()

	if icon == nil {
		iconTheme := desk.Settings().IconTheme()
		app := desk.IconProvider().FindAppFromWinInfo(win)
		if app != nil {
			icon = app.Icon(iconTheme, iconSize)
		}
	}

	exit := widget.NewButtonWithIcon("", theme.CancelIcon(), func() {})
	exit.Style = widget.PrimaryButton
	max := widget.NewButtonWithIcon("", wmTheme.MaximizeIcon, func() {})
	if win.Maximized() {
		max.Icon = theme.ViewRestoreIcon()
	}
	if windowSizeFixed(win.(*client).wm.x, win.(*client).win) ||
		!windowSizeCanMaximize(win.(*client).wm.x, win.(*client).win,
			desktop.Instance().Screens().ScreenForWindow(win)) {
		max.Disable()
	}
	titleBar := widget.NewHBox(makeFiller(),
		exit,
		max,
		widget.NewButtonWithIcon("", wmTheme.IconifyIcon, func() {}),
		widget.NewLabel(win.Title()),
		layout.NewSpacer())

	if icon != nil {
		appIcon := canvas.NewImageFromResource(icon)
		appIcon.SetMinSize(fyne.NewSize(wmTheme.TitleHeight, wmTheme.TitleHeight))
		titleBar.Append(appIcon)
		titleBar.Append(makeFiller())
	}

	return fyne.NewContainerWithLayout(layout.NewBorderLayout(titleBar, nil, nil, nil),
		titleBar)
}
