package desktop

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
	iconCount := len(objects) - 1
	barWidth := (iconCount*(iconSize+theme.Padding()) + separatorWidth)
	barLeft := (size.Width - barWidth) / 2
	iconLeft := barLeft

	mouseX := bl.mousePosition.X
	zoom := bl.mouseInside && mouseX >= barLeft && mouseX <= barLeft+barWidth

	for _, child := range objects {
		if zoom {
			iconCenter := iconLeft + iconSize/2
			offsetX := float64(mouseX - iconCenter)

			scale := iconScale - (math.Abs(offsetX) / float64(iconSize*4))
			newSize := int(math.Floor(float64(iconSize) * scale))
			if newSize < iconSize {
				newSize = iconSize
			}
			if _, ok := child.(*canvas.Rectangle); ok {
				child.Resize(fyne.NewSize(separatorWidth, newSize))
				total += separatorWidth
			} else {
				child.Resize(fyne.NewSize(newSize, newSize))
				total += newSize

				if iconLeft+iconSize+theme.Padding() < mouseX {
					offset += (newSize - iconSize)
				} else if iconLeft < mouseX {
					ratio := float64(mouseX-iconLeft) / float64(iconSize+theme.Padding())
					offset += int(float64(newSize-iconSize)*(ratio)) + theme.Padding()
				}
			}
		} else {
			if _, ok := child.(*canvas.Rectangle); ok {
				child.Resize(fyne.NewSize(separatorWidth, iconSize))
				total += separatorWidth
			} else {
				child.Resize(fyne.NewSize(iconSize, iconSize))
				total += iconSize
			}
		}
		total += theme.Padding()
		if _, ok := child.(*canvas.Rectangle); ok {
			iconLeft += separatorWidth
		} else {
			iconLeft += iconSize
		}
		iconLeft += theme.Padding()
	}

	x := 0
	x += barLeft - offset

	for _, child := range objects {
		width := child.Size().Width
		height := child.Size().Height

		if bl.mouseInside {
			child.Move(fyne.NewPos(x, int(float64(iconSize)*iconScale)-height))
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
	if bl.mouseInside {
		return fyne.NewSize((len(objects)-1)*iconSize, int(float64(iconSize)*iconScale))
	}

	return fyne.NewSize((len(objects)-1)*iconSize, iconSize)
}

// NewbarLayout returns a horizontal icon bar
func newBarLayout() barLayout {
	return barLayout{false, fyne.NewPos(0, 0)}
}
