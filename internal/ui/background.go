package ui

import (
	"image/color"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"fyne.io/fynedesk"
	wmtheme "fyne.io/fynedesk/theme"
)

type background struct {
	widget.BaseWidget

	objects   []fyne.CanvasObject
	wallpaper *canvas.Image
}

func (b *background) CreateRenderer() fyne.WidgetRenderer {
	c := container.NewMax(b.loadModules()...)
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
	b.c.Objects = b.b.objects
}

func (b *backgroundRenderer) BackgroundColor() color.Color {
	return theme.BackgroundColor()
}

func (b *backgroundRenderer) Objects() []fyne.CanvasObject {
	return b.c.Objects
}

func (b *backgroundRenderer) Destroy() {
}

func (b *background) loadModules() []fyne.CanvasObject {
	objects := []fyne.CanvasObject{b.wallpaper}

	for _, m := range fynedesk.Instance().Modules() {
		if deskMod, ok := m.(fynedesk.ScreenAreaModule); ok {
			wid := deskMod.ScreenAreaWidget()
			if wid == nil {
				continue
			}

			objects = append(objects, wid)
		}
	}

	b.objects = objects
	return objects
}

func (b *background) updateBackground(path string) {
	_, err := os.Stat(path)
	if path == "" || os.IsNotExist(err) {
		b.wallpaper.Resource = wmtheme.Background
		b.wallpaper.File = ""
	} else {
		b.wallpaper.Resource = nil
		b.wallpaper.File = path
	}
	b.loadModules()
	canvas.Refresh(b.wallpaper)
	b.Refresh()
}

func backgroundPath() string {
	pathEnv := fynedesk.Instance().Settings().Background()
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
	bg.ScaleMode = canvas.ImageScaleFastest

	ret := &background{wallpaper: bg}
	ret.ExtendBaseWidget(ret)
	return ret
}
