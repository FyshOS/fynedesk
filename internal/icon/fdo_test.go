package icon

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"fyne.io/fyne"
	_ "fyne.io/fyne/test"

	"fyne.io/fynedesk"
)

var iconTheme = "default_theme"
var iconSize = 32

type dummyWindow struct {
	title    string
	command  string
	iconName string
	class    []string
}

func (w *dummyWindow) Decorated() bool {
	return true
}

func (w *dummyWindow) Title() string {
	return w.title
}

func (w *dummyWindow) Class() []string {
	return w.class
}

func (w *dummyWindow) Command() string {
	return w.command
}

func (w *dummyWindow) IconName() string {
	return w.iconName
}

func (w *dummyWindow) Icon() fyne.Resource {
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

func (*dummyWindow) RaiseAbove(fynedesk.Window) {
	// no-op (this is instructing the window after stack changes)
}

func (*dummyWindow) RaiseToTop() {
	// no-op
}

func exists(data fynedesk.AppData) bool {
	return data != nil && data.Icon(iconTheme, iconSize) != nil
}

func setTestEnv(t *testing.T) {
	workingDir, err := os.Getwd()
	if err != nil {
		fyne.LogError("Could not get current working directory", err)
		t.FailNow()
	}
	err = os.Setenv("XDG_DATA_DIRS", filepath.Join(workingDir, "testdata"))
	if err != nil {
		fyne.LogError("Could not set test environment variable", err)
		t.FailNow()
	}
}

//applications/app1.desktop and icons/default_theme/apps/32x32/app1.png
func TestFdoLookupDefaultTheme(t *testing.T) {
	setTestEnv(t)
	data := fdoLookupApplication("app1")
	assert.Equal(t, true, exists(data))
}

//applications/com.fyne.app.desktop and icons/default_theme/apps/scalable/app2.svg
func TestFdoFileNameMisMatchAndScalable(t *testing.T) {
	setTestEnv(t)
	data := fdoLookupApplication("app2")
	assert.Equal(t, true, exists(data))
}

//applications/app3.desktop and applications/app3.png
func TestFdoIconNameIsPath(t *testing.T) {
	setTestEnv(t)
	dataLocation := os.Getenv("XDG_DATA_DIRS")
	output := fmt.Sprintf("[Desktop Entry]\nName=App3\nExec=app3\nIcon=%s\n", filepath.Join(dataLocation, "icons", "app3.png"))
	err := ioutil.WriteFile(filepath.Join(dataLocation, "applications", "app3.desktop"), []byte(output), 0644)
	if err != nil {
		fyne.LogError("Could not create desktop for Icon Name path example", err)
		t.FailNow()
	}
	data := fdoLookupApplication("app3")
	assert.Equal(t, true, exists(data))
}

//applications/app4.desktop and pixmaps/app4.png
func TestFdoIconInPixmaps(t *testing.T) {
	setTestEnv(t)
	data := fdoLookupApplication("app4")
	assert.Equal(t, true, exists(data))
}

//applications/app5.desktop and icons/hicolor/32x32/apps/app5.png
func TestFdoIconHicolorFallback(t *testing.T) {
	setTestEnv(t)
	data := fdoLookupApplication("app5")
	assert.Equal(t, true, exists(data))
}

//applications/app6.desktop and icons/hicolor/scalable/apps/app6.svg
func TestFdoIconHicolorFallbackScalable(t *testing.T) {
	setTestEnv(t)
	data := fdoLookupApplication("app6")
	assert.Equal(t, true, exists(data))
}

//applications/app7.desktop and icons/default_theme/apps/16x16/app7.png
func TestFdoLookupDefaultThemeDifferentSize(t *testing.T) {
	setTestEnv(t)
	data := fdoLookupApplication("app7")
	assert.Equal(t, true, exists(data))
}

//applications/app8.desktop and icons/third_theme/apps/32/app8.png
func TestFdoLookupAnyThemeFallback(t *testing.T) {
	setTestEnv(t)
	data := fdoLookupApplication("app8")
	assert.Equal(t, true, exists(data))
}

//applications/xterm.desktop and icons/third_theme/emblems/16x16/app9.png
func TestFdoLookupIconNotInApps(t *testing.T) {
	setTestEnv(t)
	data := fdoLookupApplication("xterm")
	assert.Equal(t, true, exists(data))
}

//missing - not able to match
func TestFdoLookupMissing(t *testing.T) {
	setTestEnv(t)
	provider := NewFDOIconProvider()
	win1 := &dummyWindow{title: "NoMatch"}
	data := provider.FindAppFromWinInfo(win1)
	assert.Equal(t, false, exists(data))
}

func TestFdoLookupIconByWinInfo(t *testing.T) {
	setTestEnv(t)
	provider := NewFDOIconProvider()

	// Test win info lookup by title - should fail as titles are too common
	win1 := &dummyWindow{title: "App1"}
	data := provider.FindAppFromWinInfo(win1)
	assert.Equal(t, false, exists(data))
	// Test win info lookup by class
	win2 := &dummyWindow{class: []string{"App2", "app2"}}
	data = provider.FindAppFromWinInfo(win2)
	assert.Equal(t, true, exists(data))
	// Test win info lookup by command
	win3 := &dummyWindow{command: "app3"}
	data = provider.FindAppFromWinInfo(win3)
	assert.Equal(t, true, exists(data))
	// Test win info lookup by icon name
	win4 := &dummyWindow{iconName: "app4"}
	data = provider.FindAppFromWinInfo(win4)
	assert.Equal(t, true, exists(data))
}

func TestFdoLookupPartialMatches(t *testing.T) {
	setTestEnv(t)
	dataMatches := fdoLookupApplicationsMatching("app")
	assert.Equal(t, true, len(dataMatches) > 1)
	for _, data := range dataMatches {
		assert.Equal(t, true, exists(data))
	}
}

func TestFdoIconProvider_findOneAppFromNames(t *testing.T) {
	setTestEnv(t)
	single := findOneAppFromNames(NewFDOIconProvider(), "missing", "app1", "xterm")
	assert.NotNil(t, single)
	assert.Equal(t, "App1", single.Name())
}

func TestFdoIconProvider_DefaultApps(t *testing.T) {
	setTestEnv(t)
	defaults := NewFDOIconProvider().DefaultApps()
	assert.True(t, len(defaults) > 0)
}

func TestFdoExtractArgs(t *testing.T) {
	params := []string{"-u", "thing", "%U"}

	extracted := extractArgs(params)

	assert.Equal(t, []string{"-u", "thing"}, extracted)
}
