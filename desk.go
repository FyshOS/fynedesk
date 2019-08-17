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

	IconProvider() IconProvider
	WindowManager() WindowManager
}

type deskLayout struct {
	app   fyne.App
	win   fyne.Window
	wm    WindowManager
	icons IconProvider

	background, bar, widgets, mouse fyne.CanvasObject
}

func (l *deskLayout) Layout(objs []fyne.CanvasObject, size fyne.Size) {
	barHeight := l.bar.MinSize().Height
	l.bar.Resize(fyne.NewSize(size.Width, barHeight))
	l.bar.Move(fyne.NewPos(0, size.Height-barHeight))

	widgetsWidth := l.widgets.MinSize().Width
	l.widgets.Resize(fyne.NewSize(widgetsWidth, size.Height))
	l.widgets.Move(fyne.NewPos(size.Width-widgetsWidth, 0))

	l.background.Resize(size)
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

		l.background = newBackground()
		l.bar = newBar(l)
		l.widgets = newWidgetPanel(l.win)
		l.mouse = newMouse()
		l.win.SetContent(fyne.NewContainerWithLayout(l,
			l.background,
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

func (l *deskLayout) IconProvider() IconProvider {
	return l.icons
}

func (l *deskLayout) WindowManager() WindowManager {
	return l.wm
}

// NewDesktop creates a new desktop in fullscreen for main usage.
// The WindowManager passed in will be used to manage the screen it is loaded on.
// An IconProvider is used to lookup application icons from the operating system.
func NewDesktop(app fyne.App, wm WindowManager, icons IconProvider) Desktop {
	return &deskLayout{app: app, wm: wm, icons: icons}
}

// NewEmbeddedDesktop creates a new windowed desktop for test purposes.
// An IconProvider is used to lookup application icons from the operating system.
// If run during CI for testing it will return an in-memory window using the
// fyne/test package.
func NewEmbeddedDesktop(app fyne.App, icons IconProvider) Desktop {
	return &deskLayout{app: app, icons: icons}
}
