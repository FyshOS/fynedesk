package desktop

import (
	"math"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/theme"
)

// Declare conformity with Layout interface
var _ fyne.Layout = (*barLayout)(nil)

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
	largestY := 0

	for _, child := range objects {
		if !child.Visible() {
			continue
		}

		if bl.mouseInside {
			mouseX := bl.mousePosition.X
			iconX := child.Position().X
			scale := 1.75 - (math.Abs(float64(mouseX-(iconX+iconSize/2))) / float64((iconSize * 4)))
			newSize := int(math.Floor(float64(iconSize) * scale))
			if newSize < iconSize {
				newSize = iconSize
			}
			if _, ok := child.(*canvas.Rectangle); ok {
				child.Resize(fyne.NewSize(2, newSize))
			} else {
				child.Resize(fyne.NewSize(newSize, newSize))
			}
			if largestY < newSize {
				largestY = newSize
			}
			total += newSize
		} else {
			if _, ok := child.(*canvas.Rectangle); ok {
				child.Resize(fyne.NewSize(2, iconSize))
			} else {
				child.Resize(fyne.NewSize(iconSize, iconSize))
			}
			total += iconSize
			largestY = iconSize
		}
	}

	x := 0
	var extra, offset int
	extra = size.Width - total - (theme.Padding() * (len(objects) - 1))
	offset = int(math.Floor(float64(extra / 2.0)))
	x += offset

	for _, child := range objects {
		if !child.Visible() {
			continue
		}

		width := child.Size().Width
		height := child.Size().Height

		child.Move(fyne.NewPos(x, largestY-height))
		x += theme.Padding() + width
	}
}

// MinSize finds the smallest size that satisfies all the child objects.
// For a barLayout this is the width of the widest item and the height is
// the sum of of all children combined with padding between each.
func (bl *barLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	spacerCount := 0
	minSize := fyne.NewSize(0, 0)
	for _, child := range objects {
		if !child.Visible() {
			continue
		}
		minSize = minSize.Add(fyne.NewSize(child.Size().Width, 0))
		minSize.Height = fyne.Max(child.Size().Height, minSize.Height)
	}
	return minSize.Add(fyne.NewSize(theme.Padding()*(len(objects)-1-spacerCount), 0))
}

// NewbarLayout returns a horizontal icon bar
func newBarLayout() barLayout {
	return barLayout{false, fyne.NewPos(0, 0)}
}
