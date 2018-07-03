package desktop

// #include <stdlib.h>
import "C"

import "github.com/fyne-io/fyne"
import "github.com/fyne-io/fyne/desktop"

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
	return fyne.NewSize(1024, 768)
}

func isEmbedded() bool {
	env := C.getenv(C.CString("DISPLAY"))
	if env != nil {
		return true
	}

	env = C.getenv(C.CString("WAYLAND_DISPLAY"))
	return env != nil
}

func newDeskLayout(bar fyne.CanvasObject) fyne.CanvasObject {
	layout := &deskLayout{bar: bar}

	return fyne.NewContainerWithLayout(layout,
		newBackground(),
		bar,
		newMouse(),
	)
}

// NewDesktop creates a new desktop window (fullscreen for main usage or a
// smaller window when used "embedded" for testing).
// This will automatically detect which mode to run in based on the presence
// of a DISPLAY or WAYLAND_DISPLAY environment variable.
func NewDesktop(app fyne.App) fyne.Window {
	var desk fyne.Window

	if isEmbedded() {
		desk = app.NewWindow("Fyne Desktop")
	} else {
		desk = desktop.CreateWindowWithEngine("drm")
		desk.SetFullscreen(true)
	}

	desk.SetContent(newDeskLayout(newBar(app)))
	return desk
}
