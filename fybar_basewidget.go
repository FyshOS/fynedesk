package desktop

import (
	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/widget"
)

// A base widget class to define the standard widget behaviours.
type fybarBaseWidget struct {
	size     fyne.Size
	position fyne.Position
	Hidden   bool
	disabled bool
}

// Get the current size of this widget.
func (w *fybarBaseWidget) Size() fyne.Size {
	return w.size
}

// Set a new size for a widget.
// Note this should not be used if the widget is being managed by a Layout within a Container.
func (w *fybarBaseWidget) resize(size fyne.Size, parent fyne.Widget) {
	w.size = size

	widget.Renderer(parent).Layout(size)
}

// Get the current position of this widget, relative to its parent.
func (w *fybarBaseWidget) Position() fyne.Position {
	return w.position
}

// Move the widget to a new position, relative to its parent.
// Note this should not be used if the widget is being managed by a Layout within a Container.
func (w *fybarBaseWidget) move(pos fyne.Position, parent fyne.Widget) {
	w.position = pos

	canvas.Refresh(parent)
}

func (w *fybarBaseWidget) minSize(parent fyne.Widget) fyne.Size {
	if widget.Renderer(parent) == nil {
		return fyne.NewSize(0, 0)
	}
	return widget.Renderer(parent).MinSize()
}

func (w *fybarBaseWidget) Visible() bool {
	return !w.Hidden
}

func (w *fybarBaseWidget) show(parent fyne.Widget) {
	if !w.Hidden {
		return
	}

	w.Hidden = false
	canvas.Refresh(parent)
}

func (w *fybarBaseWidget) hide(parent fyne.Widget) {
	if w.Hidden {
		return
	}

	w.Hidden = true
	canvas.Refresh(parent)
}

func (w *fybarBaseWidget) enable(parent fyne.Widget) {
	if !w.disabled {
		return
	}

	w.disabled = false
	canvas.Refresh(parent)
}

func (w *fybarBaseWidget) disable(parent fyne.Widget) {
	if w.disabled {
		return
	}

	w.disabled = true
	canvas.Refresh(parent)
}
