package desktop

import (
	"testing"

	"fyne.io/fyne"
	"github.com/stretchr/testify/assert"

	_ "fyne.io/fyne/test"
	"fyne.io/fyne/widget"

	wmTheme "fyne.io/desktop/theme"
)

type dummyWindow struct {
}

func (*dummyWindow) Decorated() bool {
	return true
}

func (*dummyWindow) Title() string {
	return "Xterm"
}

func (*dummyWindow) Class() []string {
	return []string{"Xterm", "xterm"}
}

func (*dummyWindow) Command() string {
	return "xterm"
}

func (*dummyWindow) IconName() string {
	return "xterm"
}

func (*dummyWindow) Icon() fyne.Resource {
	return nil
}

func (*dummyWindow) Fullscreened() bool {
	return false
}

func (*dummyWindow) Iconic() bool {
	return false
}

func (*dummyWindow) Maximized() bool {
	return false
}

func (*dummyWindow) TopWindow() bool {
	return true
}

func (*dummyWindow) SkipTaskbar() bool {
	return false
}

func (*dummyWindow) Focused() bool {
	return false
}

func (*dummyWindow) Focus() {
	// no-op
}

func (*dummyWindow) Close() {
	// no-op
}

func (*dummyWindow) Fullscreen() {
	// no-op
}

func (*dummyWindow) Unfullscreen() {
	// no-op
}

func (*dummyWindow) Iconify() {
	// no-op
}

func (*dummyWindow) Uniconify() {
	// no-op
}

func (*dummyWindow) Maximize() {
	// no-op
}

func (*dummyWindow) Unmaximize() {
	// no-op
}

func (*dummyWindow) RaiseAbove(Window) {
	// no-op (this is instructing the window after stack changes)
}

func (*dummyWindow) RaiseToTop() {
	// no-op
}

type dummyIcon struct {
}

func (d *dummyIcon) Name() string {
	return "Fyne"
}

func (d *dummyIcon) Icon(theme string, size int) fyne.Resource {
	return fyne.NewStaticResource("test.png", []byte{})
}

func (d *dummyIcon) Run([]string) error {
	// no-op
	return nil
}

func testBar(icons []string) *bar {
	testBar := newAppBar(&testDesk{settings: &testSettings{}, icons: &testAppProvider{}})
	for range icons {
		icon := barCreateIcon(testBar, false, &dummyIcon{}, nil)
		if icon != nil {
			testBar.append(icon)
		}
	}
	appBar = testBar
	return testBar
}

func TestAppBar_Append(t *testing.T) {
	icons := []string{"fyne", "fyne", "fyne", "fyne"}
	testBar := testBar(icons)
	assert.Equal(t, len(icons), len(testBar.children))
	testBar.appendSeparator()
	assert.Equal(t, len(icons)+1, len(testBar.children))
	win := &dummyWindow{}
	icon := barCreateIcon(testBar, true, &dummyIcon{}, win)
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
	testBar.mousePosition = testBar.children[0].Position()
	widget.Refresh(testBar)
	zoomTest := false
	if testBar.children[0].Size().Width > testBar.children[1].Size().Width {
		zoomTest = true
	}
	assert.Equal(t, true, zoomTest)
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

	icon := barCreateIcon(testBar, true, &testAppData{}, &dummyWindow{})
	testBar.append(icon)

	taskbarIconTest := false
	if testBar.children[len(testBar.children)-1].(*barIcon).taskbarWindow != nil {
		taskbarIconTest = true
	}
	assert.Equal(t, true, taskbarIconTest)

	testBar.desk.Settings().(*testSettings).launcherDisableTaskbar = true
	testBar.updateIconOrder()
	testBar.updateTaskbar()

	//Last Child at this point should not be the separator or a taskbar icon
	taskbarIconTest = false
	if testBar.children[len(testBar.children)-1].(*barIcon).taskbarWindow == nil {
		taskbarIconTest = true
	}
	assert.Equal(t, true, taskbarIconTest)
}
