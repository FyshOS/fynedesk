package rpc

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"

	wmTest "fyshos.com/tyde/test"

	"github.com/stretchr/testify/assert"
)

// The test desktop offers a 320x240 content area at a scale of 1, so pixels
// and units agree and a half is 160 wide.

func TestMoveCommand_Places(t *testing.T) {
	test.NewApp()
	for _, tt := range []struct {
		args  string
		x, y  int
		w, h  uint
		reply string
	}{
		{"left", 0, 0, 160, 240, "Moved Terminal to the left half"},
		{"to the right", 160, 0, 160, 240, "Moved Terminal to the right half"},
		{"top", 0, 0, 320, 120, "Moved Terminal to the top half"},
		{"bottom", 0, 120, 320, 120, "Moved Terminal to the bottom half"},
		{"top left", 0, 0, 160, 120, "Moved Terminal to the top left quarter"},
		{"Top-Right", 160, 0, 160, 120, "Moved Terminal to the top right quarter"},
		{"bottomleft", 0, 120, 160, 120, "Moved Terminal to the bottom left quarter"},
		{"bottom right", 160, 120, 160, 120, "Moved Terminal to the bottom right quarter"},
		{"100 200", 100, 200, 100, 50, "Moved Terminal to 100,200"},
	} {
		win := wmTest.NewWindow("Terminal")
		win.SetGeometry(5, 5, 100, 50)
		setupWindows(t, win)

		reply, err := Run("move " + tt.args)
		assert.NoError(t, err, tt.args)
		assert.Equal(t, tt.reply, reply, tt.args)
		assert.Equal(t, fyne.NewPos(float32(tt.x), float32(tt.y)), win.Position(), tt.args)
		assert.Equal(t, fyne.NewSize(float32(tt.w), float32(tt.h)), win.Size(), tt.args)
	}
}

func TestMoveCommand_Named(t *testing.T) {
	test.NewApp()
	top, under := wmTest.NewWindow("Terminal"), wmTest.NewWindow("Mozilla Firefox")
	setupWindows(t, top, under)

	reply, err := Run("move Firefox to top right")
	assert.NoError(t, err)
	assert.Equal(t, "Moved Mozilla Firefox to the top right quarter", reply)
	assert.Equal(t, fyne.NewPos(160, 0), under.Position())
	assert.Equal(t, fyne.NewPos(0, 0), top.Position())
	assert.Equal(t, fyne.NewSize(10, 10), top.Size(), "the window in front is left alone")
}

// A window on another desktop sits a root height away from the visible screen,
// so it is placed there rather than dragged onto the current desktop.
func TestMoveCommand_OtherDesktop(t *testing.T) {
	test.NewApp()
	win := wmTest.NewWindow("Terminal")
	win.SetDesktop(2)
	pinned := wmTest.NewWindow("Music")
	pinned.SetDesktop(2)
	pinned.Pin()
	setupWindows(t, win, pinned)

	_, err := Run("move left")
	assert.NoError(t, err)
	assert.Equal(t, fyne.NewPos(0, 480), win.Position())

	_, err = Run("move Music 10 20")
	assert.NoError(t, err)
	assert.Equal(t, fyne.NewPos(10, 20), pinned.Position(), "pinned windows are on every desktop")
}

func TestMoveCommand_Refuses(t *testing.T) {
	test.NewApp()
	win := wmTest.NewWindow("Terminal")
	setupWindows(t, win)

	_, err := Run("move")
	assert.EqualError(t, err, "move: needs somewhere to move the window: "+placeNames+", or an x and y")
	_, err = Run("move Terminal sideways")
	assert.EqualError(t, err, `move: "sideways" is not a place a window can go: `+placeNames)
	_, err = Run("move Firefox left")
	assert.EqualError(t, err, `move: no window matching "Firefox"`)

	win.Maximize()
	_, err = Run("move left")
	assert.EqualError(t, err, "move: Terminal is maximized, restore it first")
	win.Unmaximize()
	win.Fullscreen()
	_, err = Run("move left")
	assert.EqualError(t, err, "move: Terminal is fullscreen, restore it first")
	assert.Equal(t, fyne.NewSize(10, 10), win.Size())
}

func TestResizeCommand(t *testing.T) {
	test.NewApp()
	top, under := wmTest.NewWindow("Terminal"), wmTest.NewWindow("Firefox")
	top.SetGeometry(30, 40, 10, 10)
	setupWindows(t, top, under)

	reply, err := Run("resize 1280 720")
	assert.NoError(t, err)
	assert.Equal(t, "Resized Terminal to 1280x720", reply)
	assert.Equal(t, fyne.NewSize(1280, 720), top.Size())
	assert.Equal(t, fyne.NewPos(30, 40), top.Position(), "the top left stays put")

	reply, err = Run("resize Firefox to 640X480")
	assert.NoError(t, err)
	assert.Equal(t, "Resized Firefox to 640x480", reply)
	assert.Equal(t, fyne.NewSize(640, 480), under.Size())
}

func TestResizeCommand_Refuses(t *testing.T) {
	test.NewApp()
	win := wmTest.NewWindow("Terminal")
	setupWindows(t, win)

	_, err := Run("resize")
	assert.EqualError(t, err, "resize: needs a width and height in pixels")
	_, err = Run("resize big")
	assert.EqualError(t, err, "resize: needs a width and height in pixels, e.g. '1280 720'")
	_, err = Run("resize Terminal 800")
	assert.EqualError(t, err, "resize: needs a width and height in pixels, e.g. '1280 720'")
	_, err = Run("resize 0 100")
	assert.EqualError(t, err, "resize: 0x100 is not a size a window can be")

	win.Maximize()
	_, err = Run("resize 800 600")
	assert.EqualError(t, err, "resize: Terminal is maximized, restore it first")
	assert.Equal(t, fyne.NewSize(10, 10), win.Size())
}
