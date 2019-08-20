package desktop

import (
	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"github.com/magiconair/properties/assert"
	"image/color"
	"testing"
)

type testDesk struct {
}

func (*testDesk) Root() fyne.Window {
	return nil
}

func (*testDesk) Run() {
}

func (*testDesk) Settings() DeskSettings {
	return &testSettings{}
}

func (*testDesk) IconProvider() ApplicationProvider {
	return nil
}

func (*testDesk) WindowManager() WindowManager {
	return nil
}

type testSettings struct {
}

func (*testSettings) IconTheme() string {
	return "testTheme"
}

func TestDeskLayout_Layout(t *testing.T) {
	l := &deskLayout{}
	l.background = canvas.NewRectangle(color.White)
	l.bar = canvas.NewRectangle(color.Black)
	l.widgets = canvas.NewRectangle(color.Black)
	deskSize := fyne.NewSize(2000, 1000)

	l.Layout([]fyne.CanvasObject{l.background, l.bar, l.widgets}, deskSize)

	assert.Equal(t, deskSize, l.background.Size())
	assert.Equal(t, deskSize.Width, l.widgets.Position().X+l.widgets.Size().Width)
	assert.Equal(t, deskSize.Height, l.widgets.Size().Height)
	assert.Equal(t, deskSize.Width, l.bar.Size().Width)
	assert.Equal(t, deskSize.Height, l.bar.Position().Y+l.bar.Size().Height)
}
