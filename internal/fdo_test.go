package internal

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"fyne.io/desktop"
	"fyne.io/fyne"
	_ "fyne.io/fyne/test"
	"github.com/magiconair/properties/assert"
)

var iconTheme = "default_theme"
var iconSize = 32

type dummyWindow struct {
}

type dummyWindow1 struct {
	dummyWindow
}

type dummyWindow2 struct {
	dummyWindow
}

type dummyWindow3 struct {
	dummyWindow
}

type dummyWindow4 struct {
	dummyWindow
}

func (*dummyWindow) Decorated() bool {
	return true
}

func (*dummyWindow) Title() string {
	return ""
}

func (*dummyWindow1) Title() string {
	return "App1"
}

func (*dummyWindow) Class() []string {
	return []string{"", ""}
}

func (*dummyWindow2) Class() []string {
	return []string{"App2", "app2"}
}

func (*dummyWindow) Command() string {
	return ""
}

func (*dummyWindow3) Command() string {
	return "app3"
}

func (*dummyWindow) IconName() string {
	return ""
}

func (*dummyWindow4) IconName() string {
	return "app4"
}

func (*dummyWindow) Focus() {
	// no-op
}

func (*dummyWindow) Close() {
	// no-op
}

func (*dummyWindow) RaiseAbove(desktop.Window) {
	// no-op (this is instructing the window after stack changes)
}

func exists(data desktop.IconData) bool {
	var test bool
	if data != nil && data.IconPath() != "" {
		if _, err := os.Stat(data.IconPath()); err == nil {
			test = true
		}
	}
	return test
}

func setTestEnv() {
	workingDir, err := os.Getwd()
	if err != nil {
		fyne.LogError("Could not get current working directory", err)
	}
	err = os.Setenv("XDG_DATA_DIRS", filepath.Join(workingDir, "testdata"))
	if err != nil {
		fyne.LogError("Could not set test environment variable", err)
	}
}

//applications/app1.desktop and icons/default_theme/apps/32x32/app1.png
func TestFdoLookupDefaultTheme(t *testing.T) {
	setTestEnv()
	data := fdoLookupApplication(iconTheme, iconSize, "app1")
	assert.Equal(t, exists(data), true)
}

//applications/com.fyne.app.desktop and icons/default_theme/apps/scalable/app2.svg
func TestFdoFileNameMisMatchAndScalable(t *testing.T) {
	setTestEnv()
	data := fdoLookupApplication(iconTheme, iconSize, "app2")
	assert.Equal(t, exists(data), true)
}

//applications/app3.desktop and applications/app3.png
func TestFdoIconNameIsPath(t *testing.T) {
	setTestEnv()
	dataLocation := os.Getenv("XDG_DATA_DIRS")
	output := fmt.Sprintf("[Desktop Entry]\nName=App3\nExec=app3\nIcon=%s\n", filepath.Join(dataLocation, "icons", "app3.png"))
	fmt.Print(output)
	err := ioutil.WriteFile(filepath.Join(dataLocation, "applications", "app3.desktop"), []byte(output), 0644)
	if err != nil {
		fyne.LogError("Could not create desktop for Icon Name path example", err)
	}
	data := fdoLookupApplication(iconTheme, iconSize, "app3")
	assert.Equal(t, exists(data), true)
}

//applications/app4.desktop and pixmaps/app4.png
func TestFdoIconInPixmaps(t *testing.T) {
	setTestEnv()
	data := fdoLookupApplication(iconTheme, iconSize, "app4")
	assert.Equal(t, exists(data), true)
}

//applications/app5.desktop and icons/hicolor/32x32/apps/app5.png
func TestFdoIconHicolorFallback(t *testing.T) {
	setTestEnv()
	data := fdoLookupApplication(iconTheme, iconSize, "app5")
	assert.Equal(t, exists(data), true)
}

//applications/app6.desktop and icons/hicolor/scalable/apps/app6.svg
func TestFdoIconHicolorFallbackScalable(t *testing.T) {
	setTestEnv()
	data := fdoLookupApplication(iconTheme, iconSize, "app6")
	assert.Equal(t, exists(data), true)
}

//applications/app7.desktop and icons/default_theme/apps/16x16/app7.png
func TestFdoLookupDefaultThemeDifferentSize(t *testing.T) {
	setTestEnv()
	data := fdoLookupApplication(iconTheme, iconSize, "app7")
	assert.Equal(t, exists(data), true)
}

//applications/app8.desktop and icons/third_theme/apps/32/app8.png
func TestFdoLookupAnyThemeFallback(t *testing.T) {
	setTestEnv()
	data := fdoLookupApplication(iconTheme, iconSize, "app8")
	assert.Equal(t, exists(data), true)
}

//applications/app9.desktop and icons/third_theme/emblems/16x16/app9.png
func TestFdoLookupIconNotInApps(t *testing.T) {
	setTestEnv()
	data := fdoLookupApplication(iconTheme, iconSize, "app9")
	assert.Equal(t, exists(data), true)
}

func TestFdoLookupIconByWinInfo(t *testing.T) {
	setTestEnv()
	//Test win info lookup by title
	win1 := &dummyWindow1{}
	data := fdoLookupApplicationWinInfo(iconTheme, iconSize, win1)
	assert.Equal(t, exists(data), true)
	//Test win info lookup by class
	win2 := &dummyWindow2{}
	data = fdoLookupApplicationWinInfo(iconTheme, iconSize, win2)
	assert.Equal(t, exists(data), true)
	//Test win info lookup by command
	win3 := &dummyWindow3{}
	data = fdoLookupApplicationWinInfo(iconTheme, iconSize, win3)
	assert.Equal(t, exists(data), true)
	//Test win info lookup by icon name
	win4 := &dummyWindow4{}
	data = fdoLookupApplicationWinInfo(iconTheme, iconSize, win4)
	assert.Equal(t, exists(data), true)
}

func TestFdoLookupPartialMatch(t *testing.T) {
	dataMatches := fdoLookupApplicationPartialMatch(iconTheme, iconSize, "app")
	assert.Equal(t, len(dataMatches) > 1, true)
	for _, data := range dataMatches {
		assert.Equal(t, exists(data), true)
	}
}
