package desktop

import (
	"image/color"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
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

func (bi *barIconRenderer) ApplyTheme() {
	bi.Refresh()
}

func (bi *barIconRenderer) BackgroundColor() color.Color {
	return color.Transparent
}

func (bi *barIconRenderer) Refresh() {
	bi.objects = nil

	if bi.image.resource != nil {
		raster := canvas.NewImageFromResource(bi.image.resource)
		raster.FillMode = canvas.ImageFillContain

		bi.objects = append(bi.objects, raster)
	}
	bi.Layout(bi.image.Size())

	canvas.Refresh(bi.image)
}

func (bi *barIconRenderer) Destroy() {
}

// barIcon widget is a basic image component that load's its resource to match the theme.
type barIcon struct {
	baseWidget

	onTapped      func()        // The function that will be called when the icon is clicked
	resource      fyne.Resource // The image data of the image that the icon uses
	taskbarWindow Window        // The window associated with this icon if it is a taskbar icon
}

//Tapped means barIcon has been clicked
func (bi *barIcon) Tapped(*fyne.PointEvent) {
	bi.onTapped()
}

//TappedSecondary means barIcon has been clicked by a secondary binding
func (bi *barIcon) TappedSecondary(*fyne.PointEvent) {
}

// Resize sets a new size for a widget.
// Note this should not be used if the widget is being managed by a Layout within a Container.
func (bi *barIcon) Resize(size fyne.Size) {
	bi.resize(size, bi)
}

// Move the widget to a new position, relative to its parent.
// Note this should not be used if the widget is being managed by a Layout within a Container.
func (bi *barIcon) Move(pos fyne.Position) {
	bi.move(pos, bi)
}

// MinSize returns the smallest size this widget can shrink to
func (bi *barIcon) MinSize() fyne.Size {
	return bi.minSize(bi)
}

// Show this widget, if it was previously hidden
func (bi *barIcon) Show() {
	bi.show(bi)
}

// Hide this widget, if it was previously visible
func (bi *barIcon) Hide() {
	bi.hide(bi)
}

// CreateRenderer is a private method to fyne which links this widget to its renderer
func (bi *barIcon) CreateRenderer() fyne.WidgetRenderer {
	render := &barIconRenderer{image: bi}

	render.objects = []fyne.CanvasObject{}

	return render
}

func newBarIcon(res fyne.Resource) *barIcon {
	barIcon := &barIcon{resource: res}

	widget.Refresh(barIcon)
	return barIcon
}
