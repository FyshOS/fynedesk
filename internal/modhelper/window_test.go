package modhelper

import (
	"testing"

	"github.com/FyshOS/saver"

	"fyne.io/fyne/v2"

	"fyshos.com/tyde"
	wmTest "fyshos.com/tyde/test"

	"github.com/stretchr/testify/assert"
)

// stackWM is the smallest window manager the finder needs: a list, topmost first.
type stackWM struct {
	windows []tyde.Window
}

func (s *stackWM) AddWindow(win tyde.Window) { s.windows = append([]tyde.Window{win}, s.windows...) }
func (s *stackWM) RaiseToTop(tyde.Window)    {}
func (s *stackWM) RemoveWindow(tyde.Window)  {}
func (s *stackWM) TopWindow() tyde.Window {
	if len(s.windows) == 0 {
		return nil
	}
	return s.windows[0]
}
func (s *stackWM) Windows() []tyde.Window                               { return s.windows }
func (s *stackWM) AddStackListener(tyde.StackListener)                  {}
func (s *stackWM) RemoveStackListener(tyde.StackListener)               {}
func (s *stackWM) Blank()                                               {}
func (s *stackWM) Close()                                               {}
func (s *stackWM) Run()                                                 {}
func (s *stackWM) ShowOverlay(fyne.Window, fyne.Size, fyne.Position)    {}
func (s *stackWM) ShowMenuOverlay(*fyne.Menu, fyne.Size, fyne.Position) {}
func (s *stackWM) ShowScreensaver(*saver.ScreenSaver)                   {}

func setupWindows(t *testing.T, wins ...tyde.Window) {
	wm := &stackWM{}
	for i := len(wins) - 1; i >= 0; i-- { // AddWindow puts each on top, so add bottom first
		wm.AddWindow(wins[i])
	}
	wmTest.NewDesktopWithWM(wm)
	t.Cleanup(func() { tyde.SetInstance(nil) })
}

func TestFindWindow_NoWindows(t *testing.T) {
	setupWindows(t)
	_, err := FindWindow("anything")
	assert.EqualError(t, err, "no windows are open")
}

func TestFindWindow_EmptyIsTop(t *testing.T) {
	top, under := wmTest.NewWindow("Editor"), wmTest.NewWindow("Browser")
	setupWindows(t, top, under)

	win, err := FindWindow("")
	assert.NoError(t, err)
	assert.Equal(t, top, win)
}

func TestFindWindow_ExactBeatsPartial(t *testing.T) {
	partial, exact := wmTest.NewWindow("Notes - Editor"), wmTest.NewWindow("editor")
	setupWindows(t, partial, exact)

	win, err := FindWindow("Editor")
	assert.NoError(t, err)
	assert.Equal(t, exact, win)
}

func TestFindWindow_PartialTitle(t *testing.T) {
	other, want := wmTest.NewWindow("Terminal"), wmTest.NewWindow("Home - Mozilla Firefox")
	setupWindows(t, other, want)

	win, err := FindWindow("firefox")
	assert.NoError(t, err)
	assert.Equal(t, want, win)
}

func TestFindWindow_ClassAndCommand(t *testing.T) {
	byClass := wmTest.NewWindow("Inbox (3)")
	byClass.SetClass([]string{"thunderbird", "Thunderbird"})
	byCmd := wmTest.NewWindow("~")
	byCmd.SetCommand("/usr/bin/xterm")
	setupWindows(t, byClass, byCmd)

	win, err := FindWindow("Thunderbird")
	assert.NoError(t, err)
	assert.Equal(t, byClass, win)

	win, err = FindWindow("xterm")
	assert.NoError(t, err)
	assert.Equal(t, byCmd, win)
}

func TestFindWindow_NoMatch(t *testing.T) {
	setupWindows(t, wmTest.NewWindow("Terminal"))
	_, err := FindWindow(" Firefox ")
	assert.EqualError(t, err, `no window matching "Firefox"`)
}

func TestWindowTitle_Fallbacks(t *testing.T) {
	win := wmTest.NewWindow("")
	assert.Equal(t, "untitled window", WindowTitle(win))
	win.SetCommand("xterm")
	assert.Equal(t, "xterm", WindowTitle(win))
	win.SetClass([]string{"XTerm"})
	assert.Equal(t, "XTerm", WindowTitle(win))
	assert.Equal(t, "Shell", WindowTitle(wmTest.NewWindow(" Shell ")))
}
