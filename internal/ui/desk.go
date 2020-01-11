package ui

import (
	"fmt"
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
	roots    []fyne.Window
	wm       desktop.WindowManager
	icons    desktop.ApplicationProvider
	screens  desktop.ScreenList
	settings desktop.DeskSettings

	screenRootMap       map[*desktop.Screen]fyne.Window
	backgroundScreenMap map[*background]*desktop.Screen

	bar            *bar
	widgets, mouse fyne.CanvasObject

	uniqueRootID int
}

func (l *deskLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	bg := objects[0]
	screen := l.backgroundScreenMap[bg.(*background)]
	if screen == nil {
		return
	}
	bg.Resize(size)

	if screen == l.Screens().Primary() {
		barHeight := l.bar.MinSize().Height
		l.bar.Resize(fyne.NewSize(size.Width, barHeight))
		l.bar.Move(fyne.NewPos(0, size.Height-barHeight))

		widgetsWidth := l.widgets.MinSize().Width
		l.widgets.Resize(fyne.NewSize(widgetsWidth, size.Height))
		l.widgets.Move(fyne.NewPos(size.Width-widgetsWidth, 0))
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

	desk := l.app.NewWindow(fmt.Sprintf("Fyne Desktop%d", l.uniqueRootID))
	l.uniqueRootID++
	desk.SetPadded(false)
	desk.FullScreen()

	return desk
}

func (l *deskLayout) updateBackgrounds(path string) {
	for bg := range l.backgroundScreenMap {
		bg.updateBackgroundPath(path)
		canvas.Refresh(bg)
	}
}

func (l *deskLayout) RootForScreen(screen *desktop.Screen) fyne.Window {
	return l.screenRootMap[screen]
}

func (l *deskLayout) setupRoots() {
	if len(l.roots) != 0 {
		return
	}
	for _, screen := range l.screens.Screens() {
		win := l.newDesktopWindow()
		l.screenRootMap[screen] = win
		l.roots = append(l.roots, win)
		win.Canvas().SetScale(screen.CanvasScale())
		bg := newBackground()
		l.backgroundScreenMap[bg] = screen
		var container *fyne.Container
		if screen == l.screens.Primary() {
			l.bar = newBar(l)
			l.widgets = newWidgetPanel(l)
			l.mouse = newMouse()
			l.mouse.Hide() // temporarily we do not draw mouse (using X default)
			container = fyne.NewContainerWithLayout(l, bg, l.bar, l.widgets, l.mouse)
			if l.wm != nil {
				win.SetOnClosed(func() {
					l.wm.Close()
				})
			}
			l.mouse.Hide() // temporarily we do not draw mouse (using X default)
		} else {
			container = fyne.NewContainerWithLayout(l, bg)
		}
		win.SetContent(container)
	}
}

func (l *deskLayout) Run() {
	if l.wm == nil {
		l.RootForScreen(l.screens.Primary()).ShowAndRun()
		return
	}
	debug.SetPanicOnFault(true)

	defer func() {
		if r := recover(); r != nil {
			fyne.LogError("Crashed: "+string(debug.Stack()), nil)
			if l.wm != nil {
				l.wm.Close() // attempt to close cleanly to leave X server running
			}
		}
	}()

	for _, win := range l.roots {
		if win == l.RootForScreen(l.screens.Primary()) {
			continue
		}
		win.Show()
	}
	l.RootForScreen(l.screens.Primary()).ShowAndRun()
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

func setupInitialVars(desk *deskLayout) {
	desktop.SetInstance(desk)
	desk.settings = newDeskSettings()
	desk.addSettingsChangeListener()
	desk.screenRootMap = make(map[*desktop.Screen]fyne.Window)
	desk.backgroundScreenMap = make(map[*background]*desktop.Screen)
	desk.setupRoots()
}

// NewDesktop creates a new desktop in fullscreen for main usage.
// The WindowManager passed in will be used to manage the screen it is loaded on.
// An ApplicationProvider is used to lookup application icons from the operating system.
func NewDesktop(app fyne.App, wm desktop.WindowManager, icons desktop.ApplicationProvider, screenProvider desktop.ScreenList) desktop.Desktop {
	desk := &deskLayout{app: app, wm: wm, icons: icons, screens: screenProvider}
	setupInitialVars(desk)
	return desk
}

// NewEmbeddedDesktop creates a new windowed desktop for test purposes.
// An ApplicationProvider is used to lookup application icons from the operating system.
// If run during CI for testing it will return an in-memory window using the
// fyne/test package.
func NewEmbeddedDesktop(app fyne.App, icons desktop.ApplicationProvider) desktop.Desktop {
	desk := &deskLayout{app: app, icons: icons, screens: newEmbeddedScreensProvider()}
	setupInitialVars(desk)
	return desk
}
