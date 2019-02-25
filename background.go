package desktop

import (
	"image/color"
	"os"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/theme"
)

func wallpaperPath() string {
	pathEnv := os.Getenv("FYNE_DESKTOP")
	if pathEnv == "" {
		return ""
	}

	if stat, err := os.Stat(pathEnv); os.IsNotExist(err) || !stat.Mode().IsRegular() {
		return ""
	}

	return pathEnv
}

func stripesPattern(x, y, w, h int) color.Color {
	if x%20 == y%20 || (x+y)%20 == 0 {
		return theme.ButtonColor()
	}

	return theme.BackgroundColor()
}

func newBackground() fyne.CanvasObject {
	imagePath := wallpaperPath()
	if imagePath != "" {
		return canvas.NewImageFromFile(imagePath)
	}
	return canvas.NewRasterWithPixels(stripesPattern)
}
