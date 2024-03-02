package ui

import (
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"fyshos.com/fynedesk"
	"github.com/FyshOS/backgrounds/builtin"
)

type background struct {
	widget.BaseWidget

	wallpaper *fyne.Container
}

func (b *background) CreateRenderer() fyne.WidgetRenderer {
	c := container.NewStack(b.loadModules()...)
	return widget.NewSimpleRenderer(c)
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

	return objects
}

func (b *background) updateBackground(path string) {
	_, err := os.Stat(path)
	if path == "" || os.IsNotExist(err) {
		set := fyne.CurrentApp().Settings()
		src := &builtin.Builtin{}
		b.wallpaper.Objects[0] = src.Load(set.Theme(), set.ThemeVariant())
	} else {
		bg := canvas.NewImageFromFile(path)
		bg.ScaleMode = canvas.ImageScaleFastest
		b.wallpaper.Objects[0] = bg
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
	var bg fyne.CanvasObject
	imagePath := backgroundPath()
	if imagePath != "" {
		img := canvas.NewImageFromFile(imagePath)
		img.ScaleMode = canvas.ImageScaleFastest
		bg = img
	} else {
		set := fyne.CurrentApp().Settings()
		b := &builtin.Builtin{}
		bg = b.Load(set.Theme(), set.ThemeVariant())
	}

	ret := &background{wallpaper: container.NewStack(bg)}
	ret.ExtendBaseWidget(ret)
	return ret
}
