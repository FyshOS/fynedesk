package desktop

import "fyne.io/fyne"
import "fyne.io/fyne/canvas"

import wmtheme "github.com/fyne-io/desktop/theme"

var mouse *canvas.Image

func newMouse() fyne.CanvasObject {
	mouse = canvas.NewImageFromResource(wmtheme.PointerDefault)

	if isEmbedded() {
		// hide the mouse cursor as the parent desktop will paint one
		mouse.Hidden = true
		return mouse
	}

	mouse.Resize(fyne.NewSize(24, 24))

	return mouse
}
