package wm

import (
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"fyne.io/fynedesk"
)

const (
	// CloseCursor is the mouse cursor that indicates a close action
	CloseCursor desktop.StandardCursor = iota + desktop.VResizeCursor // add to the end of the fyne list
)

type closeButton struct {
	widget.Button
}

func (c *closeButton) Cursor() desktop.Cursor {
	return CloseCursor
}

func (c *closeButton) MouseIn(*desktop.MouseEvent) {
	c.Importance = widget.HighImportance
	c.Refresh()
}

func (c *closeButton) MouseMoved(*desktop.MouseEvent) {
}

func (c *closeButton) MouseOut() {
	c.Importance = widget.MediumImportance
	c.Refresh()
}

func newCloseButton(win fynedesk.Window) *closeButton {
	b := &closeButton{}
	b.ExtendBaseWidget(b)
	b.Importance = widget.LowImportance
	b.OnTapped = func() {
		win.Close()
	}

	b.Icon = theme.CancelIcon()
	return b
}
