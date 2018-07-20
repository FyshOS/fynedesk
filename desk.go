package desktop

import "github.com/fyne-io/fyne"

type deskLayout struct {
	bar fyne.CanvasObject
}

func (l *deskLayout) Layout(objs []fyne.CanvasObject, size fyne.Size) {
	for _, child := range objs {
		if child == l.bar {
			barHeight := l.bar.MinSize().Height
			child.Resize(fyne.NewSize(size.Width, barHeight))
			child.Move(fyne.NewPos(0, size.Height-barHeight))
			return
		}

		child.Resize(size)
	}
}

func (l *deskLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(1280, 720)
}

func newDeskLayout(bar fyne.CanvasObject) fyne.CanvasObject {
	layout := &deskLayout{bar: bar}

	return fyne.NewContainerWithLayout(layout,
		newBackground(),
		bar,
		newMouse(),
	)
}

// NewDesktop creates a new desktop in fullscreen for main usage
// or a window in the current desktop for test purposes.
func NewDesktop(app fyne.App) fyne.Window {
	desk := NewDesktopWindow(app)
	initInput()

	desk.SetContent(newDeskLayout(newBar(app)))
	return desk
}
