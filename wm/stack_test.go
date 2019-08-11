package wm

import (
	"github.com/fyne-io/desktop"
	"github.com/magiconair/properties/assert"
	"testing"
)

type dummyWindow struct {
}

func (*dummyWindow) Decorated() bool {
	return true
}

func (*dummyWindow) Title() string {
	return "Dummy"
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

func TestStack_AddWindow(t *testing.T) {
	stack := &stack{}
	win := &dummyWindow{}

	stack.AddWindow(win)
	assert.Equal(t, 1, len(stack.Windows()))
}

func TestStack_RemoveWindow(t *testing.T) {
	stack := &stack{}
	win := &dummyWindow{}

	stack.AddWindow(win)
	stack.RemoveWindow(win)
	assert.Equal(t, 0, len(stack.Windows()))

}

func TestStack_RaiseToTop(t *testing.T) {
	stack := &stack{}
	win1 := &dummyWindow{}
	win2 := &dummyWindow{}

	stack.AddWindow(win1)
	stack.AddWindow(win2)
	assert.Equal(t, win1, stack.TopWindow())

	stack.RaiseToTop(win2)
	assert.Equal(t, win2, stack.TopWindow())
}
