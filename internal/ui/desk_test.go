package ui

import (
	"os"
	"path/filepath"
	"testing"

	"fyne.io/fyne/theme"
	"github.com/stretchr/testify/assert"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"

	"fyne.io/fynedesk"
	"fyne.io/fynedesk/test"
	wmTheme "fyne.io/fynedesk/theme"
)

func TestDeskLayout_Layout(t *testing.T) {
	l := &deskLayout{screens: test.NewScreensProvider(&fynedesk.Screen{Name: "Screen0", X: 0, Y: 0,
		Width: 2000, Height: 1000, Scale: 1.0})}
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
	l.screens = test.NewScreensProvider()
	env := l.scaleVars(1.8)
	assert.Contains(t, env, "QT_SCALE_FACTOR=1.8")
	assert.Contains(t, env, "GDK_SCALE=2")
	assert.Contains(t, env, "ELM_SCALE=1.8")
}

func TestScaleVars_Down(t *testing.T) {
	l := &deskLayout{}
	l.screens = test.NewScreensProvider()
	env := l.scaleVars(0.9)
	assert.Contains(t, env, "QT_SCALE_FACTOR=1.0")
	assert.Contains(t, env, "GDK_SCALE=1")
	assert.Contains(t, env, "ELM_SCALE=0.9")
}

func TestBackgroundChange(t *testing.T) {
	l := &deskLayout{screens: test.NewScreensProvider(&fynedesk.Screen{Name: "Screen0", X: 0, Y: 0,
		Width: 2000, Height: 1000, Scale: 1.0})}
	fynedesk.SetInstance(l)
	l.settings = test.NewSettings()
	bg := newBackground()
	l.backgroundScreenMap = make(map[*background]*fynedesk.Screen)
	l.backgroundScreenMap[bg] = l.screens.Primary()

	workingDir, err := os.Getwd()
	if err != nil {
		fyne.LogError("Could not get current working directory", err)
		t.FailNow()
	}
	assert.Equal(t, wmTheme.Background, bg.wallpaper.Resource)

	l.settings.(*test.Settings).SetBackground(filepath.Join(workingDir, "testdata", "fyne.png"))
	l.updateBackgrounds(l.Settings().Background())
	assert.Equal(t, l.settings.Background(), bg.wallpaper.File)
}
