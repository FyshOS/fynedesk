package ui

import (
	"math"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/theme"
)

// Declare conformity with Layout interface
var _ fyne.Layout = (*barLayout)(nil)

const separatorWidth = 2

//barLayout returns a layout used for zooming linear groups of icons
type barLayout struct {
	bar *bar

	mouseInside   bool          // Is the mouse inside of the layout?
	mousePosition fyne.Position // Current coordinates of the mouse cursor
}

//setPointerInside tells the barLayout that the mouse is inside of the Layout.
func (bl *barLayout) setPointerInside(inside bool) {
	bl.mouseInside = inside
}

//setPointerPosition tells the barLayout that the mouse position has been updated.
func (bl *barLayout) setPointerPosition(position fyne.Position) {
	bl.mousePosition = position
}

// Layout is called to pack all icons into a specified size.  It also handles the zooming effect of the icons.
func (bl *barLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	total := 0
	offset := 0
	barWidth := 0
	iconCount := len(objects)
	if !bl.bar.disableTaskbar {
		iconCount--
		barWidth = (iconCount * (bl.bar.iconSize + theme.Padding())) + separatorWidth
	} else {
		barWidth = iconCount * (bl.bar.iconSize + theme.Padding())
	}
	barLeft := (size.Width - barWidth) / 2
	iconLeft := barLeft

	mouseX := bl.mousePosition.X
	zoom := bl.mouseInside && mouseX >= barLeft && mouseX <= barLeft+barWidth

	for _, child := range objects {
		if zoom && !bl.bar.disableZoom {
			iconCenter := iconLeft + bl.bar.iconSize/2
			offsetX := float64(mouseX - iconCenter)

			scale := float64(bl.bar.iconScale) - (math.Abs(offsetX) / float64(bl.bar.iconSize*4))
			newSize := int(math.Floor(float64(bl.bar.iconSize) * scale))
			if newSize < bl.bar.iconSize {
				newSize = bl.bar.iconSize
			}
			if _, ok := child.(*canvas.Rectangle); ok {
				child.Resize(fyne.NewSize(separatorWidth, newSize))
				total += separatorWidth
			} else {
				child.Resize(fyne.NewSize(newSize, newSize))
				total += newSize

				if iconLeft+bl.bar.iconSize+theme.Padding() < mouseX {
					offset += (newSize - bl.bar.iconSize)
				} else if iconLeft < mouseX {
					ratio := float64(mouseX-iconLeft) / float64(bl.bar.iconSize+theme.Padding())
					offset += int(float64(newSize-bl.bar.iconSize)*(ratio)) + theme.Padding()
				}
			}
		} else {
			if _, ok := child.(*canvas.Rectangle); ok {
				child.Resize(fyne.NewSize(separatorWidth, bl.bar.iconSize))
				total += separatorWidth
			} else {
				child.Resize(fyne.NewSize(bl.bar.iconSize, bl.bar.iconSize))
				total += bl.bar.iconSize
			}
		}
		total += theme.Padding()
		if _, ok := child.(*canvas.Rectangle); ok {
			iconLeft += separatorWidth
		} else {
			iconLeft += bl.bar.iconSize
		}
		iconLeft += theme.Padding()
	}

	x := 0
	x += barLeft - offset

	for _, child := range objects {
		width := child.Size().Width
		height := child.Size().Height

		if bl.mouseInside {
			child.Move(fyne.NewPos(x, int(float32(bl.bar.iconSize)*bl.bar.iconScale)-height))
		} else {
			child.Move(fyne.NewPos(x, 0))
		}
		x += width + theme.Padding()
	}
}

// MinSize finds the smallest size that satisfies all the child objects.
// For a barLayout this is the width of the widest item and the height is
// the sum of of all children combined with padding between each.
func (bl *barLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	barWidth := 0
	iconCount := len(objects)
	if !bl.bar.disableTaskbar {
		iconCount--
		barWidth = (iconCount * (bl.bar.iconSize + theme.Padding())) + separatorWidth
	} else {
		barWidth = iconCount * (bl.bar.iconSize + theme.Padding())
	}
	if bl.mouseInside {

		return fyne.NewSize(barWidth, int(float32(bl.bar.iconSize)*bl.bar.iconScale))
	}

	return fyne.NewSize(barWidth, bl.bar.iconSize)
}

// NewbarLayout returns a horizontal icon bar
func newBarLayout(bar *bar) barLayout {
	return barLayout{bar, false, fyne.NewPos(0, 0)}
}
