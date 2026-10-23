package desktops

import (
	"testing"

	"github.com/FyshOS/saver"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"

	"fyshos.com/tyde"
	wmTest "fyshos.com/tyde/test"

	"github.com/stretchr/testify/assert"
)

// stackWM is a window list, topmost first, that the pager can listen to.
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

// setup runs the module against a desktop with the given desktops and windows,
// returning its commands by name.
func setup(t *testing.T, names []string, count int, wins ...tyde.Window) (*wmTest.Desktop, map[string]tyde.Command) {
	test.NewApp()
	wm := &stackWM{}
	for i := len(wins) - 1; i >= 0; i-- {
		wm.AddWindow(wins[i])
	}
	desk := wmTest.NewDesktopWithWM(wm)
	t.Cleanup(func() { tyde.SetInstance(nil) })
	settings := desk.Settings().(*wmTest.Settings)
	settings.SetDesktopCount(count)
	settings.SetDesktopNames(names)

	mod := newDesktops().(*desktops)
	cmds := map[string]tyde.Command{}
	for _, c := range mod.Commands() {
		cmds[c.Name] = c
	}
	return desk, cmds
}

func TestDesktopCommand_Current(t *testing.T) {
	desk, cmds := setup(t, []string{"", "Work"}, 3)

	reply, err := cmds["desktop"].Run("")
	assert.NoError(t, err)
	assert.Equal(t, "Desktop 1 of 3", reply)

	desk.SetDesktop(1)
	reply, err = cmds["desktop"].Run("")
	assert.NoError(t, err)
	assert.Equal(t, "Desktop Work (2 of 3)", reply)
}

func TestDesktopCommand_Switch(t *testing.T) {
	desk, cmds := setup(t, []string{"Main", "Work"}, 4)

	for _, tt := range []struct {
		arg   string
		want  int
		reply string
	}{
		{"3", 2, "Desktop 3"},
		{"next", 3, "Desktop 4"},
		{"previous", 2, "Desktop 3"},
		{"work", 1, "Desktop Work"},
		{"prev", 0, "Desktop Main"},
	} {
		reply, err := cmds["desktop"].Run(tt.arg)
		assert.NoError(t, err, tt.arg)
		assert.Equal(t, tt.want, desk.Desktop(), tt.arg)
		assert.Equal(t, tt.reply, reply, tt.arg)
	}
}

func TestDesktopCommand_Refuses(t *testing.T) {
	desk, cmds := setup(t, []string{"Main"}, 2)

	_, err := cmds["desktop"].Run("previous")
	assert.EqualError(t, err, "already on the first desktop")
	_, err = cmds["desktop"].Run("3")
	assert.EqualError(t, err, "there are 2 desktops, no desktop 3")
	_, err = cmds["desktop"].Run("0")
	assert.EqualError(t, err, "there are 2 desktops, no desktop 0")
	_, err = cmds["desktop"].Run("Games")
	assert.EqualError(t, err, `no desktop called "Games"`)
	assert.Equal(t, 0, desk.Desktop())

	desk.SetDesktop(1)
	_, err = cmds["desktop"].Run("next")
	assert.EqualError(t, err, "already on the last desktop")
}

// A name is only good for a desktop that is still offered, even if a longer
// names list remains from when there were more.
func TestDesktopCommand_NameBeyondCount(t *testing.T) {
	_, cmds := setup(t, []string{"Main", "Work", "Old"}, 2)
	_, err := cmds["desktop"].Run("Old")
	assert.EqualError(t, err, `no desktop called "Old"`)
}

func TestSendCommand(t *testing.T) {
	top, other := wmTest.NewWindow("Terminal"), wmTest.NewWindow("Firefox")
	_, cmds := setup(t, []string{"Main", "Work"}, 3, top, other)

	reply, err := cmds["send"].Run("to 2")
	assert.NoError(t, err)
	assert.Equal(t, "Terminal to desktop Work", reply)
	assert.Equal(t, 1, top.Desktop())
	assert.Equal(t, 0, other.Desktop())

	reply, err = cmds["send"].Run("firefox to 3")
	assert.NoError(t, err)
	assert.Equal(t, "Firefox to desktop 3", reply)
	assert.Equal(t, 2, other.Desktop())

	reply, err = cmds["send"].Run("next") // relative to the current desktop, as the key bindings are
	assert.NoError(t, err)
	assert.Equal(t, "Terminal to desktop Work", reply)
}

func TestSendCommand_Refuses(t *testing.T) {
	_, cmds := setup(t, nil, 2, wmTest.NewWindow("Terminal"))

	_, err := cmds["send"].Run("")
	assert.EqualError(t, err, "needs a desktop to send the window to")
	_, err = cmds["send"].Run("Firefox to 2")
	assert.EqualError(t, err, `no window matching "Firefox"`)
	_, err = cmds["send"].Run("Terminal to 5")
	assert.EqualError(t, err, "there are 2 desktops, no desktop 5")
}

func TestSendCommand_NoWindows(t *testing.T) {
	_, cmds := setup(t, nil, 2)
	_, err := cmds["send"].Run("2")
	assert.EqualError(t, err, "no windows are open")
}

func TestOverviewCommand(t *testing.T) {
	_, cmds := setup(t, nil, 2)
	reply, err := cmds["overview"].Run("")
	assert.NoError(t, err)
	assert.Equal(t, "Desktop overview", reply)
}
