package desktop

import (
	"math"

	"fyne.io/fyne"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
)

// Declare conformity with Layout interface
var _ fyne.Layout = (*BarLayout)(nil)

//BarLayout returns a layout used for zooming linear groups of icons
type BarLayout struct {
	MouseInside   bool
	MousePosition fyne.Position
}

func isHorizontalSpacer(obj fyne.CanvasObject) bool {
	if spacer, ok := obj.(layout.SpacerObject); ok {
		return spacer.ExpandHorizontal()
	}

	return false
}

//SetPointerInside tells the BarLayout that the mouse is inside of the Layout.
func (bl *BarLayout) SetPointerInside(inside bool) {
	bl.MouseInside = inside
}

//SetPointerPosition tells the BarLayout that the mouse position has been updated.
func (bl *BarLayout) SetPointerPosition(position fyne.Position) {
	bl.MousePosition = position
}

func (bl *BarLayout) isSpacer(obj fyne.CanvasObject) bool {
	// invisible spacers don't impact layout
	if !obj.Visible() {
		return false
	}

	return isHorizontalSpacer(obj)
}

// Layout is called to pack all child objects into a specified size.
// For a VBoxLayout this will pack objects into a single column where each item
// is full width but the height is the minimum required.
// Any spacers added will pad the view, sharing the space if there are two or more.
func (bl *BarLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	spacers := make([]fyne.CanvasObject, 0)
	total := 0
	largestY := 0

	for _, child := range objects {
		if !child.Visible() {
			continue
		}

		if bl.isSpacer(child) {
			spacers = append(spacers, child)
			continue
		}

		if bl.MouseInside {
			mouseX := bl.MousePosition.X
			iconX := child.Position().X
			scale := 1.75 - (math.Abs(float64(mouseX-(iconX+fyconSize/2))) / float64((fyconSize * 4)))
			newSize := int(math.Floor(float64(fyconSize) * scale))
			if newSize < fyconSize {
				newSize = fyconSize
			}
			child.Resize(fyne.NewSize(newSize, newSize))
			if largestY < newSize {
				largestY = newSize
			}
			total += newSize
		} else {
			child.Resize(fyne.NewSize(fyconSize, fyconSize))
			total += fyconSize
			largestY = fyconSize
		}
	}

	x := 0
	var extra int
	extra = size.Width - total - (theme.Padding() * (len(objects) - len(spacers) - 1))

	extraCell := 0
	if len(spacers) > 0 {
		extraCell = int(float64(extra) / float64(len(spacers)))
	}

	for _, child := range objects {
		if !child.Visible() {
			continue
		}

		width := child.Size().Width
		height := child.Size().Height

		if bl.isSpacer(child) {
			x += extraCell
			continue
		}
		child.Move(fyne.NewPos(x, largestY-height))
		x += theme.Padding() + width
	}
}

// MinSize finds the smallest size that satisfies all the child objects.
// For a BarLayout this is the width of the widest item and the height is
// the sum of of all children combined with padding between each.
func (bl *BarLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	spacerCount := 0
	minSize := fyne.NewSize(0, 0)
	for _, child := range objects {
		if !child.Visible() {
			continue
		}

		if bl.isSpacer(child) {
			spacerCount++
			continue
		}
		minSize = minSize.Add(fyne.NewSize(child.Size().Width, 0))
		minSize.Height = fyne.Max(child.Size().Height, minSize.Height)
	}
	return minSize.Add(fyne.NewSize(theme.Padding()*(len(objects)-1-spacerCount), 0))
}

// NewBarLayout returns a horizontal fybar
func NewBarLayout() BarLayout {
	return BarLayout{false, fyne.NewPos(0, 0)}
}
