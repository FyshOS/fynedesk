package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"fyne.io/fyne/test"
	"fyne.io/fyne/theme"
	"github.com/stretchr/testify/assert"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"

	"fyne.io/fynedesk"
	wmTheme "fyne.io/fynedesk/theme"
)

type testDesk struct {
	settings fynedesk.DeskSettings
	icons    fynedesk.ApplicationProvider
	screens  fynedesk.ScreenList
}

func (*testDesk) Root() fyne.Window {
	return test.NewWindow(nil)
}

func (*testDesk) Run() {
}

func (*testDesk) RunApp(app fynedesk.AppData) error {
	return app.Run([]string{}) // no added env
}

func (td *testDesk) Settings() fynedesk.DeskSettings {
	return td.settings
}

func (*testDesk) ContentSizePixels(screen *fynedesk.Screen) (uint32, uint32) {
	return uint32(320), uint32(240)
}

func (td *testDesk) IconProvider() fynedesk.ApplicationProvider {
	return td.icons
}

func (*testDesk) WindowManager() fynedesk.WindowManager {
	return nil
}

func (td *testDesk) Screens() fynedesk.ScreenList {
	return td.screens
}

func (*testDesk) Modules() []fynedesk.Module {
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

	moduleNames []string
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

func (ts *testSettings) ModuleNames() []string {
	return ts.moduleNames
}

func (*testSettings) AddChangeListener(listener chan fynedesk.DeskSettings) {
	return
}

type testScreensProvider struct {
	screens []*fynedesk.Screen
	primary *fynedesk.Screen
	active  *fynedesk.Screen
}

func (tsp testScreensProvider) RefreshScreens() {
	return
}

func (tsp testScreensProvider) AddChangeListener(func()) {
	// no-op
}

func (tsp testScreensProvider) Screens() []*fynedesk.Screen {
	return tsp.screens
}

func (tsp testScreensProvider) Active() *fynedesk.Screen {
	return tsp.screens[0]
}

func (tsp testScreensProvider) Primary() *fynedesk.Screen {
	return tsp.screens[0]
}

func (tsp testScreensProvider) Scale() float32 {
	return 1.0
}

func (tsp testScreensProvider) ScreenForWindow(win fynedesk.Window) *fynedesk.Screen {
	return tsp.Screens()[0]
}

func (tsp testScreensProvider) ScreenForGeometry(x int, y int, width int, height int) *fynedesk.Screen {
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
	screens []*fynedesk.Screen
	apps    []fynedesk.AppData
}

func (tap *testAppProvider) AvailableApps() []fynedesk.AppData {
	return tap.apps
}

func (tap *testAppProvider) AvailableThemes() []string {
	return nil
}

func (tap *testAppProvider) FindAppFromName(appName string) fynedesk.AppData {
	return &testAppData{name: appName}
}

func (tap *testAppProvider) FindAppFromWinInfo(win fynedesk.Window) fynedesk.AppData {
	return &testAppData{}
}

func (tap *testAppProvider) FindAppsMatching(pattern string) []fynedesk.AppData {
	var ret []fynedesk.AppData
	for _, app := range tap.apps {
		if !strings.Contains(strings.ToLower(app.Name()), strings.ToLower(pattern)) {
			continue
		}

		ret = append(ret, app)
	}

	return ret
}

func (tap *testAppProvider) DefaultApps() []fynedesk.AppData {
	return nil
}

func newTestAppProvider(appNames []string) *testAppProvider {
	provider := &testAppProvider{}

	for _, name := range appNames {
		provider.apps = append(provider.apps, &testAppData{name: name})
	}

	return provider
}

func TestDeskLayout_Layout(t *testing.T) {
	l := &deskLayout{screens: &testScreensProvider{screens: []*fynedesk.Screen{{Name: "Screen0", X: 0, Y: 0,
		Width: 2000, Height: 1000, Scale: 1.0}}}}
	l.bar = testBar([]string{})
	l.widgets = newWidgetPanel(l)
	bg := &background{wallpaper: canvas.NewImageFromResource(theme.FyneLogo())}
	l.backgroundScreenMap = make(map[*background]*fynedesk.Screen)
	l.backgroundScreenMap[bg] = l.screens.Primary()
	deskSize := fyne.NewSize(2000, 1000)

	l.Layout([]fyne.CanvasObject{bg, l.bar, l.widgets}, deskSize)

	assert.Equal(t, bg.Size(), deskSize)
	assert.Equal(t, l.widgets.Position().X+l.widgets.Size().Width, deskSize.Width)
	assert.Equal(t, l.widgets.Size().Height, deskSize.Height)
	assert.Equal(t, l.bar.Size().Width, deskSize.Width)
	assert.Equal(t, l.bar.Position().Y+l.bar.Size().Height, deskSize.Height)
}

func TestScaleVars_Up(t *testing.T) {
	l := &deskLayout{}
	l.screens = &testScreensProvider{}
	env := l.scaleVars(1.8)
	assert.Contains(t, env, "QT_SCALE_FACTOR=1.8")
	assert.Contains(t, env, "GDK_SCALE=2")
	assert.Contains(t, env, "ELM_SCALE=1.8")
}

func TestScaleVars_Down(t *testing.T) {
	l := &deskLayout{}
	l.screens = &testScreensProvider{}
	env := l.scaleVars(0.9)
	assert.Contains(t, env, "QT_SCALE_FACTOR=1.0")
	assert.Contains(t, env, "GDK_SCALE=1")
	assert.Contains(t, env, "ELM_SCALE=0.9")
}

func TestBackgroundChange(t *testing.T) {
	l := &deskLayout{screens: &testScreensProvider{screens: []*fynedesk.Screen{{Name: "Screen0", X: 0, Y: 0,
		Width: 2000, Height: 1000, Scale: 1.0}}}}
	fynedesk.SetInstance(l)
	l.settings = &testSettings{}
	bg := newBackground()
	l.backgroundScreenMap = make(map[*background]*fynedesk.Screen)
	l.backgroundScreenMap[bg] = l.screens.Primary()

	workingDir, err := os.Getwd()
	if err != nil {
		fyne.LogError("Could not get current working directory", err)
		t.FailNow()
	}
	assert.Equal(t, wmTheme.Background, bg.wallpaper.Resource)

	l.settings.(*testSettings).background = filepath.Join(workingDir, "testdata", "fyne.png")
	l.updateBackgrounds(l.Settings().Background())
	assert.Equal(t, l.settings.Background(), bg.wallpaper.File)
}
