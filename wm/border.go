package wm

import (
	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
)

func newBorder(title string) fyne.CanvasObject {
	titleBar := widget.NewHBox(canvas.NewLine(theme.BackgroundColor()),
		widget.NewButton("x", func() {}), widget.NewLabel(title))

	return fyne.NewContainerWithLayout(layout.NewBorderLayout(titleBar, nil, nil, nil),
		titleBar)
}
