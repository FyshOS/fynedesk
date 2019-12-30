package ui

import (
	"image/color"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"

	"fyne.io/desktop"
)

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

	onTapped      func()          // The function that will be called when the icon is clicked
	resource      fyne.Resource   // The image data of the image that the icon uses
	appData       desktop.AppData // The application data corresponding to this icon.
	taskbarWindow desktop.Window  // The window associated with this icon if it is a taskbar icon
}

//Tapped means barIcon has been clicked
func (bi *barIcon) Tapped(*fyne.PointEvent) {
	bi.onTapped()
}

//TappedSecondary means barIcon has been clicked by a secondary binding
func (bi *barIcon) TappedSecondary(*fyne.PointEvent) {
}

// CreateRenderer is a private method to fyne which links this widget to its renderer
func (bi *barIcon) CreateRenderer() fyne.WidgetRenderer {
	render := &barIconRenderer{image: bi}
	render.Refresh()

	return render
}

func newBarIcon(res fyne.Resource, appData desktop.AppData) *barIcon {
	barIcon := &barIcon{resource: res, appData: appData}
	barIcon.ExtendBaseWidget(barIcon)

	return barIcon
}
