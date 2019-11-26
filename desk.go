package desktop // import "fyne.io/desktop"

import (
	"fmt"
	"github.com/BurntSushi/xgbutil/xinerama"
	"log"
	"math"
	"runtime/debug"

	"fyne.io/fyne"
)

// Desktop defines an embedded or full desktop envionment that we can run.
type Desktop interface {
	Root() fyne.Window
	Run()
	RunApp(AppData) error
	Settings() DeskSettings
	ContentSizePixels(headIndex int) (uint32, uint32)

	IconProvider() ApplicationProvider
	WindowManager() WindowManager
}

var instance Desktop

type deskLayout struct {
	app      fyne.App
	win      fyne.Window
	wm       WindowManager
	icons    ApplicationProvider
	settings DeskSettings

	heads xinerama.Heads

	background          []fyne.CanvasObject
	bar, widgets, mouse fyne.CanvasObject
}

func (l *deskLayout) Layout(objs []fyne.CanvasObject, size fyne.Size) {
	var x, y, w, h int = 0, 0, 0, 0
	if l.heads != nil && len(l.heads) > 1 && len(l.background) > 1 {
		x, y, w, h = l.heads[0].Pieces()
		size.Width = w
		size.Height = h
		for i := 1; i < len(l.heads); i++ {
			xx, yy, ww, hh := l.heads[i].Pieces()
			l.background[i].Move(fyne.NewPos(xx, yy))
			l.background[i].Resize(fyne.NewSize(ww, hh))
		}
	}
	barHeight := l.bar.MinSize().Height
	l.bar.Resize(fyne.NewSize(size.Width, barHeight))
	l.bar.Move(fyne.NewPos(x, size.Height-barHeight))

	widgetsWidth := l.widgets.MinSize().Width
	l.widgets.Resize(fyne.NewSize(widgetsWidth, size.Height))
	l.widgets.Move(fyne.NewPos(size.Width-widgetsWidth, y))

	l.background[0].Resize(size)
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
		l.background = append(l.background, newBackground())
		l.bar = newBar(l)
		l.widgets = newWidgetPanel(l)
		l.mouse = newMouse()
		container := fyne.NewContainerWithLayout(l,
			l.background[0],
			l.bar,
			l.widgets,
			l.mouse,
		)
		if l.heads != nil && len(l.heads) > 1 {
			for i := 1; i < len(l.heads); i++ {
				l.background = append(l.background, newBackground())
				container.AddObject(l.background[i])
			}
		}
		l.win.SetContent(container)
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
			if l.wm != nil {
				l.wm.Close() // attempt to close cleanly to leave X server running
			}
		}
	}()

	l.Root().ShowAndRun()
}

func (l *deskLayout) RunApp(app AppData) error {
	vars := l.scaleVars(l.Root().Canvas().Scale())
	return app.Run(vars)
}

func (l *deskLayout) Settings() DeskSettings {
	return l.settings
}

func (l *deskLayout) ContentSizePixels(headIndex int) (uint32, uint32) {
	if l.background[headIndex] != nil {
		screenW := uint32(float32(l.background[headIndex].Size().Width) * l.Root().Canvas().Scale())
		screenH := uint32(float32(l.background[headIndex].Size().Height) * l.Root().Canvas().Scale())
		if headIndex == 0 {
			return screenW - uint32(float32(l.widgets.Size().Width)*l.Root().Canvas().Scale()), screenH
		}
		return screenW, screenH
	}
	return 0, 0
}

func (l *deskLayout) IconProvider() ApplicationProvider {
	return l.icons
}

func (l *deskLayout) WindowManager() WindowManager {
	return l.wm
}

func (l *deskLayout) scaleVars(scale float32) []string {
	intScale := int(math.Round(float64(scale)))

	return []string{
		fmt.Sprintf("QT_SCALE_FACTOR=%1.1f", scale),
		fmt.Sprintf("GDK_SCALE=%d", intScale),
		fmt.Sprintf("ELM_SCALE=%1.1f", scale),
	}
}

// Instance returns the current desktop environment and provides access to injected functionality.
func Instance() Desktop {
	return instance
}

// NewDesktop creates a new desktop in fullscreen for main usage.
// The WindowManager passed in will be used to manage the screen it is loaded on.
// An ApplicationProvider is used to lookup application icons from the operating system.
func NewDesktop(app fyne.App, wm WindowManager, icons ApplicationProvider, heads xinerama.Heads) Desktop {
	instance = &deskLayout{app: app, wm: wm, icons: icons, settings: NewDeskSettings(), heads: heads}

	return instance
}

// NewEmbeddedDesktop creates a new windowed desktop for test purposes.
// An ApplicationProvider is used to lookup application icons from the operating system.
// If run during CI for testing it will return an in-memory window using the
// fyne/test package.
func NewEmbeddedDesktop(app fyne.App, icons ApplicationProvider) Desktop {
	instance = &deskLayout{app: app, icons: icons, settings: NewDeskSettings()}
	return instance
}
