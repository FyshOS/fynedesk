package driver

import (
	"fyne.io/fyne"
	"fyne.io/fyne/widget"
)

func walkObjectTree(
	obj fyne.CanvasObject,
	beforeChildren func(fyne.CanvasObject, fyne.Position),
) {
	doWalkObjectTree(obj, fyne.NewPos(0, 0), beforeChildren)
}

func doWalkObjectTree(
	obj fyne.CanvasObject,
	offset fyne.Position,
	callback func(fyne.CanvasObject, fyne.Position),
) {
	pos := obj.Position().Add(offset)

	var children []fyne.CanvasObject
	switch co := obj.(type) {
	case *fyne.Container:
		children = co.Objects
	case fyne.Widget:
		children = widget.Renderer(co).Objects()
	}

	callback(obj, pos)

	for _, child := range children {
		doWalkObjectTree(child, pos, callback)
	}
}
