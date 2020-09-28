package wm

import (
	"fyne.io/fyne/driver/desktop"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	"fyne.io/fynedesk"
)

const (
	// CloseCursor is the mouse cursor that indicates a close action
	CloseCursor desktop.Cursor = iota + desktop.VResizeCursor // add to the end of the fyne list
)

type closeButton struct {
	widget.Button
}

func (c *closeButton) Cursor() desktop.Cursor {
	return CloseCursor
}

func (c *closeButton) MouseIn(*desktop.MouseEvent) {
	c.Style = widget.PrimaryButton
	c.Refresh()
}

func (c *closeButton) MouseMoved(*desktop.MouseEvent) {
}

func (c *closeButton) MouseOut() {
	c.Style = widget.DefaultButton
	c.Refresh()
}

func newCloseButton(win fynedesk.Window) *closeButton {
	b := &closeButton{}
	b.ExtendBaseWidget(b)
	b.HideShadow = true
	b.OnTapped = func() {
		win.Close()
	}

	b.Icon = theme.CancelIcon()
	return b
}
