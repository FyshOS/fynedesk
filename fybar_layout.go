package desktop

import (
	"math"

	"fyne.io/fyne"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
)

// Declare conformity with Layout interface
var _ fyne.Layout = (*FybarLayout)(nil)

//FybarLayout returns a layout used for zooming linear groups of icons
type FybarLayout struct {
	horizontal    bool
	MouseInside   bool
	MousePosition fyne.Position
}

func isVerticalSpacer(obj fyne.CanvasObject) bool {
	if spacer, ok := obj.(layout.SpacerObject); ok {
		return spacer.ExpandVertical()
	}

	return false
}

func isHorizontalSpacer(obj fyne.CanvasObject) bool {
	if spacer, ok := obj.(layout.SpacerObject); ok {
		return spacer.ExpandHorizontal()
	}

	return false
}

//SetPointerInside tells the FybarLayout that the mouse is inside of the Layout.
func (fbl *FybarLayout) SetPointerInside(inside bool) {
	fbl.MouseInside = inside
}

//SetPointerPosition tells the FybarLayout that the mouse position has been updated.
func (fbl *FybarLayout) SetPointerPosition(position fyne.Position) {
	fbl.MousePosition = position
}

func (fbl *FybarLayout) isSpacer(obj fyne.CanvasObject) bool {
	// invisible spacers don't impact layout
	if !obj.Visible() {
		return false
	}

	if fbl.horizontal {
		return isHorizontalSpacer(obj)
	}
	return isVerticalSpacer(obj)
}

// Layout is called to pack all child objects into a specified size.
// For a VBoxLayout this will pack objects into a single column where each item
// is full width but the height is the minimum required.
// Any spacers added will pad the view, sharing the space if there are two or more.
func (fbl *FybarLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	spacers := make([]fyne.CanvasObject, 0)
	total := 0
	largestX := 0
	largestY := 0

	for _, child := range objects {
		if !child.Visible() {
			continue
		}

		if fbl.isSpacer(child) {
			spacers = append(spacers, child)
			continue
		}

		if fbl.horizontal {
			if fbl.MouseInside {
				mouseX := fbl.MousePosition.X
				iconX := child.Position().X
				scale := 1.75 - (math.Abs(float64(mouseX-(iconX+int(fyconSize/2.0))) / float64((fyconSize * 4.0))))
				newSize := int(math.Floor(float64(float64(fyconSize) * scale)))
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
		} else {
			if fbl.MouseInside {
				mouseY := fbl.MousePosition.Y
				iconY := child.Position().Y
				scale := 1.75 - (math.Abs(float64(mouseY-(iconY+int(fyconSize/2.0))) / float64((fyconSize * 4.0))))
				newSize := int(math.Floor(float64(float64(fyconSize) * scale)))
				child.Resize(fyne.NewSize(newSize, newSize))
				if largestX < newSize {
					largestX = newSize
				}
				total += newSize
			} else {
				child.Resize(fyne.NewSize(fyconSize, fyconSize))
				total += fyconSize
				largestX = fyconSize
			}
		}
	}

	x, y := 0, 0
	var extra int
	if fbl.horizontal {
		extra = size.Width - total - (theme.Padding() * (len(objects) - len(spacers) - 1))
	} else {
		extra = size.Height - total - (theme.Padding() * (len(objects) - len(spacers) - 1))
	}
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

		if fbl.isSpacer(child) {
			if fbl.horizontal {
				x += extraCell
			} else {
				y += extraCell
			}
			continue
		}
		if fbl.horizontal {
			child.Move(fyne.NewPos(x, largestY-height))
			x += theme.Padding() + width
		} else {
			child.Move(fyne.NewPos(largestX-width, y))
			y += theme.Padding() + height
		}
	}
}

// MinSize finds the smallest size that satisfies all the child objects.
// For a FybarLayout this is the width of the widest item and the height is
// the sum of of all children combined with padding between each.
func (fbl *FybarLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	spacerCount := 0
	minSize := fyne.NewSize(0, 0)
	for _, child := range objects {
		if !child.Visible() {
			continue
		}

		if fbl.isSpacer(child) {
			spacerCount++
			continue
		}

		if fbl.horizontal {
			minSize = minSize.Add(fyne.NewSize(child.Size().Width, 0))
			minSize.Height = fyne.Max(child.Size().Height, minSize.Height)
		} else {
			minSize = minSize.Add(fyne.NewSize(0, child.Size().Height))
			minSize.Width = fyne.Max(child.Size().Width, minSize.Width)
		}
	}
	if fbl.horizontal {
		return minSize.Add(fyne.NewSize(theme.Padding()*(len(objects)-1-spacerCount), 0))
	}
	return minSize.Add(fyne.NewSize(0, theme.Padding()*(len(objects)-1-spacerCount)))
}

// NewHFybarLayout returns a horizontal fybar
func NewHFybarLayout() FybarLayout {
	return FybarLayout{true, false, fyne.NewPos(0, 0)}
}

// NewVFybarLayout returns a vertical fybar
func NewVFybarLayout() FybarLayout {
	return FybarLayout{false, false, fyne.NewPos(0, 0)}
}
