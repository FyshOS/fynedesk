package ui

import "fyne.io/fyne"
import "fyne.io/fyne/canvas"

import wmtheme "fyne.io/desktop/theme"

var mouse *canvas.Image

func newMouse() fyne.CanvasObject {
	mouse = canvas.NewImageFromResource(wmtheme.PointerDefault)
	mouse.Resize(fyne.NewSize(24, 24))

	return mouse
}
