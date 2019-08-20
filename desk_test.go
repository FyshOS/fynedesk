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

	assert.Equal(t, l.background.Size(), deskSize)
	assert.Equal(t, l.widgets.Position().X+l.widgets.Size().Width, deskSize.Width)
	assert.Equal(t, l.widgets.Size().Height, deskSize.Height)
	assert.Equal(t, l.bar.Size().Width, deskSize.Width)
	assert.Equal(t, l.bar.Position().Y+l.bar.Size().Height, deskSize.Height)
}
