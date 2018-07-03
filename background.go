package desktop

import "image/color"

import "github.com/fyne-io/fyne"
import "github.com/fyne-io/fyne/canvas"
import "github.com/fyne-io/fyne/theme"

func stripesPattern(x, y, w, h int) color.RGBA {
	if x%20 == y%20 || (x+y)%20 == 0 {
		return theme.ButtonColor()
	}

	return theme.BackgroundColor()
}

func newBackground() fyne.CanvasObject {
	return canvas.NewRaster(stripesPattern)
}
