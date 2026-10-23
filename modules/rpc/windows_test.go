package rpc

import (
	"testing"

	"github.com/FyshOS/saver"

	"fyne.io/fyne/v2"

	"fyshos.com/tyde"
	wmTest "fyshos.com/tyde/test"

	"github.com/stretchr/testify/assert"
)

// stackWM is a window list, topmost first.
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

// setupWindows runs the desktop with the given windows open, first on top.
func setupWindows(t *testing.T, wins ...tyde.Window) *wmTest.Desktop {
	wm := &stackWM{}
	for i := len(wins) - 1; i >= 0; i-- {
		wm.AddWindow(wins[i])
	}
	desk := wmTest.NewDesktopWithWM(wm)
	t.Cleanup(func() { tyde.SetInstance(nil) })
	return desk
}

func TestWindowsCommand(t *testing.T) {
	top := wmTest.NewWindow("Terminal")
	top.Focus()
	top.Maximize()
	away := wmTest.NewWindow("Firefox")
	away.SetDesktop(1)
	away.Iconify()
	pinned := wmTest.NewWindow("Music")
	pinned.Pin()
	pinned.Fullscreen()
	desk := setupWindows(t, top, away, pinned)
	desk.Settings().(*wmTest.Settings).SetDesktopNames([]string{"", "Work"})

	reply, err := Run("windows")
	assert.NoError(t, err)
	assert.Equal(t, "Terminal: desktop 1, maximized, focused\n"+
		"Firefox: desktop Work, minimized\n"+
		"Music: every desktop, fullscreen", reply)
}

func TestWindowsCommand_None(t *testing.T) {
	setupWindows(t)
	reply, err := Run("windows")
	assert.NoError(t, err)
	assert.Equal(t, "No windows are open", reply)
}

func TestFocusCommand(t *testing.T) {
	top, away := wmTest.NewWindow("Terminal"), wmTest.NewWindow("Firefox")
	away.SetDesktop(2)
	away.Iconify()
	desk := setupWindows(t, top, away)

	reply, err := Run("focus firefox")
	assert.NoError(t, err)
	assert.Equal(t, "Focused Firefox", reply)
	assert.Equal(t, 2, desk.Desktop(), "follows the window to its desktop")
	assert.False(t, away.Iconic())
	assert.True(t, away.TopWindow())
	assert.True(t, away.Focused())
	assert.False(t, top.Focused())
}

func TestFocusCommand_Refuses(t *testing.T) {
	setupWindows(t, wmTest.NewWindow("Terminal"))

	_, err := Run("focus")
	assert.EqualError(t, err, "focus: needs the title of the window to focus")
	_, err = Run("focus Firefox")
	assert.EqualError(t, err, `focus: no window matching "Firefox"`)
}

// Commands given no window act on the one in front.
func TestWindowCommands_FocusedByDefault(t *testing.T) {
	top, under := wmTest.NewWindow("Terminal"), wmTest.NewWindow("Firefox")
	setupWindows(t, top, under)

	reply, err := Run("minimize")
	assert.NoError(t, err)
	assert.Equal(t, "Minimized Terminal", reply)
	assert.True(t, top.Iconic())
	assert.False(t, under.Iconic())

	reply, err = Run("pin")
	assert.NoError(t, err)
	assert.Equal(t, "Pinned Terminal", reply)
	assert.True(t, top.Pinned())

	reply, err = Run("unpin")
	assert.NoError(t, err)
	assert.Equal(t, "Unpinned Terminal", reply)
	assert.False(t, top.Pinned())
}

func TestWindowCommands_Named(t *testing.T) {
	top, under := wmTest.NewWindow("Terminal"), wmTest.NewWindow("Firefox")
	under.Iconify()
	setupWindows(t, top, under)

	reply, err := Run("maximize firefox")
	assert.NoError(t, err)
	assert.Equal(t, "Maximized Firefox", reply)
	assert.True(t, under.Maximized())
	assert.False(t, under.Iconic(), "a minimized window is brought back to be maximized")
	assert.False(t, top.Maximized())

	reply, err = Run("fullscreen firefox")
	assert.NoError(t, err)
	assert.Equal(t, "Fullscreened Firefox", reply)
	assert.True(t, under.Fullscreened())

	reply, err = Run("restore firefox")
	assert.NoError(t, err)
	assert.Equal(t, "Restored Firefox", reply)
	assert.False(t, under.Fullscreened())
	assert.False(t, under.Maximized())

	reply, err = Run("close firefox")
	assert.NoError(t, err)
	assert.Equal(t, "Closed Firefox", reply)
}

func TestWindowCommands_NoWindows(t *testing.T) {
	setupWindows(t)
	_, err := Run("close")
	assert.EqualError(t, err, "close: no windows are open")
}
