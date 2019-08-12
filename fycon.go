package desktop

import (
	"image/color"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
)

type fyconRenderer struct {
	objects []fyne.CanvasObject

	image *Fycon
}

func (fy *fyconRenderer) MinSize() fyne.Size {
	size := theme.IconInlineSize()
	return fyne.NewSize(size, size)
}

func (fy *fyconRenderer) Layout(size fyne.Size) {
	if len(fy.objects) == 0 {
		return
	}

	fy.objects[0].Resize(size)
}

func (fy *fyconRenderer) Objects() []fyne.CanvasObject {
	return fy.objects
}

func (fy *fyconRenderer) ApplyTheme() {
	fy.Refresh()
}

func (fy *fyconRenderer) BackgroundColor() color.Color {
	return color.Transparent
}

func (fy *fyconRenderer) Refresh() {
	fy.objects = nil

	if fy.image.Resource != nil {
		raster := canvas.NewImageFromResource(fy.image.Resource)
		raster.FillMode = canvas.ImageFillContain

		fy.objects = append(fy.objects, raster)
	}
	fy.Layout(fy.image.Size())

	canvas.Refresh(fy.image)
}

func (fy *fyconRenderer) Destroy() {
}

// Fycon widget is a basic image component that load's its resource to match the theme.
type Fycon struct {
	baseWidget

	OnTapped      func()
	Resource      fyne.Resource // The resource for this Fycon
	TaskbarWindow Window
}

//Tapped means Fycon has been clicked
func (fy *Fycon) Tapped(*fyne.PointEvent) {
	fy.OnTapped()
}

//TappedSecondary means Fycon has been clicked by a secondary binding
func (fy *Fycon) TappedSecondary(*fyne.PointEvent) {
}

// Resize sets a new size for a widget.
// Note this should not be used if the widget is being managed by a Layout within a Container.
func (fy *Fycon) Resize(size fyne.Size) {
	fy.resize(size, fy)
}

// Move the widget to a new position, relative to its parent.
// Note this should not be used if the widget is being managed by a Layout within a Container.
func (fy *Fycon) Move(pos fyne.Position) {
	fy.move(pos, fy)
}

// MinSize returns the smallest size this widget can shrink to
func (fy *Fycon) MinSize() fyne.Size {
	return fy.minSize(fy)
}

// Show this widget, if it was previously hidden
func (fy *Fycon) Show() {
	fy.show(fy)
}

// Hide this widget, if it was previously visible
func (fy *Fycon) Hide() {
	fy.hide(fy)
}

// SetResource updates the resource rendered in this Fycon widget
func (fy *Fycon) SetResource(res fyne.Resource) {
	fy.Resource = res

	widget.Refresh(fy)
}

// CreateRenderer is a private method to Fyne which links this widget to its renderer
func (fy *Fycon) CreateRenderer() fyne.WidgetRenderer {
	render := &fyconRenderer{image: fy}

	render.objects = []fyne.CanvasObject{}

	return render
}

// NewFycon returns a new Fycon widget that displays a themed Fycon resource
func NewFycon(res fyne.Resource) *Fycon {
	fycon := &Fycon{Resource: res}

	widget.Refresh(fycon)
	return fycon
}
