package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"

	wmtheme "fyne.io/fynedesk/theme"
)

var mouse *canvas.Image

func newMouse() fyne.CanvasObject {
	mouse = canvas.NewImageFromResource(wmtheme.PointerDefault)
	mouse.Resize(fyne.NewSize(24, 24))

	return mouse
}
