// +build linux

package wm

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"fyne.io/fynedesk"
	"fyne.io/fynedesk/test"
)

func TestStack_AddWindow(t *testing.T) {
	stack := &stack{}
	win := test.NewWindow("")

	stack.AddWindow(win)
	assert.Equal(t, 1, len(stack.Windows()))
}

func TestStack_RaiseToTop(t *testing.T) {
	fynedesk.SetInstance(test.NewDesktopWithWM(&x11WM{}))
	stack := &stack{}
	win1 := test.NewWindow("")
	win2 := test.NewWindow("")

	stack.AddWindow(win1)
	stack.AddWindow(win2)
	assert.Equal(t, win1, stack.TopWindow())

	// TODO -re-add when embedded wm is merged so we can use it above
	//stack.RaiseToTop(win2)
	//assert.Equal(t, win2, stack.TopWindow())
}

func TestStack_RemoveWindow(t *testing.T) {
	stack := &stack{}
	win := test.NewWindow("")

	stack.AddWindow(win)
	stack.RemoveWindow(win)
	assert.Equal(t, 0, len(stack.Windows()))

}
