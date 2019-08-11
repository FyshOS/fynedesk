package desktop

import (
	"log"
	"runtime/debug"

	"fyne.io/fyne"
)

// Desktop defines an embedded or full desktop envionment that we can run.
type Desktop interface {
	Root() fyne.Window
	Run()
}

type deskLayout struct {
	app fyne.App
	win fyne.Window
	wm  WindowManager

	bar, widgets, mouse fyne.CanvasObject
}

func (l *deskLayout) Layout(objs []fyne.CanvasObject, size fyne.Size) {
	for _, child := range objs {
		switch child {
		case l.bar:
			barHeight := l.bar.MinSize().Height
			child.Resize(fyne.NewSize(size.Width, barHeight))
			child.Move(fyne.NewPos(0, size.Height-barHeight))
		case l.widgets:
			widgetsWidth := l.widgets.MinSize().Width
			child.Resize(fyne.NewSize(widgetsWidth, size.Height))
			child.Move(fyne.NewPos(size.Width-widgetsWidth, 0))
		default:
			if child != l.mouse {
				child.Resize(size)
			}
		}
	}
}

func (l *deskLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(1280, 720)
}

func (l *deskLayout) newDesktopWindow() fyne.Window {
	desk := l.app.NewWindow("Fyne Desktop")
	desk.SetPadded(false)

	if l.wm != nil {
		desk.FullScreen()
	}

	return desk
}

func (l *deskLayout) Root() fyne.Window {
	if l.win == nil {
		l.win = l.newDesktopWindow()

		l.bar = newBar(l.wm)
		l.widgets = newWidgetPanel(l.win)
		l.mouse = newMouse()
		l.win.SetContent(fyne.NewContainerWithLayout(l,
			newBackground(),
			l.bar,
			l.widgets,
			mouse,
		))

		l.mouse.Hide() // temporarily we do not handle mouse (using X default)
		if l.wm != nil {
			l.win.SetOnClosed(func() {
				l.wm.Close()
			})
		}
	}

	return l.win
}

func (l *deskLayout) Run() {
	debug.SetPanicOnFault(true)

	defer func() {
		if r := recover(); r != nil {
			log.Println("Crashed!!!")
			l.wm.Close() // attempt to close cleanly to leave X server running
		}
	}()

	l.Root().ShowAndRun()
}

// NewDesktop creates a new desktop in fullscreen for main usage.
// The WindowManager passed in will be used to manage the screen it is loaded on.
func NewDesktop(app fyne.App, wm WindowManager) Desktop {
	return &deskLayout{app: app, wm: wm}
}

// NewEmbeddedDesktop creates a new windowed desktop for test purposes.
// If run during CI for testing it will return an in-memory window using the
// fyne/test package.
func NewEmbeddedDesktop(app fyne.App) Desktop {
	return &deskLayout{app: app}
}
