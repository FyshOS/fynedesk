package desktop

import (
	"image/color"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	wmTheme "fyne.io/desktop/theme"
	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/widget"
)

type testDesk struct {
}

func (*testDesk) Root() fyne.Window {
	return nil
}

func (*testDesk) Run() {
}

func (*testDesk) RunApp(app AppData) error {
	return app.Run([]string{}) // no added env
}

func (*testDesk) Settings() DeskSettings {
	return &testSettings{}
}

func (*testDesk) ContentSizePixels(screen *Screen) (uint32, uint32) {
	return uint32(320), uint32(240)
}

func (*testDesk) IconProvider() ApplicationProvider {
	return nil
}

func (*testDesk) WindowManager() WindowManager {
	return nil
}

func (*testDesk) Screens() ScreenList {
	return nil
}

type testSettings struct {
	background             string
	iconTheme              string
	launcherIcons          []string
	launcherIconSize       int
	launcherZoomScale      float64
	launcherDisableZoom    bool
	launcherDisableTaskbar bool
}

func (ts *testSettings) IconTheme() string {
	return ts.iconTheme
}

func (ts *testSettings) Background() string {
	return ts.background
}

func (ts *testSettings) LauncherIcons() []string {
	return ts.launcherIcons
}

func (ts *testSettings) LauncherIconSize() int {
	if ts.launcherIconSize == 0 {
		return 32
	}
	return ts.launcherIconSize
}

func (ts *testSettings) LauncherDisableTaskbar() bool {
	return ts.launcherDisableTaskbar
}

func (ts *testSettings) LauncherDisableZoom() bool {
	return ts.launcherDisableZoom
}

func (ts *testSettings) LauncherZoomScale() float64 {
	if ts.launcherZoomScale == 0 {
		return 1.0
	}
	return ts.launcherZoomScale
}

func (*testSettings) AddChangeListener(listener chan DeskSettings) {
	return
}

type testScreensProvider struct {
	screens []*Screen
}

func (tsp testScreensProvider) Screens() []*Screen {
	if tsp.screens == nil {
		tsp.screens = []*Screen{{"Screen0", 0, 0, 2000, 1000}}
	}
	return tsp.screens
}

func (tsp testScreensProvider) Active() *Screen {
	return tsp.Screens()[0]
}

func (tsp testScreensProvider) Primary() *Screen {
	return tsp.Screens()[0]
}

func (tsp testScreensProvider) Scale() float32 {
	return 1.0
}

func (tsp testScreensProvider) ScreenForWindow(win Window) *Screen {
	return tsp.Screens()[0]
}

type testAppData struct {
	name string
}

func (tad *testAppData) Name() string {
	return tad.name
}

func (tad *testAppData) Run([]string) error {
	return nil
}

func (tad *testAppData) Icon(theme string, size int) fyne.Resource {
	if theme == "" {
		return nil
	} else if theme == "Maximize" {
		return wmTheme.MaximizeIcon
	}
	return wmTheme.IconifyIcon
}

type testAppProvider struct {
	screens []*Screen
}

func (tap *testAppProvider) AvailableApps() []AppData {
	return nil
}

func (tap *testAppProvider) AvailableThemes() []string {
	return nil
}

func (tap *testAppProvider) FindAppFromName(appName string) AppData {
	return &testAppData{name: appName}
}

func (tap *testAppProvider) FindAppFromWinInfo(win Window) AppData {
	return &testAppData{}
}

func (tap *testAppProvider) FindAppsMatching(pattern string) []AppData {
	return nil
}

func (tap *testAppProvider) DefaultApps() []AppData {
	return nil
}

func newTestAppProvider() ApplicationProvider {
	return &testAppProvider{}
}

func newTestScreensProvider() ScreenList {
	return &testScreensProvider{}
}

func newTestDesktop() *deskLayout {
	l := &deskLayout{}
	instance = l
	l.settings = &testSettings{}
	l.icons = newTestAppProvider()
	l.screens = newTestScreensProvider()
	l.backgrounds = []fyne.CanvasObject{newBackground()}
	l.screenBackgroundMap = make(map[*Screen]fyne.CanvasObject)
	l.screenBackgroundMap[l.Screens().Screens()[0]] = l.backgrounds[0]
	l.bar = newBar(l)
	l.widgets = canvas.NewRectangle(color.Black)

	return l
}

func TestDeskLayout_Layout(t *testing.T) {
	l := newTestDesktop()

	deskSize := fyne.NewSize(2000, 1000)
	l.Layout([]fyne.CanvasObject{l.backgrounds[0], l.bar, l.widgets}, deskSize)
	assert.Equal(t, l.backgrounds[0].Size(), deskSize)
	assert.Equal(t, l.widgets.Position().X+l.widgets.Size().Width, deskSize.Width)
	assert.Equal(t, l.widgets.Size().Height, deskSize.Height)
	assert.Equal(t, l.bar.Size().Width, deskSize.Width)
	assert.Equal(t, l.bar.Position().Y+l.bar.Size().Height, deskSize.Height)
}

func TestScaleVars(t *testing.T) {
	l := newTestDesktop()
	env := l.scaleVars(1.8)
	assert.Contains(t, env, "QT_SCALE_FACTOR=1.8")
	assert.Contains(t, env, "GDK_SCALE=2")
	assert.Contains(t, env, "ELM_SCALE=1.8")
}

func TestBackgroundChange(t *testing.T) {
	l := newTestDesktop()

	workingDir, err := os.Getwd()
	if err != nil {
		fyne.LogError("Could not get current working directory", err)
		t.FailNow()
	}
	assert.Equal(t, wmTheme.Background, l.backgrounds[0].(*canvas.Image).Resource)

	l.settings.(*testSettings).background = filepath.Join(workingDir, "testdata", "fyne.png")
	l.updateBackgrounds(l.Settings().Background())
	assert.Equal(t, l.settings.Background(), l.backgrounds[0].(*canvas.Image).File)
}

func TestIconsAndIconThemeChange(t *testing.T) {
	l := newTestDesktop()

	assert.Equal(t, 0, len(l.bar.(*bar).icons))
	l.settings.(*testSettings).launcherIcons = []string{"App1", "App2", "App3"}
	l.bar.(*bar).updateIconOrder()

	assert.Equal(t, 3, len(l.bar.(*bar).icons))

	l.settings.(*testSettings).iconTheme = "Maximize"
	l.bar.(*bar).updateIcons()

	assert.Equal(t, "Maximize", l.settings.IconTheme())
	assert.Equal(t, wmTheme.MaximizeIcon, l.bar.(*bar).children[0].(*barIcon).resource)

	l.settings.(*testSettings).iconTheme = "TestIconTheme"
	l.bar.(*bar).updateIcons()

	assert.Equal(t, "TestIconTheme", l.settings.IconTheme())
	assert.Equal(t, wmTheme.IconifyIcon, l.bar.(*bar).children[0].(*barIcon).resource)
}

func TestIconOrderChange(t *testing.T) {
	l := newTestDesktop()

	assert.Equal(t, 0, len(l.bar.(*bar).icons))

	l.settings.(*testSettings).launcherIcons = []string{"App1", "App2", "App3"}
	l.bar.(*bar).updateIconOrder()
	assert.Equal(t, "App1", l.bar.(*bar).children[0].(*barIcon).appData.Name())
	assert.Equal(t, "App2", l.bar.(*bar).children[1].(*barIcon).appData.Name())
	assert.Equal(t, "App3", l.bar.(*bar).children[2].(*barIcon).appData.Name())

	l.settings.(*testSettings).launcherIcons = []string{"App3", "App1", "App2"}
	l.bar.(*bar).updateIconOrder()
	assert.Equal(t, "App3", l.bar.(*bar).children[0].(*barIcon).appData.Name())
	assert.Equal(t, "App1", l.bar.(*bar).children[1].(*barIcon).appData.Name())
	assert.Equal(t, "App2", l.bar.(*bar).children[2].(*barIcon).appData.Name())
}

func TestIconSizeChange(t *testing.T) {
	l := newTestDesktop()

	l.settings.(*testSettings).launcherIcons = []string{"App1", "App2", "App3"}
	l.bar.(*bar).updateIconOrder()

	assert.Equal(t, 32, l.bar.(*bar).icons[0].Size().Width)

	l.settings.(*testSettings).launcherIconSize = 64
	l.bar.(*bar).iconSize = l.settings.LauncherIconSize()
	l.bar.(*bar).updateIcons()

	assert.Equal(t, 64, l.bar.(*bar).icons[0].Size().Width)
}

func TestZoomScaleChange(t *testing.T) {
	l := newTestDesktop()

	l.settings.(*testSettings).launcherIcons = []string{"App1", "App2", "App3"}
	l.bar.(*bar).updateIconOrder()

	l.bar.(*bar).mouseInside = true
	l.bar.(*bar).mousePosition = l.bar.(*bar).children[0].Position()
	widget.Refresh(l.bar.(*bar))
	firstWidth := l.bar.(*bar).children[0].Size().Width

	l.settings.(*testSettings).launcherZoomScale = 2.0
	l.bar.(*bar).iconScale = float32(l.settings.LauncherZoomScale())
	l.bar.(*bar).updateIcons()

	l.bar.(*bar).mouseInside = true
	l.bar.(*bar).mousePosition = l.bar.(*bar).children[0].Position()
	widget.Refresh(l.bar.(*bar))
	secondWidth := l.bar.(*bar).children[0].Size().Width

	zoomTest := false
	if secondWidth > firstWidth {
		zoomTest = true
	}
	assert.Equal(t, true, zoomTest)
}

func TestIconZoomDisabled(t *testing.T) {
	l := newTestDesktop()

	l.settings.(*testSettings).launcherIcons = []string{"App1", "App2", "App3"}
	l.settings.(*testSettings).launcherZoomScale = 2.0
	l.bar.(*bar).iconScale = float32(l.settings.LauncherZoomScale())
	l.bar.(*bar).updateIconOrder()

	l.bar.(*bar).mouseInside = true
	l.bar.(*bar).mousePosition = l.bar.(*bar).children[0].Position()
	widget.Refresh(l.bar.(*bar))

	width := l.bar.(*bar).children[0].Size().Width
	assert.NotEqual(t, l.settings.LauncherIconSize(), width)

	l.settings.(*testSettings).launcherDisableZoom = true
	l.bar.(*bar).disableZoom = true
	l.bar.(*bar).updateIconOrder()

	l.bar.(*bar).mouseInside = true
	l.bar.(*bar).mousePosition = l.bar.(*bar).children[0].Position()
	widget.Refresh(l.bar.(*bar))

	width = l.bar.(*bar).children[0].Size().Width
	assert.Equal(t, l.settings.LauncherIconSize(), width)
}

func TestIconTaskbarDisabled(t *testing.T) {
	l := newTestDesktop()

	l.settings.(*testSettings).launcherIcons = []string{"App1", "App2", "App3"}
	l.bar.(*bar).updateIconOrder()

	separatorTest := false
	if len(l.bar.(*bar).icons) == len(l.bar.(*bar).children)-1 {
		separatorTest = true
	}
	assert.Equal(t, true, separatorTest)

	icon := barCreateIcon(l.bar.(*bar), true, &testAppData{}, &dummyWindow{})
	l.bar.(*bar).append(icon)

	taskbarIconTest := false
	if l.bar.(*bar).children[len(l.bar.(*bar).children)-1].(*barIcon).taskbarWindow != nil {
		taskbarIconTest = true
	}
	assert.Equal(t, true, taskbarIconTest)

	l.settings.(*testSettings).launcherDisableTaskbar = true
	l.bar.(*bar).updateIconOrder()
	l.bar.(*bar).updateTaskbar()

	//Last Child at this point should not be the separator or a taskbar icon
	taskbarIconTest = false
	if l.bar.(*bar).children[len(l.bar.(*bar).children)-1].(*barIcon).taskbarWindow == nil {
		taskbarIconTest = true
	}
	assert.Equal(t, true, taskbarIconTest)
}
