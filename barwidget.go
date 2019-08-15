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

//bar is the main widget housing the icon launcher
type bar struct {
	baseWidget

	children      []fyne.CanvasObject // Icons that are laid out by the barr
	mouseInside   bool                // Is the mouse inside of the bar?
	mousePosition fyne.Position       // The current coordinates of the mouse cursor
}

//MouseIn alerts the widget that the mouse has entered
func (b *bar) MouseIn(*desktop.MouseEvent) {
	b.mouseInside = true
}

//MouseOut alerts the widget that the mouse has left
func (b *bar) MouseOut() {
	b.mouseInside = false
	widget.Renderer(b).Layout(b.Size())
}

//MouseMoved alerts the widget that the mouse has changed position
func (b *bar) MouseMoved(event *desktop.MouseEvent) {
	b.mousePosition = event.Position
	widget.Renderer(b).Layout(b.Size())
}

//Resize resizes the widget to the provided size
func (b *bar) Resize(size fyne.Size) {
	b.resize(size, b)
}

//Move moves the widget to the provide position
func (b *bar) Move(pos fyne.Position) {
	b.move(pos, b)
}

//MinSize returns the minimum size of the widget
func (b *bar) MinSize() fyne.Size {
	return b.minSize(b)
}

//Show makes the widget visible
func (b *bar) Show() {
	b.show(b)
}

//Hide makes the widget hidden
func (b *bar) Hide() {
	b.hide(b)
}

//append adds an object to the end of the widget
func (b *bar) append(object fyne.CanvasObject) {
	if b.Hidden && object.Visible() {
		object.Hide()
	}
	b.children = append(b.children, object)

	widget.Refresh(b)
}

//appendSeparator adds a separator between the default icons and the taskbar
func (b *bar) appendSeparator() {
	object := canvas.NewRectangle(theme.TextColor())
	if b.Hidden && object.Visible() {
		object.Hide()
	}
	b.children = append(b.children, object)

	widget.Refresh(b)
}

//appendToTaskbar adds an object to the taskbar area of the widget just before the final spacer
func (b *bar) appendToTaskbar(object fyne.CanvasObject) {
	if b.Hidden && object.Visible() {
		object.Hide()
	}
	b.children[len(b.children)-1] = object
	b.append(layout.NewSpacer())

	widget.Refresh(b)
}

//removeFromTaskbar removes an object from the taskbar area of the widget
func (b *bar) removeFromTaskbar(object fyne.CanvasObject) {
	if b.Hidden && object.Visible() {
		object.Hide()
	}
	for i, fycon := range b.children {
		if fycon == object {
			b.children = append(b.children[:i], b.children[i+1:]...)
		}
	}

	widget.Refresh(b)
}

//CreateRenderer creates the renderer that will be responsible for painting the widget
func (b *bar) CreateRenderer() fyne.WidgetRenderer {
	return &barRenderer{objects: b.children, layout: newBarLayout(), appBar: b}
}

//newAppBar returns a horizontal list of icons for an icon launcher
func newAppBar(children ...fyne.CanvasObject) *bar {
	bar := &bar{children: children}
	widget.Renderer(bar).Layout(bar.MinSize())
	return bar
}

//barRenderer privdes the renderer functions for the bar Widget
type barRenderer struct {
	layout barLayout

	appBar  *bar
	objects []fyne.CanvasObject
}

//MinSize returns the layout's Min Size
func (b *barRenderer) MinSize() fyne.Size {
	return b.layout.MinSize(b.objects)
}

//Layout recalculates the widget
func (b *barRenderer) Layout(size fyne.Size) {
	b.layout.setPointerInside(b.appBar.mouseInside)
	b.layout.setPointerPosition(b.appBar.mousePosition)
	b.layout.Layout(b.objects, size)
}

//ApplyTheme sets the theme object on the widget
func (b *barRenderer) ApplyTheme() {
}

//BackgroundColor returns the background color of the widget
func (b *barRenderer) BackgroundColor() color.Color {
	return color.Transparent
}

//Objects returns the objects associated with the widget
func (b *barRenderer) Objects() []fyne.CanvasObject {
	return b.objects
}

//Refresh will recalculate the widget and repaint it
func (b *barRenderer) Refresh() {
	b.objects = b.appBar.children
	b.Layout(b.appBar.Size())

	canvas.Refresh(b.appBar)
}

//Destroy destroys the renderer
func (b *barRenderer) Destroy() {

}
