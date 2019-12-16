package desktop

import (
	"image/color"
	"testing"

	"github.com/stretchr/testify/assert"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
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
}

func (*testSettings) IconTheme() string {
	return "testTheme"
}

func (*testSettings) Background() string {
	return ""
}

func (*testSettings) DefaultApps() string {
	return ""
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

func newTestScreensProvider() ScreenList {
	return &testScreensProvider{}
}

func TestDeskLayout_Layout(t *testing.T) {
	l := &deskLayout{}
	instance = l
	l.screens = newTestScreensProvider()
	l.backgrounds = []fyne.CanvasObject{canvas.NewRectangle(color.White)}
	l.screenBackgroundMap = make(map[*Screen]fyne.CanvasObject)
	l.screenBackgroundMap[l.Screens().Screens()[0]] = l.backgrounds[0]
	l.bar = canvas.NewRectangle(color.Black)
	l.widgets = canvas.NewRectangle(color.Black)
	deskSize := fyne.NewSize(2000, 1000)

	l.Layout([]fyne.CanvasObject{l.backgrounds[0], l.bar, l.widgets}, deskSize)

	assert.Equal(t, l.backgrounds[0].Size(), deskSize)
	assert.Equal(t, l.widgets.Position().X+l.widgets.Size().Width, deskSize.Width)
	assert.Equal(t, l.widgets.Size().Height, deskSize.Height)
	assert.Equal(t, l.bar.Size().Width, deskSize.Width)
	assert.Equal(t, l.bar.Position().Y+l.bar.Size().Height, deskSize.Height)
}

func TestScaleVars(t *testing.T) {
	l := &deskLayout{}
	instance = l
	l.screens = newTestScreensProvider()
	env := l.scaleVars(1.8)
	assert.Contains(t, env, "QT_SCALE_FACTOR=1.8")
	assert.Contains(t, env, "GDK_SCALE=2")
	assert.Contains(t, env, "ELM_SCALE=1.8")
}
