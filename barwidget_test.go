package desktop

import (
	"testing"

	"fyne.io/fyne"
	"github.com/stretchr/testify/assert"

	_ "fyne.io/fyne/test"
	"fyne.io/fyne/widget"
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
func (*dummyWindow) Iconic() bool {
	return false
}

func (*dummyWindow) Maximized() bool {
	return false
}

func (*dummyWindow) TopWindow() bool {
	return true
}

func (*dummyWindow) Focus() {
	// no-op
}

func (*dummyWindow) Close() {
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

func (d *dummyIcon) Run() error {
	// no-op
	return nil
}

func testBar(icons []string) *bar {
	appBar := newAppBar(&testDesk{})
	for range icons {
		icon := barCreateIcon(appBar, false, &dummyIcon{}, nil)
		if icon != nil {
			appBar.append(icon)
		}
	}

	return appBar
}

func TestAppBar_Append(t *testing.T) {
	icons := []string{"fyne", "fyne", "fyne", "fyne"}
	appBar := testBar(icons)
	assert.Equal(t, len(icons), len(appBar.children))
	appBar.appendSeparator()
	assert.Equal(t, len(icons)+1, len(appBar.children))
	win := &dummyWindow{}
	icon := barCreateIcon(appBar, true, &dummyIcon{}, win)
	appBar.append(icon)
	assert.Equal(t, len(icons)+2, len(appBar.children))
	appBar.removeFromTaskbar(icon)
	assert.Equal(t, len(icons)+1, len(appBar.children))
}

func TestAppBar_Zoom(t *testing.T) {
	icons := []string{"fyne", "fyne", "fyne", "fyne"}
	appBar := testBar(icons)
	appBar.mouseInside = true
	appBar.mousePosition = appBar.children[0].Position()
	widget.Refresh(appBar)
	zoomTest := false
	if appBar.children[0].Size().Width > appBar.children[1].Size().Width {
		zoomTest = true
	}
	assert.Equal(t, true, zoomTest)
}
