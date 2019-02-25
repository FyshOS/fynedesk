package desktop

import "image/color"

import "fyne.io/fyne"
import "fyne.io/fyne/canvas"
import "fyne.io/fyne/theme"

func stripesPattern(x, y, w, h int) color.Color {
	if x%20 == y%20 || (x+y)%20 == 0 {
		return theme.ButtonColor()
	}

	return theme.BackgroundColor()
}

func newBackground() fyne.CanvasObject {
	return canvas.NewRasterWithPixels(stripesPattern)
}
