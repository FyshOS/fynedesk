package ui

import (
	"image/color"
	"testing"

	"fyne.io/fyne"
	"fyne.io/fyne/test"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"

	wmTest "fyne.io/fynedesk/test"
	wmTheme "fyne.io/fynedesk/theme"

	"github.com/stretchr/testify/assert"
)

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
	testBar := newBar(wmTest.NewDesktopWithWM(&embededWM{}))
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
	win := wmTest.NewWindow("")
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

func TestAppBarBackground(t *testing.T) {
	icons := []string{"fyne"}
	testBar := testBar(icons)
	testBar.disableTaskbar = true

	grad := test.WidgetRenderer(testBar).(*barRenderer).background
	assert.Equal(t, color.Transparent, grad.EndColor)
	assert.Equal(t, theme.BackgroundColor(), grad.StartColor)
	assert.Equal(t, testBar.iconSize+theme.Padding()*2, grad.Size().Width)
}

func TestIconsAndIconThemeChange(t *testing.T) {
	testBar := testBar(nil)

	assert.Equal(t, 0, len(testBar.icons))
	testBar.desk.Settings().(*wmTest.Settings).SetLauncherIcons([]string{"App1", "App2", "App3"})
	testBar.updateIconOrder()

	assert.Equal(t, 3, len(testBar.icons))

	testBar.desk.Settings().(*wmTest.Settings).SetIconTheme("Maximize")
	testBar.updateIcons()

	assert.Equal(t, "Maximize", testBar.desk.Settings().IconTheme())
	assert.Equal(t, wmTheme.MaximizeIcon, testBar.children[0].(*barIcon).resource)

	testBar.desk.Settings().(*wmTest.Settings).SetIconTheme("TestIconTheme")
	testBar.updateIcons()

	assert.Equal(t, "TestIconTheme", testBar.desk.Settings().IconTheme())
	assert.Equal(t, wmTheme.IconifyIcon, testBar.children[0].(*barIcon).resource)
}

func TestIconOrderChange(t *testing.T) {
	testBar := testBar(nil)

	assert.Equal(t, 0, len(testBar.icons))

	testBar.desk.Settings().(*wmTest.Settings).SetLauncherIcons([]string{"App1", "App2", "App3"})
	testBar.updateIconOrder()
	assert.Equal(t, "App1", testBar.children[0].(*barIcon).appData.Name())
	assert.Equal(t, "App2", testBar.children[1].(*barIcon).appData.Name())
	assert.Equal(t, "App3", testBar.children[2].(*barIcon).appData.Name())

	testBar.desk.Settings().(*wmTest.Settings).SetLauncherIcons([]string{"App3", "App1", "App2"})
	testBar.updateIconOrder()
	assert.Equal(t, "App3", testBar.children[0].(*barIcon).appData.Name())
	assert.Equal(t, "App1", testBar.children[1].(*barIcon).appData.Name())
	assert.Equal(t, "App2", testBar.children[2].(*barIcon).appData.Name())
}

func TestIconSizeChange(t *testing.T) {
	testBar := testBar(nil)

	testBar.desk.Settings().(*wmTest.Settings).SetLauncherIcons([]string{"App1", "App2", "App3"})
	testBar.updateIconOrder()

	assert.Equal(t, 32, testBar.icons[0].Size().Width)

	testBar.desk.Settings().(*wmTest.Settings).SetLauncherIconSize(64)
	testBar.iconSize = testBar.desk.Settings().LauncherIconSize()
	testBar.updateIcons()

	assert.Equal(t, 64, testBar.icons[0].Size().Width)
}

func TestZoomScaleChange(t *testing.T) {
	testBar := testBar(nil)

	testBar.desk.Settings().(*wmTest.Settings).SetLauncherIcons([]string{"App1", "App2", "App3"})
	testBar.updateIconOrder()

	testBar.mouseInside = true
	testBar.mousePosition = testBar.children[0].Position()
	widget.Refresh(testBar)
	firstWidth := testBar.children[0].Size().Width

	testBar.desk.Settings().(*wmTest.Settings).SetLauncherZoomScale(2.0)
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

	testBar.desk.Settings().(*wmTest.Settings).SetLauncherIcons([]string{"App1", "App2", "App3"})
	testBar.desk.Settings().(*wmTest.Settings).SetLauncherZoomScale(2.0)
	testBar.iconScale = float32(testBar.desk.Settings().LauncherZoomScale())
	testBar.updateIconOrder()

	testBar.mouseInside = true
	testBar.mousePosition = testBar.children[0].Position()
	widget.Refresh(testBar)

	width := testBar.children[0].Size().Width
	assert.NotEqual(t, testBar.desk.Settings().LauncherIconSize(), width)

	testBar.desk.Settings().(*wmTest.Settings).SetLauncherDisableZoom(true)
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

	testBar.desk.Settings().(*wmTest.Settings).SetLauncherIcons([]string{"App1", "App2", "App3"})
	testBar.updateIconOrder()

	separatorTest := false
	if len(testBar.icons) == len(testBar.children)-1 {
		separatorTest = true
	}
	assert.Equal(t, true, separatorTest)

	icon := testBar.createIcon(wmTest.NewAppData("dummy"), wmTest.NewWindow(""))
	testBar.append(icon)

	taskbarIconTest := false
	if testBar.children[len(testBar.children)-1].(*barIcon).windowData != nil {
		taskbarIconTest = true
	}
	assert.Equal(t, true, taskbarIconTest)

	testBar.desk.Settings().(*wmTest.Settings).SetLauncherDisableTaskbar(true)
	testBar.updateIconOrder()
	testBar.updateTaskbar()

	//Last Child at this point should not be the separator or a taskbar icon
	taskbarIconTest = false
	if testBar.children[len(testBar.children)-1].(*barIcon).windowData == nil {
		taskbarIconTest = true
	}
	assert.Equal(t, true, taskbarIconTest)
}
