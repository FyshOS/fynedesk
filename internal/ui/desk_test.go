package ui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"

	"fyne.io/fynedesk"
	wmTest "fyne.io/fynedesk/test"
	wmTheme "fyne.io/fynedesk/theme"
)

func TestDeskLayout_Layout(t *testing.T) {
	l := &desktop{screens: wmTest.NewScreensProvider(&fynedesk.Screen{Name: "Screen0", X: 0, Y: 0,
		Width: 2000, Height: 1000, Scale: 1.0}), settings: wmTest.NewSettings()}
	l.bar = testBar([]string{})
	l.widgets = newWidgetPanel(l)
	bg := &background{wallpaper: canvas.NewImageFromResource(theme.FyneLogo())}
	deskSize := fyne.NewSize(2000, 1000)

	l.Layout([]fyne.CanvasObject{bg, l.bar, l.widgets}, deskSize)

	assert.Equal(t, deskSize, bg.Size())
	assert.Equal(t, deskSize.Width, l.widgets.Position().X+l.widgets.Size().Width)
	assert.Equal(t, deskSize.Height, l.widgets.Size().Height)
	if l.Settings().NarrowLeftLauncher() {
		assert.Equal(t, deskSize.Width, wmTheme.NarrowBarWidth)
		assert.Equal(t, deskSize.Height, l.bar.Size().Height)
	} else {
		assert.Equal(t, deskSize.Width, l.bar.Size().Width)
		assert.Equal(t, deskSize.Height, l.bar.Position().Y+l.bar.Size().Height-1) // -1 rounding fix, desk.go:49
	}
}

func TestScaleVars_Up(t *testing.T) {
	l := &desktop{}
	l.screens = wmTest.NewScreensProvider()
	l.screens.Screens()[0].Scale = 1.8
	env := l.scaleVars(1.8)
	assert.Contains(t, env, "QT_SCREEN_SCALE_FACTORS=Screen0=1.8")
	assert.Contains(t, env, "GDK_SCALE=2")
	assert.Contains(t, env, "ELM_SCALE=1.8")
}

func TestScaleVars_Down(t *testing.T) {
	l := &desktop{}
	l.screens = wmTest.NewScreensProvider()
	l.screens.Screens()[0].Scale = 0.9
	env := l.scaleVars(0.9)
	assert.Contains(t, env, "QT_SCREEN_SCALE_FACTORS=Screen0=1.0")
	assert.Contains(t, env, "GDK_SCALE=1")
	assert.Contains(t, env, "ELM_SCALE=0.9")
}

func TestBackgroundChange(t *testing.T) {
	l := &desktop{screens: wmTest.NewScreensProvider(&fynedesk.Screen{Name: "Screen0", X: 0, Y: 0,
		Width: 2000, Height: 1000, Scale: 1.0})}
	fynedesk.SetInstance(l)
	l.app = test.NewApp()
	l.settings = wmTest.NewSettings()
	l.setupRoot()

	bg := l.root.Content().(*fyne.Container).Objects[0].(*background)

	workingDir, err := os.Getwd()
	if err != nil {
		fyne.LogError("Could not get current working directory", err)
		t.FailNow()
	}
	assert.Equal(t, wmTheme.Background, bg.wallpaper.Resource)

	l.settings.(*wmTest.Settings).SetBackground(filepath.Join(workingDir, "testdata", "fyne.png"))
	l.updateBackgrounds(l.Settings().Background())
	assert.Equal(t, l.settings.Background(), bg.wallpaper.File)
}
