package ui

import (
	"image/color"
	"os"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"

	"fyne.io/desktop"
	wmtheme "fyne.io/desktop/theme"
)

type background struct {
	widget.BaseWidget

	wallpaper     *canvas.Image
	wallpaperPath string
}

func (b *background) CreateRenderer() fyne.WidgetRenderer {
	mods := desktop.Instance().Modules()
	objects := []fyne.CanvasObject{b.wallpaper}

	for _, m := range mods {
		if deskMod, ok := m.(desktop.ScreenAreaModule); ok {
			wid := deskMod.ScreenAreaWidget()
			if wid == nil {
				continue
			}

			objects = append(objects, wid)
		}
	}

	c := fyne.NewContainerWithLayout(layout.NewMaxLayout(), objects...)
	return &backgroundRenderer{b: b, c: c}
}

type backgroundRenderer struct {
	b *background
	c *fyne.Container
}

func (b *backgroundRenderer) Layout(s fyne.Size) {
	b.c.Layout.Layout(b.c.Objects, s)
}

func (b *backgroundRenderer) MinSize() fyne.Size {
	return b.c.Layout.MinSize(b.c.Objects)
}

func (b *backgroundRenderer) Refresh() {
}

func (b *backgroundRenderer) BackgroundColor() color.Color {
	return theme.BackgroundColor()
}

func (b *backgroundRenderer) Objects() []fyne.CanvasObject {
	return b.c.Objects
}

func (b *backgroundRenderer) Destroy() {
}

func (b *background) updateBackgroundPath(path string) {
	_, err := os.Stat(path)
	if path == "" || os.IsNotExist(err) {
		b.wallpaper.Resource = wmtheme.Background
		b.wallpaper.File = ""
		return
	}

	b.wallpaper.Resource = nil
	b.wallpaper.File = path
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

func newBackground() *background {
	var bg *canvas.Image
	imagePath := backgroundPath()
	if imagePath != "" {
		bg = canvas.NewImageFromFile(imagePath)
	} else {
		bg = canvas.NewImageFromResource(wmtheme.Background)
	}

	ret := &background{wallpaper: bg}
	ret.ExtendBaseWidget(ret)
	return ret
}
