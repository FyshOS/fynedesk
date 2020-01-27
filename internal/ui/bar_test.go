package ui

import (
	"strings"
	"testing"

	"fyne.io/fyne"
	"github.com/stretchr/testify/assert"

	_ "fyne.io/fyne/test"
	"fyne.io/fyne/widget"

	"fyne.io/desktop"
	wmTheme "fyne.io/desktop/theme"
)

type dummyWindow struct {
	name   string // a name to override the dummy default
	raised bool   // this flag shows we were requested to raise above the other windows
}

func (w *dummyWindow) Decorated() bool {
	return true
}

func (w *dummyWindow) Title() string {
	if w.name == "" {
		return "Xterm"
	}

	return w.name
}

func (w *dummyWindow) Class() []string {
	return []string{w.Title(), "xterm"}
}

func (w *dummyWindow) Command() string {
	return strings.ToLower(w.Title())
}

func (w *dummyWindow) IconName() string {
	return strings.ToLower(w.Title())
}

func (w *dummyWindow) Icon() fyne.Resource {
	return nil
}

func (w *dummyWindow) Fullscreened() bool {
	return false
}

func (w *dummyWindow) Iconic() bool {
	return false
}

func (w *dummyWindow) Maximized() bool {
	return false
}

func (w *dummyWindow) TopWindow() bool {
	return true
}

func (w *dummyWindow) SkipTaskbar() bool {
	return false
}

func (w *dummyWindow) Focused() bool {
	return false
}

func (w *dummyWindow) Focus() {
	// no-op
}

func (w *dummyWindow) Close() {
	// no-op
}

func (w *dummyWindow) Fullscreen() {
	// no-op
}

func (w *dummyWindow) Unfullscreen() {
	// no-op
}

func (w *dummyWindow) Iconify() {
	// no-op
}

func (w *dummyWindow) Uniconify() {
	// no-op
}

func (w *dummyWindow) Maximize() {
	// no-op
}

func (w *dummyWindow) Unmaximize() {
	// no-op
}

func (w *dummyWindow) RaiseAbove(desktop.Window) {
	// no-op (this is instructing the window after stack changes)
}

func (w *dummyWindow) RaiseToTop() {
	w.raised = true
}

type dummyIcon struct {
	name string
}

func (d *dummyIcon) Name() string {
	return d.name
}

func (d *dummyIcon) Icon(theme string, size int) fyne.Resource {
	return fyne.NewStaticResource("test.png", []byte{})
}

func (d *dummyIcon) Run([]string) error {
	// no-op
	return nil
}

func testBar(icons []string) *bar {
	testBar := newBar(&testDesk{settings: &testSettings{}, icons: &testAppProvider{}})
	testBar.children = []fyne.CanvasObject{} // remove divider, then we add it again later
	for _, name := range icons {
		icon := testBar.createIcon(&dummyIcon{name: name}, nil)
		if icon != nil {
			testBar.append(icon)
		}
	}
	return testBar
}

func TestAppBar_Append(t *testing.T) {
	icons := []string{"fyne", "fyne", "fyne", "fyne"}
	testBar := testBar(icons)
	assert.Equal(t, len(icons), len(testBar.children))
	testBar.appendSeparator()
	assert.Equal(t, len(icons)+1, len(testBar.children))
	win := &dummyWindow{}
	icon := testBar.createIcon(&dummyIcon{}, win)
	testBar.append(icon)
	assert.Equal(t, len(icons)+2, len(testBar.children))
	testBar.removeFromTaskbar(icon)
	assert.Equal(t, len(icons)+1, len(testBar.children))
}

func TestAppBar_Zoom(t *testing.T) {
	icons := []string{"fyne", "fyne", "fyne", "fyne"}
	testBar := testBar(icons)
	testBar.disableZoom = false
	testBar.iconSize = 32
	testBar.iconScale = 2.0
	testBar.mouseInside = true
	testBar.mousePosition = testBar.children[0].Position().Add(fyne.NewPos(5, 5))
	testBar.Refresh()
	assert.Equal(t, true, testBar.children[0].Size().Width > testBar.children[1].Size().Width)
}

func TestIconsAndIconThemeChange(t *testing.T) {
	testBar := testBar(nil)

	assert.Equal(t, 0, len(testBar.icons))
	testBar.desk.Settings().(*testSettings).launcherIcons = []string{"App1", "App2", "App3"}
	testBar.updateIconOrder()

	assert.Equal(t, 3, len(testBar.icons))

	testBar.desk.Settings().(*testSettings).iconTheme = "Maximize"
	testBar.updateIcons()

	assert.Equal(t, "Maximize", testBar.desk.Settings().IconTheme())
	assert.Equal(t, wmTheme.MaximizeIcon, testBar.children[0].(*barIcon).resource)

	testBar.desk.Settings().(*testSettings).iconTheme = "TestIconTheme"
	testBar.updateIcons()

	assert.Equal(t, "TestIconTheme", testBar.desk.Settings().IconTheme())
	assert.Equal(t, wmTheme.IconifyIcon, testBar.children[0].(*barIcon).resource)
}

func TestIconOrderChange(t *testing.T) {
	testBar := testBar(nil)

	assert.Equal(t, 0, len(testBar.icons))

	testBar.desk.Settings().(*testSettings).launcherIcons = []string{"App1", "App2", "App3"}
	testBar.updateIconOrder()
	assert.Equal(t, "App1", testBar.children[0].(*barIcon).appData.Name())
	assert.Equal(t, "App2", testBar.children[1].(*barIcon).appData.Name())
	assert.Equal(t, "App3", testBar.children[2].(*barIcon).appData.Name())

	testBar.desk.Settings().(*testSettings).launcherIcons = []string{"App3", "App1", "App2"}
	testBar.updateIconOrder()
	assert.Equal(t, "App3", testBar.children[0].(*barIcon).appData.Name())
	assert.Equal(t, "App1", testBar.children[1].(*barIcon).appData.Name())
	assert.Equal(t, "App2", testBar.children[2].(*barIcon).appData.Name())
}

func TestIconSizeChange(t *testing.T) {
	testBar := testBar(nil)

	testBar.desk.Settings().(*testSettings).launcherIcons = []string{"App1", "App2", "App3"}
	testBar.updateIconOrder()

	assert.Equal(t, 32, testBar.icons[0].Size().Width)

	testBar.desk.Settings().(*testSettings).launcherIconSize = 64
	testBar.iconSize = testBar.desk.Settings().LauncherIconSize()
	testBar.updateIcons()

	assert.Equal(t, 64, testBar.icons[0].Size().Width)
}

func TestZoomScaleChange(t *testing.T) {
	testBar := testBar(nil)

	testBar.desk.Settings().(*testSettings).launcherIcons = []string{"App1", "App2", "App3"}
	testBar.updateIconOrder()

	testBar.mouseInside = true
	testBar.mousePosition = testBar.children[0].Position()
	widget.Refresh(testBar)
	firstWidth := testBar.children[0].Size().Width

	testBar.desk.Settings().(*testSettings).launcherZoomScale = 2.0
	testBar.iconScale = float32(testBar.desk.Settings().LauncherZoomScale())
	testBar.updateIcons()

	testBar.mouseInside = true
	testBar.mousePosition = testBar.children[0].Position()
	widget.Refresh(testBar)
	secondWidth := testBar.children[0].Size().Width

	zoomTest := false
	if secondWidth > firstWidth {
		zoomTest = true
	}
	assert.Equal(t, true, zoomTest)
}

func TestIconZoomDisabled(t *testing.T) {
	testBar := testBar(nil)

	testBar.desk.Settings().(*testSettings).launcherIcons = []string{"App1", "App2", "App3"}
	testBar.desk.Settings().(*testSettings).launcherZoomScale = 2.0
	testBar.iconScale = float32(testBar.desk.Settings().LauncherZoomScale())
	testBar.updateIconOrder()

	testBar.mouseInside = true
	testBar.mousePosition = testBar.children[0].Position()
	widget.Refresh(testBar)

	width := testBar.children[0].Size().Width
	assert.NotEqual(t, testBar.desk.Settings().LauncherIconSize(), width)

	testBar.desk.Settings().(*testSettings).launcherDisableZoom = true
	testBar.disableZoom = true
	testBar.updateIconOrder()

	testBar.mouseInside = true
	testBar.mousePosition = testBar.children[0].Position()
	widget.Refresh(testBar)

	width = testBar.children[0].Size().Width
	assert.Equal(t, testBar.desk.Settings().LauncherIconSize(), width)
}

func TestIconTaskbarDisabled(t *testing.T) {
	testBar := testBar(nil)

	testBar.desk.Settings().(*testSettings).launcherIcons = []string{"App1", "App2", "App3"}
	testBar.updateIconOrder()

	separatorTest := false
	if len(testBar.icons) == len(testBar.children)-1 {
		separatorTest = true
	}
	assert.Equal(t, true, separatorTest)

	icon := testBar.createIcon(&testAppData{}, &dummyWindow{})
	testBar.append(icon)

	taskbarIconTest := false
	if testBar.children[len(testBar.children)-1].(*barIcon).windowData != nil {
		taskbarIconTest = true
	}
	assert.Equal(t, true, taskbarIconTest)

	testBar.desk.Settings().(*testSettings).launcherDisableTaskbar = true
	testBar.updateIconOrder()
	testBar.updateTaskbar()

	//Last Child at this point should not be the separator or a taskbar icon
	taskbarIconTest = false
	if testBar.children[len(testBar.children)-1].(*barIcon).windowData == nil {
		taskbarIconTest = true
	}
	assert.Equal(t, true, taskbarIconTest)
}
