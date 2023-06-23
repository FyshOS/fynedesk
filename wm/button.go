package wm

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"fyshos.com/fynedesk"
)

const (
	// CloseCursor is the mouse cursor that indicates a close action
	CloseCursor desktop.StandardCursor = iota + desktop.VResizeCursor // add to the end of the fyne list
)

type closeButton struct {
	widget.Button

	bg *canvas.Rectangle
}

func (c *closeButton) Cursor() desktop.Cursor {
	return CloseCursor
}

func (c *closeButton) MouseIn(*desktop.MouseEvent) {
	c.bg.FillColor = theme.ErrorColor()
	c.bg.Refresh()
}

func (c *closeButton) MouseMoved(*desktop.MouseEvent) {
}

func (c *closeButton) MouseOut() {
	c.bg.FillColor = color.Transparent
	c.bg.Refresh()
}

func newCloseButton(win fynedesk.Window) fyne.CanvasObject {
	b := &closeButton{}
	b.ExtendBaseWidget(b)
	b.Importance = widget.LowImportance
	b.bg = canvas.NewRectangle(color.Transparent)
	b.bg.CornerRadius = theme.InputRadiusSize()

	b.OnTapped = func() {
		win.Close()
	}

	b.Icon = theme.CancelIcon()
	return container.NewMax(b.bg, b)
}
