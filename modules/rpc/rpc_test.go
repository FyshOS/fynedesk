package rpc

import (
	"net/rpc"
	"os"
	"strconv"
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyshos.com/tyde/internal/modhelper"

	"fyshos.com/tyde"
	wmTest "fyshos.com/tyde/test"

	"github.com/stretchr/testify/assert"
)

func TestSocketPath_XDGRuntime(t *testing.T) {
	t.Setenv("XDG_RUNTIME_DIR", "/run/user/1000")
	assert.Equal(t, "/run/user/1000/tyde.sock", SocketPath())
}

func TestSocketPath_Fallback(t *testing.T) {
	t.Setenv("XDG_RUNTIME_DIR", "")
	expected := "/tmp/tyde-" + strconv.Itoa(os.Getuid()) + ".sock"
	assert.Equal(t, expected, SocketPath())
}

func TestRPCModule_Metadata(t *testing.T) {
	m := &rpcModule{}
	assert.Equal(t, "Remote Control", m.Metadata().Name)
}

func TestRPCModule_Destroy_Empty(t *testing.T) {
	m := &rpcModule{}
	m.Destroy() // listener nil, should not panic
}

func TestLaunchInput_NoDesktop(t *testing.T) {
	tyde.SetInstance(nil)
	_, err := LaunchInput("anything")
	assert.EqualError(t, err, "desktop not running")
}

func TestService_Launch_NoDesktop(t *testing.T) {
	tyde.SetInstance(nil)
	var reply string
	err := (&Service{}).Launch("welcome", &reply)
	assert.EqualError(t, err, "desktop not running")
}

// fakeSuggestModule is a LaunchSuggestionModule used to test the rpc service.
type fakeSuggestModule struct {
	name        string
	suggestions []tyde.LaunchSuggestion
}

func (f *fakeSuggestModule) Destroy() {}

func (f *fakeSuggestModule) Metadata() tyde.ModuleMetadata {
	return tyde.ModuleMetadata{Name: f.name}
}

func (f *fakeSuggestModule) LaunchSuggestions(_ string) []tyde.LaunchSuggestion {
	return f.suggestions
}

func newDeskWithModules(mods []tyde.Module) *wmTest.Desktop {
	d := wmTest.NewDesktop()
	d.SetModules(mods)
	return d
}

// fakeCommandModule offers a command the way a real module does: the launcher
// entry and the command are the same item.
type fakeCommandModule struct {
	fakeSuggestModule

	ran string
}

func (f *fakeCommandModule) Commands() []tyde.Command {
	c := tyde.Command{
		Name:        "brightness",
		Description: "Set the screen brightness",
		Args:        "a percentage",
	}

	modhelper.AnswerWith(&c, f.item)
	return []tyde.Command{c}
}

// item is the seam a real module shares between its launcher entry and its
// command - one function, called two ways.
func (f *fakeCommandModule) item(args string) tyde.LaunchSuggestion {
	if args == "" {
		return nil
	}
	f.ran = args
	return &fakeSuggestion{title: "Brightness " + args + "%"}
}

// fakeSilentModule describes a command it cannot perform, which is a module bug
// rather than something a caller should have to defend against.
type fakeSilentModule struct {
	fakeSuggestModule
}

func (f *fakeSilentModule) Commands() []tyde.Command {
	return []tyde.Command{{Name: "ghost", Description: "Does nothing at all"}}
}

func TestCommands_Builtins(t *testing.T) {
	tyde.SetInstance(nil)

	var names []string
	for _, c := range Commands() {
		names = append(names, c.Name)
	}
	assert.Equal(t, []string{
		"welcome", "fathom", "restart",
		"windows", "focus", "close", "minimize", "maximize", "fullscreen", "restore", "pin", "unpin",
	}, names)
}

func TestCommands_FromModule(t *testing.T) {
	mods := []tyde.Module{
		&fakeSuggestModule{name: "Launcher: Calculate"}, // offers no commands
		&fakeCommandModule{fakeSuggestModule: fakeSuggestModule{name: "Brightness"}},
	}
	tyde.SetInstance(newDeskWithModules(mods))
	t.Cleanup(func() { tyde.SetInstance(nil) })

	var names []string
	for _, c := range Commands() {
		names = append(names, c.Name)
	}
	assert.Contains(t, names, "brightness")
	assert.Contains(t, names, "welcome") // the desktop's own are still there
}

// A command with nothing behind it is dropped where the commands are gathered,
// so no caller can be handed one it cannot run.
func TestCommands_SkipsCommandWithNoAction(t *testing.T) {
	tyde.SetInstance(newDeskWithModules([]tyde.Module{
		&fakeSilentModule{fakeSuggestModule{name: "Ghost"}},
	}))
	t.Cleanup(func() { tyde.SetInstance(nil) })

	for _, c := range Commands() {
		assert.NotEqual(t, "ghost", c.Name)
	}
}

func TestService_ListCommands_NoDesktop(t *testing.T) {
	tyde.SetInstance(nil)
	var reply []tyde.Command
	err := (&Service{}).ListCommands("", &reply)
	assert.EqualError(t, err, "desktop not running")
}

// A named command answers, not the launch suggestions - even when a module
// would have matched the same words.
func TestRun_Command(t *testing.T) {
	test.NewApp()
	mod := &fakeCommandModule{fakeSuggestModule: fakeSuggestModule{
		name:        "Brightness",
		suggestions: []tyde.LaunchSuggestion{&fakeSuggestion{title: "not this"}},
	}}
	tyde.SetInstance(newDeskWithModules([]tyde.Module{mod}))
	t.Cleanup(func() { tyde.SetInstance(nil) })

	reply, err := Run("brightness 50")
	assert.NoError(t, err)
	assert.Equal(t, "Brightness 50%", reply)
	assert.Equal(t, "50", mod.ran)
}

// What a command refuses is reported under its name, so the reason reads as an
// answer to what was asked.
func TestRun_CommandRefuses(t *testing.T) {
	test.NewApp()
	tyde.SetInstance(newDeskWithModules([]tyde.Module{
		&fakeCommandModule{fakeSuggestModule: fakeSuggestModule{name: "Brightness"}},
	}))
	t.Cleanup(func() { tyde.SetInstance(nil) })

	_, err := Run("brightness")
	assert.EqualError(t, err, "brightness: needs something to act on")
}

// The desktop's own commands work the same way.
func TestRun_Builtin(t *testing.T) {
	test.NewApp()
	tyde.SetInstance(newDeskWithModules(nil))
	t.Cleanup(func() { tyde.SetInstance(nil) })

	// The test desktop has no welcome window, so the command says so.
	_, err := Run("welcome")
	assert.EqualError(t, err, "welcome: this desktop has no welcome window")
}

// TestSocket_ListCommands runs the discovery call the way tyde_ctl and any
// other client does - over the socket - so the command descriptions are known
// to survive the wire.
func TestSocket_ListCommands(t *testing.T) {
	t.Setenv("XDG_RUNTIME_DIR", t.TempDir()) // never touch the live session's socket
	tyde.SetInstance(newDeskWithModules([]tyde.Module{
		&fakeCommandModule{fakeSuggestModule: fakeSuggestModule{name: "Brightness"}},
	}))
	t.Cleanup(func() { tyde.SetInstance(nil) })

	mod := newRPC()
	t.Cleanup(mod.Destroy)

	client, err := rpc.Dial("unix", SocketPath())
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	var reply []tyde.Command
	assert.NoError(t, client.Call("Service.ListCommands", "", &reply))

	// The description arrives; the action does not travel, which is why a client
	// asks the desktop to run a command by name rather than running one itself.
	assert.Contains(t, reply, tyde.Command{
		Name:        "brightness",
		Description: "Set the screen brightness",
		Args:        "a percentage",
	})
}

func TestLaunchInput_NoMatch(t *testing.T) {
	mods := []tyde.Module{
		&fakeSuggestModule{name: "Launcher: Calculate"},
	}
	tyde.SetInstance(newDeskWithModules(mods))
	t.Cleanup(func() { tyde.SetInstance(nil) })

	_, err := LaunchInput("xyz")
	if assert.Error(t, err) {
		assert.True(t, strings.HasPrefix(err.Error(), "no match for "))
	}
}

func TestLaunchInput_SkipsSearchModule(t *testing.T) {
	// "Search" module produces a suggestion but should be skipped (catch-all).
	mods := []tyde.Module{
		&fakeSuggestModule{
			name:        "Launcher: Search",
			suggestions: []tyde.LaunchSuggestion{&fakeSuggestion{title: "search-result"}},
		},
	}
	tyde.SetInstance(newDeskWithModules(mods))
	t.Cleanup(func() { tyde.SetInstance(nil) })

	_, err := LaunchInput("anything")
	assert.Error(t, err) // no non-search match
}

// A command from a module that is not loaded is not a command at all: there is
// nothing to sniff a module interface for, and nothing that could run it.
func TestRun_ModuleNotLoaded(t *testing.T) {
	tyde.SetInstance(newDeskWithModules([]tyde.Module{
		&fakeSuggestModule{name: "Launcher: Calculate"},
	}))
	t.Cleanup(func() { tyde.SetInstance(nil) })

	_, err := Run("brightness 50")
	assert.EqualError(t, err, `no match for "brightness 50"`)
}

type fakeSuggestion struct {
	title string
}

func (f *fakeSuggestion) Icon() fyne.Resource { return nil }
func (f *fakeSuggestion) Title() string       { return f.title }
func (f *fakeSuggestion) Launch()             {}

// fakeChatModule stands in for the AI assistant, which the desktop finds by the
// one thing it needs from it.
type fakeChatModule struct {
	shown bool
}

func (f *fakeChatModule) Destroy() {}
func (f *fakeChatModule) Metadata() tyde.ModuleMetadata {
	return tyde.ModuleMetadata{Name: "AI Assistant"}
}
func (f *fakeChatModule) ShowChat() { f.shown = true }

// "tyde_ctl fathom" is what the Fathom desktop entry runs, so it opens the
// assistant rather than asking it a question called "fathom".
func TestRun_Fathom(t *testing.T) {
	test.NewApp()
	chat := &fakeChatModule{}
	tyde.SetInstance(newDeskWithModules([]tyde.Module{
		&fakeSuggestModule{
			name:        "AI Assistant suggestions",
			suggestions: []tyde.LaunchSuggestion{&fakeSuggestion{title: "Ask AI: fathom"}},
		},
		chat,
	}))
	t.Cleanup(func() { tyde.SetInstance(nil) })

	reply, err := Run("fathom")
	assert.NoError(t, err)
	assert.Equal(t, "Fathom AI", reply)
	assert.True(t, chat.shown)
}

func TestRun_Fathom_NoAssistant(t *testing.T) {
	test.NewApp()
	tyde.SetInstance(newDeskWithModules(nil))
	t.Cleanup(func() { tyde.SetInstance(nil) })

	_, err := Run("fathom")
	assert.EqualError(t, err, "fathom: the AI assistant module is not enabled")
}
