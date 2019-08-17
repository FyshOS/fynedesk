package desktop

import (
	"testing"

	_ "fyne.io/fyne/test"
	"fyne.io/fyne/widget"
	"github.com/magiconair/properties/assert"
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

func (*dummyWindow) Focus() {
	// no-op
}

func (*dummyWindow) Close() {
	// no-op
}

func (*dummyWindow) RaiseAbove(Window) {
	// no-op (this is instructing the window after stack changes)
}

type dummyIcon struct {
}

func (d *dummyIcon) Name() string {
	return "Fyne"
}

func (d *dummyIcon) IconName() string {
	return "fyne"
}

func (d *dummyIcon) IconPath() string {
	return "testdata/fyne.png"
}

func (d *dummyIcon) Exec() string {
	return ""
}

func TestAppBar(t *testing.T) {
	appBar := newAppBar(nil)
	icons := []string{"fyne", "fyne", "fyne", "fyne"}
	for range icons {
		icon := barCreateIcon(false, &dummyIcon{}, nil)
		if icon != nil {
			appBar.append(icon)
		}
	}
	assert.Equal(t, len(icons), len(appBar.children))
	appBar.appendSeparator()
	assert.Equal(t, len(icons)+1, len(appBar.children))
	win := &dummyWindow{}
	icon := barCreateIcon(true, &dummyIcon{}, win)
	appBar.append(icon)
	assert.Equal(t, len(icons)+2, len(appBar.children))
	appBar.removeFromTaskbar(icon)
	assert.Equal(t, len(icons)+1, len(appBar.children))
	appBar.mouseInside = true
	appBar.mousePosition = appBar.children[0].Position()
	widget.Refresh(appBar)
	zoomTest := false
	if appBar.children[0].Size().Width > appBar.children[1].Size().Width {
		zoomTest = true
	}
	assert.Equal(t, true, zoomTest)
}
