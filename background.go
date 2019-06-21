package desktop

import (
	"os"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"

	wmtheme "github.com/fyne-io/desktop/theme"
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

func newBackground() fyne.CanvasObject {
	imagePath := wallpaperPath()
	if imagePath != "" {
		return canvas.NewImageFromFile(imagePath)
	}
	return canvas.NewImageFromResource(wmtheme.Background)
}
