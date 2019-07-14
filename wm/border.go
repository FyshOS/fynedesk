package wm

import (
	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
)

func newBorder(title string) fyne.CanvasObject {
	filler := canvas.NewRectangle(theme.BackgroundColor()) // make a border on the X axis only
	filler.SetMinSize(fyne.NewSize(0, 2)) // 0 wide forced
	titleBar := widget.NewHBox(filler,
		widget.NewButtonWithIcon("", theme.CancelIcon(), func() {}),
		widget.NewLabel(title))

	return fyne.NewContainerWithLayout(layout.NewBorderLayout(titleBar, nil, nil, nil),
		titleBar)
}
