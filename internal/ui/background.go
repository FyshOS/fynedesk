package ui

import (
	"os"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"

	"fyne.io/desktop"
	wmtheme "fyne.io/desktop/theme"
)

func updateBackgroundPath(background *canvas.Image, path string) {
	_, err := os.Stat(path)
	if path == "" || os.IsNotExist(err) {
		background.Resource = wmtheme.Background
		background.File = ""
		return
	}
	background.Resource = nil
	background.File = path
}

func backgroundPath() string {
	pathEnv := desktop.Instance().Settings().Background()
	if pathEnv == "" {
		return ""
	}

	if stat, err := os.Stat(pathEnv); os.IsNotExist(err) || !stat.Mode().IsRegular() {
		return ""
	}

	return pathEnv
}

func newBackground() fyne.CanvasObject {
	imagePath := backgroundPath()
	if imagePath != "" {
		return canvas.NewImageFromFile(imagePath)
	}
	return canvas.NewImageFromResource(wmtheme.Background)
}
