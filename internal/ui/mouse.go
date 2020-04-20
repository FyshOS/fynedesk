package ui

import (
	"fyne.io/fyne"
	"fyne.io/fyne/canvas"

	wmtheme "fyne.io/fynedesk/theme"
)

var mouse *canvas.Image

func newMouse() fyne.CanvasObject {
	mouse = canvas.NewImageFromResource(wmtheme.PointerDefault)
	mouse.Resize(fyne.NewSize(24, 24))

	return mouse
}
