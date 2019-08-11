package desktop

import (
	"image/color"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/driver/desktop"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
)

//Bar is the main widget housing the icon launcher
type Bar struct {
	baseWidget

	Children      []fyne.CanvasObject
	MouseInside   bool
	MousePosition fyne.Position
}

//MouseIn alerts the widget that the mouse has entered
func (b *Bar) MouseIn(*desktop.MouseEvent) {
	b.MouseInside = true
}

//MouseOut alerts the widget that the mouse has left
func (b *Bar) MouseOut() {
	b.MouseInside = false
	widget.Renderer(b).Layout(b.Size())
}

//MouseMoved alerts the widget that the mouse has changed position
func (b *Bar) MouseMoved(event *desktop.MouseEvent) {
	b.MousePosition = event.Position
	widget.Renderer(b).Layout(b.Size())
}

//Resize resizes the widget to the provided size
func (b *Bar) Resize(size fyne.Size) {
	b.resize(size, b)
}

//Move moves the widget to the provide position
func (b *Bar) Move(pos fyne.Position) {
	b.move(pos, b)
}

//MinSize returns the minimum size of the widget
func (b *Bar) MinSize() fyne.Size {
	return b.minSize(b)
}

//Show makes the widget visible
func (b *Bar) Show() {
	b.show(b)
}

//Hide makes the widget hidden
func (b *Bar) Hide() {
	b.hide(b)
}

//Prepend adds an object to the begging of the widget
func (b *Bar) Prepend(object fyne.CanvasObject) {
	if b.Hidden && object.Visible() {
		object.Hide()
	}
	b.Children = append([]fyne.CanvasObject{object}, b.Children...)

	widget.Refresh(b)
}

//Append adds an object to the end of the widget
func (b *Bar) Append(object fyne.CanvasObject) {
	if b.Hidden && object.Visible() {
		object.Hide()
	}
	b.Children = append(b.Children, object)

	widget.Refresh(b)
}

//AppendSeparator adds a separator between the default icons and the taskbar
func (b *Bar) AppendSeparator() {
	object := canvas.NewRectangle(theme.BackgroundColor())
	if b.Hidden && object.Visible() {
		object.Hide()
	}
	b.Children = append(b.Children, object)

	widget.Refresh(b)
}

//AppendTaskbar adds an object to the taskbar area of the widget just before the final spacer
func (b *Bar) AppendTaskbar(object fyne.CanvasObject) {
	if b.Hidden && object.Visible() {
		object.Hide()
	}
	b.Children[len(b.Children)-1] = object
	b.Append(layout.NewSpacer())

	widget.Refresh(b)
}

//CreateRenderer creates the renderer that will be responsible for painting the widget
func (b *Bar) CreateRenderer() fyne.WidgetRenderer {
	return &BarRenderer{objects: b.Children, layout: NewBarLayout(), Bar: b}
}

//NewAppBar returns a horizontal list of icons for an icon launcher
func NewAppBar(children ...fyne.CanvasObject) *Bar {
	Bar := &Bar{Children: children}
	widget.Renderer(Bar).Layout(Bar.MinSize())
	return Bar
}

//BarRenderer privdes the renderer functions for the Bar Widget
type BarRenderer struct {
	layout BarLayout

	Bar     *Bar
	objects []fyne.CanvasObject
}

//MinSize returns the layout's Min Size
func (b *BarRenderer) MinSize() fyne.Size {
	return b.layout.MinSize(b.objects)
}

//Layout recalculates the widget
func (b *BarRenderer) Layout(size fyne.Size) {
	b.layout.SetPointerInside(b.Bar.MouseInside)
	b.layout.SetPointerPosition(b.Bar.MousePosition)
	b.layout.Layout(b.objects, size)
}

//ApplyTheme sets the theme object on the widget
func (b *BarRenderer) ApplyTheme() {
}

//BackgroundColor returns the background color of the widget
func (b *BarRenderer) BackgroundColor() color.Color {
	return color.Transparent
}

//Objects returns the objects associated with the widget
func (b *BarRenderer) Objects() []fyne.CanvasObject {
	return b.objects
}

//Refresh will recalculate the widget and repaint it
func (b *BarRenderer) Refresh() {
	b.objects = b.Bar.Children
	b.Layout(b.Bar.Size())

	canvas.Refresh(b.Bar)
}

//Destroy destroys the renderer
func (b *BarRenderer) Destroy() {

}
