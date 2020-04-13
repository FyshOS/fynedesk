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

const (
	// RootWindowName is the base string that all root windows will have in their title and is used to identify root windows.
	RootWindowName = "Fyne Desktop"
)

type deskLayout struct {
	app      fyne.App
	wm       desktop.WindowManager
	icons    desktop.ApplicationProvider
	screens  desktop.ScreenList
	settings desktop.DeskSettings

	backgroundScreenMap map[*background]*desktop.Screen

	bar            *bar
	widgets, mouse fyne.CanvasObject
	controlWin     fyne.Window
	primaryWin     fyne.Window
	roots          []fyne.Window
	refreshing     bool
}

func (l *deskLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	bg := objects[0].(*background)
	screen := l.backgroundScreenMap[bg]
	if screen == nil {
		return
	}
	bg.Resize(size)

	if screen == l.screens.Primary() {
		barHeight := l.bar.MinSize().Height
		l.bar.Resize(fyne.NewSize(size.Width, barHeight))
		l.bar.Move(fyne.NewPos(0, size.Height-barHeight))
		l.bar.Refresh()

		widgetsWidth := l.widgets.MinSize().Width
		l.widgets.Resize(fyne.NewSize(widgetsWidth, size.Height))
		l.widgets.Move(fyne.NewPos(size.Width-widgetsWidth, 0))
		l.widgets.Refresh()
	}
}

func (l *deskLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if l.wm == nil {
		return fyne.NewSize(1024, 576)
	}
	return fyne.NewSize(640, 480) // tiny - the window manager will scale up to screen size
}

func (l *deskLayout) newDesktopWindow(outputName string) fyne.Window {
	if l.wm == nil {
		win := l.app.NewWindow("Embedded " + RootWindowName)
		win.SetPadded(false)
		return win
	}

	desk := l.app.NewWindow(fmt.Sprintf("%s%s", RootWindowName, outputName))
	desk.SetPadded(false)
	desk.SetFullScreen(true)

	return desk
}

func (l *deskLayout) updateBackgrounds(path string) {
	for bg := range l.backgroundScreenMap {
		bg.updateBackgroundPath(path)
		canvas.Refresh(bg)
	}
}

func (l *deskLayout) createPrimaryContent() {
	l.bar = newBar(l)
	l.widgets = newWidgetPanel(l)
	l.mouse = newMouse()
}

func (l *deskLayout) createRoot(screen *desktop.Screen) {
	win := l.newDesktopWindow(screen.Name)
	l.roots = append(l.roots, win)
	bg := newBackground()
	l.backgroundScreenMap[bg] = screen
	if screen == l.screens.Primary() {
		l.primaryWin = win
		l.createPrimaryContent()
		win.SetOnClosed(func() {
			if !l.refreshing {
				l.controlWin.Close()
			}
		})
		win.SetContent(fyne.NewContainerWithLayout(l, bg, l.bar, l.widgets, l.mouse))
		l.mouse.Hide()
	} else {
		win.SetContent(fyne.NewContainerWithLayout(l, bg))
	}
	win.Show()
}

func (l *deskLayout) ensureSufficientRoots() {
	if len(l.screens.Screens()) >= len(l.roots) {
		diff := len(l.screens.Screens()) - len(l.roots)
		count := len(l.screens.Screens()) - diff - 1
		for i := 0; i < diff; i++ {
			l.createRoot(l.screens.Screens()[count])
			count++
		}
	} else {
		diff := len(l.roots) - len(l.screens.Screens())
		count := len(l.roots) - diff - 1
		for i := 0; i < diff; i++ {
			root := l.roots[count]
			root.SetOnClosed(nil)
			bg := root.Content().(*fyne.Container).Objects[0].(*background)
			delete(l.backgroundScreenMap, bg)
			root.Close()
		}
		l.roots = l.roots[:len(l.screens.Screens())]
	}
}

func (l *deskLayout) setupRoots() {
	if len(l.roots) == 0 {
		for _, screen := range l.screens.Screens() {
			l.createRoot(screen)
		}
		return
	}
	l.ensureSufficientRoots()
	for i, root := range l.roots {
		screen := l.screens.Screens()[i]
		root.Hide()
		root.SetOnClosed(nil)
		root.SetTitle(fmt.Sprintf("%s%s", RootWindowName, screen.Name))
		bg := root.Content().(*fyne.Container).Objects[0].(*background)
		l.backgroundScreenMap[bg] = screen
		if screen == l.screens.Primary() {
			l.primaryWin = root
			if l.bar == nil && l.widgets == nil && l.mouse == nil {
				l.createPrimaryContent()
			}
			root.SetOnClosed(func() {
				if !l.refreshing {
					l.controlWin.Close()
				}
			})
			root.SetContent(fyne.NewContainerWithLayout(l, bg, l.bar, l.widgets, l.mouse))
			l.mouse.Hide() // temporarily we do not draw mouse (using X default)
		} else {
			root.SetContent(fyne.NewContainerWithLayout(l, bg))
		}
		root.Show()
	}
}

func (l *deskLayout) Run() {
	if l.wm == nil {
		l.roots[0].ShowAndRun()
		return
	}
	if l.controlWin == nil {
		return
	}
	debug.SetPanicOnFault(true)

	defer func() {
		if r := recover(); r != nil {
			debug.PrintStack()
			if l.wm != nil {
				l.wm.Close() // attempt to close cleanly to leave X server running
			}
		}
	}()

	l.controlWin.ShowAndRun()
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
	// Qt toolkit cannot handle scale < 1
	positiveScale := math.Max(1.0, float64(scale))

	return []string{
		fmt.Sprintf("QT_SCALE_FACTOR=%1.1f", positiveScale),
		fmt.Sprintf("GDK_SCALE=%d", intScale),
		fmt.Sprintf("ELM_SCALE=%1.1f", scale),
	}
}

func (l *deskLayout) screensChanged() {
	l.refreshing = true
	l.setupRoots()
	l.refreshing = false
}

// MouseInNotify can be called by the window manager to alert the desktop that the cursor has entered the canvas
func (l *deskLayout) MouseInNotify(pos fyne.Position) {
	if l.bar == nil {
		return
	}
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
	if l.bar == nil {
		return
	}
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

func (l *deskLayout) setupInitialVars() {
	desktop.SetInstance(l)
	l.settings = newDeskSettings()
	l.addSettingsChangeListener()
	l.backgroundScreenMap = make(map[*background]*desktop.Screen)

	if l.wm != nil {
		l.controlWin = l.app.NewWindow(RootWindowName)
		l.controlWin.Resize(fyne.NewSize(1, 1))
		l.controlWin.SetMaster()
		l.controlWin.SetOnClosed(func() {
			l.wm.Close()
		})
	}

	l.setupRoots()
}

// NewDesktop creates a new desktop in fullscreen for main usage.
// The WindowManager passed in will be used to manage the screen it is loaded on.
// An ApplicationProvider is used to lookup application icons from the operating system.
func NewDesktop(app fyne.App, wm desktop.WindowManager, icons desktop.ApplicationProvider, screenProvider desktop.ScreenList) desktop.Desktop {
	desk := &deskLayout{app: app, wm: wm, icons: icons, screens: screenProvider}
	screenProvider.AddChangeListener(desk.screensChanged)
	desk.setupInitialVars()
	return desk
}

// NewEmbeddedDesktop creates a new windowed desktop for test purposes.
// An ApplicationProvider is used to lookup application icons from the operating system.
// If run during CI for testing it will return an in-memory window using the
// fyne/test package.
func NewEmbeddedDesktop(app fyne.App, icons desktop.ApplicationProvider) desktop.Desktop {
	desk := &deskLayout{app: app, icons: icons, screens: newEmbeddedScreensProvider()}
	desk.setupInitialVars()
	return desk
}
