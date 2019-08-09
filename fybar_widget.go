package desktop

import (
	"image/color"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/driver/desktop"
)

//Fybar is the main widget housing the icon launcher
type Fybar struct {
	fybarBaseWidget

	Children      []fyne.CanvasObject
	Horizontal    bool
	MouseInside   bool
	MousePosition fyne.Position
}

//MouseIn alerts the widget that the mouse has entered
func (fb *Fybar) MouseIn(*desktop.MouseEvent) {
	fb.MouseInside = true
}

//MouseOut alerts the widget that the mouse has left
func (fb *Fybar) MouseOut() {
	fb.MouseInside = false
	Renderer(fb).Layout(fb.Size())
}

//MouseMoved alerts the widget that the mouse has changed position
func (fb *Fybar) MouseMoved(event *desktop.MouseEvent) {
	fb.MousePosition = event.Position
	Renderer(fb).Layout(fb.Size())
}

//Resize resizes the widget to the provided size
func (fb *Fybar) Resize(size fyne.Size) {
	fb.resize(size, fb)
}

//Move moves the widget to the provide position
func (fb *Fybar) Move(pos fyne.Position) {
	fb.move(pos, fb)
}

//MinSize returns the minimum size of the widget
func (fb *Fybar) MinSize() fyne.Size {
	return fb.minSize(fb)
}

//Show makes the widget visible
func (fb *Fybar) Show() {
	fb.show(fb)
}

//Hide makes the widget hidden
func (fb *Fybar) Hide() {
	fb.hide(fb)
}

//Prepend adds an object to the begging of the widget
func (fb *Fybar) Prepend(object fyne.CanvasObject) {
	if fb.Hidden && object.Visible() {
		object.Hide()
	}
	fb.Children = append([]fyne.CanvasObject{object}, fb.Children...)

	Refresh(fb)
}

//Append adds an object to the end of the widget
func (fb *Fybar) Append(object fyne.CanvasObject) {
	if fb.Hidden && object.Visible() {
		object.Hide()
	}
	fb.Children = append(fb.Children, object)

	Refresh(fb)
}

//CreateRenderer creates the renderer that will be responsible for painting the widget
func (fb *Fybar) CreateRenderer() fyne.WidgetRenderer {
	var lay FybarLayout
	if fb.Horizontal {
		lay = NewHFybarLayout()
	} else {
		lay = NewVFybarLayout()
	}

	return &fybarRenderer{objects: fb.Children, layout: lay, fybar: fb}
}

//NewHFybar returns a horizontal list of icons for an icon launcher
func NewHFybar(children ...fyne.CanvasObject) *Fybar {
	fybar := &Fybar{Horizontal: true, Children: children}
	Renderer(fybar).Layout(fybar.MinSize())
	return fybar
}

//NewVFybar returns a vertical list of icons for an icon launcher
func NewVFybar(children ...fyne.CanvasObject) *Fybar {
	fybar := &Fybar{Horizontal: false, Children: children}
	Renderer(fybar).Layout(fybar.MinSize())
	return fybar
}

type fybarRenderer struct {
	layout FybarLayout

	fybar   *Fybar
	objects []fyne.CanvasObject
}

func (fb *fybarRenderer) MinSize() fyne.Size {
	return fb.layout.MinSize(fb.objects)
}

func (fb *fybarRenderer) Layout(size fyne.Size) {
	fb.layout.SetPointerInside(fb.fybar.MouseInside)
	fb.layout.SetPointerPosition(fb.fybar.MousePosition)
	fb.layout.Layout(fb.objects, size)
}

func (fb *fybarRenderer) ApplyTheme() {
}

func (fb *fybarRenderer) BackgroundColor() color.Color {
	return color.Transparent
}

func (fb *fybarRenderer) Objects() []fyne.CanvasObject {
	return fb.objects
}

func (fb *fybarRenderer) Refresh() {
	fb.objects = fb.fybar.Children
	fb.Layout(fb.fybar.Size())

	canvas.Refresh(fb.fybar)
}

func (fb *fybarRenderer) Destroy() {

}
