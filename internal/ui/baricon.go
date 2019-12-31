package ui

import (
	"image/color"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"

	"fyne.io/desktop"
)

// appWindow describes a type of icon that refers to an open window rather than an app.
// The findApp function can be used to attempt looking up the application from it's window.
type appWindow struct {
	win desktop.Window
	bar *bar
}

// findApp will try to return an application data associated with a window.
// This may fail for many reasons, usually related too bad window metadata, and will then return nil.
func (a *appWindow) findApp() desktop.AppData {
	if a.win == nil {
		return nil
	}

	return a.bar.desk.IconProvider().FindAppFromWinInfo(a.win)
}

type barIconRenderer struct {
	objects []fyne.CanvasObject

	image *barIcon
}

func (bi *barIconRenderer) MinSize() fyne.Size {
	size := theme.IconInlineSize()
	return fyne.NewSize(size, size)
}

func (bi *barIconRenderer) Layout(size fyne.Size) {
	if len(bi.objects) == 0 {
		return
	}

	bi.objects[0].Resize(size)
}

func (bi *barIconRenderer) Objects() []fyne.CanvasObject {
	return bi.objects
}

func (bi *barIconRenderer) BackgroundColor() color.Color {
	return color.Transparent
}

func (bi *barIconRenderer) Refresh() {
	bi.objects = nil

	if bi.image.resource != nil {
		raster := canvas.NewImageFromResource(bi.image.resource)
		raster.FillMode = canvas.ImageFillContain

		bi.objects = []fyne.CanvasObject{raster}
	}
	bi.Layout(bi.image.Size())

	canvas.Refresh(bi.image)
}

func (bi *barIconRenderer) Destroy() {
}

// barIcon widget is a basic image component that load's its resource to match the theme.
type barIcon struct {
	widget.BaseWidget

	onTapped   func()          // The function that will be called when the icon is clicked
	resource   fyne.Resource   // The image data of the image that the icon uses
	appData    desktop.AppData // The application data corresponding to this icon.(if it is a launcher)
	windowData *appWindow      // The window data associated with this icon (if it is a task window)
}

//Tapped means barIcon has been clicked
func (bi *barIcon) Tapped(*fyne.PointEvent) {
	bi.onTapped()
}

func addToBar(icon desktop.AppData) {
	settings := desktop.Instance().Settings()
	icons := settings.LauncherIcons()
	icons = append(icons, icon.Name())

	settings.(*deskSettings).setLauncherIcons(icons)
}

func removeFromBar(icon desktop.AppData) {
	settings := desktop.Instance().Settings()
	icons := settings.LauncherIcons()

	index := -1
	for i, defaultApp := range icons {
		if defaultApp == icon.Name() {
			index = i
			break
		}
	}
	if index >= 0 {
		icons = append(icons[:index], icons[index+1:]...)
	}
	settings.(*deskSettings).setLauncherIcons(icons)
}

//TappedSecondary means barIcon has been clicked by a secondary binding
func (bi *barIcon) TappedSecondary(*fyne.PointEvent) {
	app := bi.appData
	if app == nil && bi.windowData != nil {
		app = bi.windowData.findApp()
	}
	if app == nil || app.Name() == "" {
		return
	}

	var menu *widget.PopUp
	addRemove := fyne.NewMenuItem("Remove "+app.Name(), func() {
		if bi.windowData != nil {
			addToBar(app)
		} else {
			removeFromBar(app)
		}
		menu.Hide()
	})

	if bi.windowData != nil {
		addRemove.Label = "Pin " + app.Name()
	}

	c := fyne.CurrentApp().Driver().CanvasForObject(bi)
	pos := fyne.CurrentApp().Driver().AbsolutePositionForObject(bi)
	menu = widget.NewPopUpMenuAtPosition(fyne.NewMenu("", []*fyne.MenuItem{addRemove}...), c, pos)
}

// CreateRenderer is a private method to fyne which links this widget to its renderer
func (bi *barIcon) CreateRenderer() fyne.WidgetRenderer {
	render := &barIconRenderer{image: bi}
	render.Refresh()

	return render
}

func newBarIcon(res fyne.Resource, appData desktop.AppData, winData *appWindow) *barIcon {
	barIcon := &barIcon{resource: res, appData: appData, windowData: winData}
	barIcon.ExtendBaseWidget(barIcon)

	return barIcon
}
