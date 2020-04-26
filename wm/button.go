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

func newCloseButton(win fynedesk.Window) *closeButton {
	b := &closeButton{}
	b.ExtendBaseWidget(b)
	b.OnTapped = func() {
		win.Close()
	}

	b.Icon = theme.CancelIcon()
	b.Style = widget.PrimaryButton

	return b
}
