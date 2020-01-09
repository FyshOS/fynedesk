package ui

import (
	"fmt"
	"log"
	"math"
	"runtime/debug"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	deskDriver "fyne.io/fyne/driver/desktop"

	"fyne.io/desktop"
	"fyne.io/desktop/internal/modules/builtin"
)

type deskLayout struct {
	app      fyne.App
	win      fyne.Window
	wm       desktop.WindowManager
	icons    desktop.ApplicationProvider
	screens  desktop.ScreenList
	settings desktop.DeskSettings

	backgrounds         []*background
	bar                 *bar
	widgets, mouse      fyne.CanvasObject
	container           *fyne.Container
	screenBackgroundMap map[*desktop.Screen]fyne.CanvasObject
}

type embeddedScreensProvider struct {
	screens []*desktop.Screen
}

func applyScale(coord int, scale float32) int {
	newCoord := int(math.Round(float64(coord) / float64(scale)))
	return newCoord
}

func removeScale(coord int, scale float32) int {
	newCoord := int(math.Round(float64(coord) * float64(scale)))
	return newCoord
}

func (l *deskLayout) Layout(objs []fyne.CanvasObject, size fyne.Size) {
	screens := l.screens.Screens()
	primary := l.screens.Primary()
	x := applyScale(primary.X, l.screens.Primary().CanvasScale()) // TODO here we need to get the right screen
	y := applyScale(primary.Y, l.screens.Primary().CanvasScale())
	w := applyScale(primary.Width, l.screens.Primary().CanvasScale())
	h := applyScale(primary.Height, l.screens.Primary().CanvasScale())
	size.Width = w
	size.Height = h
	if screens != nil && len(screens) > 1 && len(l.backgrounds) > 1 {
		for i := 0; i < len(screens); i++ {
			if screens[i] == primary {
				continue
			}
			xx := applyScale(screens[i].X, l.screens.Primary().CanvasScale())
			yy := applyScale(screens[i].Y, l.screens.Primary().CanvasScale())
			ww := applyScale(screens[i].Width, l.screens.Primary().CanvasScale())
			hh := applyScale(screens[i].Height, l.screens.Primary().CanvasScale())
			background := l.screenBackgroundMap[screens[i]]
			if background != nil {
				background.Move(fyne.NewPos(xx, yy))
				background.Resize(fyne.NewSize(ww, hh))
			}
		}
	}

	barHeight := l.bar.MinSize().Height
	l.bar.Resize(fyne.NewSize(size.Width, y+barHeight))
	l.bar.Move(fyne.NewPos(x, y+size.Height-barHeight))

	widgetsWidth := l.widgets.MinSize().Width
	l.widgets.Resize(fyne.NewSize(widgetsWidth, size.Height))
	l.widgets.Move(fyne.NewPos(x+size.Width-widgetsWidth, y))

	var background fyne.CanvasObject
	if len(screens) > 1 {
		background = l.screenBackgroundMap[primary]
	} else {
		background = l.backgrounds[0]
	}
	if background != nil {
		background.Move(fyne.NewPos(x, y))
		background.Resize(size)
	}
}

func (l *deskLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if l.wm == nil {
		return fyne.NewSize(1024, 576)
	}
	return fyne.NewSize(640, 480) // tiny - the window manager will scale up to screen size
}

func (l *deskLayout) newDesktopWindow() fyne.Window {
	if l.wm == nil {
		win := l.app.NewWindow("Fyne Desktop (Embedded)")
		win.SetPadded(false)
		return win
	}

	desk := l.app.NewWindow("Fyne Desktop")
	desk.SetPadded(false)
	desk.FullScreen()

	return desk
}

func (l *deskLayout) updateBackgrounds(path string) {
	for _, background := range l.backgrounds {
		background.updateBackgroundPath(path)
		canvas.Refresh(background)
	}
}

func (l *deskLayout) Root() fyne.Window {
	if l.win != nil {
		return l.win
	}

	l.win = l.newDesktopWindow()
	l.backgrounds = append(l.backgrounds, newBackground())
	l.screenBackgroundMap[l.screens.Screens()[0]] = l.backgrounds[0]
	l.bar = newBar(l)
	l.widgets = newWidgetPanel(l)
	l.mouse = newMouse()
	l.container = fyne.NewContainerWithLayout(l, l.backgrounds[0])
	if l.screens.Screens() != nil && len(l.screens.Screens()) > 1 {
		for i := 1; i < len(l.screens.Screens()); i++ {
			l.backgrounds = append(l.backgrounds, newBackground())
			l.screenBackgroundMap[l.screens.Screens()[i]] = l.backgrounds[i]
			l.container.AddObject(l.backgrounds[i])
		}
	}
	l.container.AddObject(l.bar)
	l.container.AddObject(l.widgets)
	l.container.AddObject(mouse)

	l.win.SetContent(l.container)
	l.mouse.Hide() // temporarily we do not draw mouse (using X default)
	if l.wm != nil {
		l.win.SetOnClosed(func() {
			l.wm.Close()
		})
	}

	return l.win
}

func (l *deskLayout) Run() {
	if l.wm == nil {
		l.Root().ShowAndRun()
		return
	}
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

func (l *deskLayout) RunApp(app desktop.AppData) error {
	vars := l.scaleVars(l.Screens().Active().CanvasScale())
	return app.Run(vars)
}

func (l *deskLayout) Settings() desktop.DeskSettings {
	return l.settings
}

func (l *deskLayout) ContentSizePixels(screen *desktop.Screen) (uint32, uint32) {
	screenW := uint32(screen.Width)
	screenH := uint32(screen.Height)
	if l.screens.Primary() == screen {
		return screenW - uint32(float32(l.widgets.Size().Width)*screen.CanvasScale()), screenH
	}
	return screenW, screenH
}

func (l *deskLayout) IconProvider() desktop.ApplicationProvider {
	return l.icons
}

func (l *deskLayout) WindowManager() desktop.WindowManager {
	return l.wm
}

func (l *deskLayout) Modules() []desktop.Module {
	return []desktop.Module{builtin.NewBattery(), builtin.NewBrightness()}
}

func (l *deskLayout) scaleVars(scale float32) []string {
	intScale := int(math.Round(float64(scale)))

	return []string{
		fmt.Sprintf("QT_SCALE_FACTOR=%1.1f", scale),
		fmt.Sprintf("GDK_SCALE=%d", intScale),
		fmt.Sprintf("ELM_SCALE=%1.1f", scale),
	}
}

// MouseInNotify can be called by the window manager to alert the desktop that the cursor has entered the canvas
func (l *deskLayout) MouseInNotify(pos fyne.Position) {
	mouseX, mouseY := pos.X, pos.Y
	barX, barY := l.bar.Position().X, l.bar.Position().Y
	barWidth, barHeight := l.bar.Size().Width, l.bar.Size().Height
	if mouseX >= barX && mouseX <= barX+barWidth {
		if mouseY >= barY && mouseY <= barY+barHeight {
			l.bar.MouseIn(&deskDriver.MouseEvent{PointEvent: fyne.PointEvent{AbsolutePosition: pos, Position: pos}})
		}
	}
}

// MouseOutNotify can be called by the window manager to alert the desktop that the cursor has left the canvas
func (l *deskLayout) MouseOutNotify() {
	l.bar.MouseOut()
}

func (l *deskLayout) startSettingsChangeListener(listener chan desktop.DeskSettings) {
	for {
		_ = <-listener
		l.updateBackgrounds(l.Settings().Background())
		l.bar.iconSize = l.Settings().LauncherIconSize()
		l.bar.iconScale = float32(l.Settings().LauncherZoomScale())
		l.bar.disableZoom = l.Settings().LauncherDisableZoom()
		l.bar.updateIcons()
		l.bar.updateIconOrder()
		l.bar.updateTaskbar()
	}
}

func (l *deskLayout) addSettingsChangeListener() {
	listener := make(chan desktop.DeskSettings)
	l.Settings().AddChangeListener(listener)
	go l.startSettingsChangeListener(listener)
}

// Screens returns the screens provider of the current desktop environment for access to screen functionality.
func (l *deskLayout) Screens() desktop.ScreenList {
	return l.screens
}

func (esp embeddedScreensProvider) Screens() []*desktop.Screen {
	l := desktop.Instance().(*deskLayout)
	if esp.screens == nil {
		scale := desktop.Instance().(*deskLayout).Root().Canvas().Scale()

		esp.screens = []*desktop.Screen{{Name: "Screen0", X: 0, Y: 0,
			Width:  removeScale(int(l.Root().Canvas().Size().Width), scale),
			Height: removeScale(int(l.Root().Canvas().Size().Height), scale)}}
	}
	return esp.screens
}

func (esp embeddedScreensProvider) Active() *desktop.Screen {
	return esp.Screens()[0]
}

func (esp embeddedScreensProvider) Primary() *desktop.Screen {
	return esp.Screens()[0]
}

func (esp embeddedScreensProvider) ScreenForWindow(win desktop.Window) *desktop.Screen {
	return esp.Screens()[0]
}

func (esp embeddedScreensProvider) ScreenForGeometry(x int, y int, width int, height int) *desktop.Screen {
	return esp.Screens()[0]
}

// NewEmbeddedScreensProvider returns a screen provider for use in embedded desktop mode
func NewEmbeddedScreensProvider() desktop.ScreenList {
	return &embeddedScreensProvider{}
}

// NewDesktop creates a new desktop in fullscreen for main usage.
// The WindowManager passed in will be used to manage the screen it is loaded on.
// An ApplicationProvider is used to lookup application icons from the operating system.
func NewDesktop(app fyne.App, wm desktop.WindowManager, icons desktop.ApplicationProvider, screenProvider desktop.ScreenList) desktop.Desktop {
	desk := &deskLayout{app: app, wm: wm, icons: icons, screens: screenProvider}
	desktop.SetInstance(desk)
	desk.settings = newDeskSettings()
	desk.addSettingsChangeListener()
	desk.screenBackgroundMap = make(map[*desktop.Screen]fyne.CanvasObject)
	return desk
}

// NewEmbeddedDesktop creates a new windowed desktop for test purposes.
// An ApplicationProvider is used to lookup application icons from the operating system.
// If run during CI for testing it will return an in-memory window using the
// fyne/test package.
func NewEmbeddedDesktop(app fyne.App, icons desktop.ApplicationProvider) desktop.Desktop {
	desk := &deskLayout{app: app, icons: icons, screens: NewEmbeddedScreensProvider()}
	desktop.SetInstance(desk)
	desk.settings = newDeskSettings()
	desk.addSettingsChangeListener()
	desk.screenBackgroundMap = make(map[*desktop.Screen]fyne.CanvasObject)
	return desk
}
