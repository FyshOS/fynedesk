package rpc

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"

	"fyshos.com/tyde"
	"fyshos.com/tyde/internal/modhelper"
)

// places are the names a window can be moved to, each a fraction of the
// screen's usable area: the left edge and top as fractions of the width and
// height, then the width and height as fractions.
var places = map[string][4]float32{
	"left":         {0, 0, 0.5, 1},
	"right":        {0.5, 0, 0.5, 1},
	"top":          {0, 0, 1, 0.5},
	"bottom":       {0, 0.5, 1, 0.5},
	"top left":     {0, 0, 0.5, 0.5},
	"top right":    {0.5, 0, 0.5, 0.5},
	"bottom left":  {0, 0.5, 0.5, 0.5},
	"bottom right": {0.5, 0.5, 0.5, 0.5},
}

const placeNames = "left, right, top, bottom or a corner such as 'top left'"

var geometryCommands = []tyde.Command{
	{
		Name:        "move",
		Description: "Move a window to a part of its screen, sized to fit a half or quarter, or to an exact position",
		Args:        "the window's title or part of it (the focused window when none), then " + placeNames + ", or an x and y in pixels from the screen's top left",
		Examples:    []string{"move left", "move Firefox to top right", "move Firefox 100 200"},
		Run:         moveWindow,
	},
	{
		Name:        "resize",
		Description: "Resize a window, keeping its top left corner where it is",
		Args:        "the window's title or part of it (the focused window when none), then the width and height in pixels",
		Examples:    []string{"resize 1280 720", "resize Firefox 1280x720"},
		Run:         resizeWindow,
	},
}

func moveWindow(args string) (string, error) {
	title, where, err := splitPlace(args)
	if err != nil {
		return "", err
	}
	win, err := placeableWindow(title)
	if err != nil {
		return "", err
	}

	screen, scale := screenOf(win)
	cx, cy, cw, ch := tyde.Instance().ContentBoundsPixels(screen)
	x, y := screen.X+int(cx), screen.Y+int(cy)+desktopOffset(win)

	switch where.kind {
	case placeNamed:
		frac := places[where.name]
		x += int(float32(cw) * frac[0])
		y += int(float32(ch) * frac[1])
		w, h := int(float32(cw)*frac[2]), int(float32(ch)*frac[3])
		win.Resize(fyne.NewSize(float32(w)/scale, float32(h)/scale))
	case placeAt:
		x, y = screen.X+where.x, screen.Y+where.y+desktopOffset(win)
	}
	win.Move(fyne.NewPos(float32(x)/scale, float32(y)/scale))

	return "Moved " + modhelper.WindowTitle(win) + " to " + where.String(), nil
}

func resizeWindow(args string) (string, error) {
	title, w, h, err := splitSize(args)
	if err != nil {
		return "", err
	}
	win, err := placeableWindow(title)
	if err != nil {
		return "", err
	}

	_, scale := screenOf(win)
	win.Resize(fyne.NewSize(float32(w)/scale, float32(h)/scale))
	return fmt.Sprintf("Resized %s to %dx%d", modhelper.WindowTitle(win), w, h), nil
}

// placeableWindow finds the named window, refusing one whose geometry the window
// manager owns.
func placeableWindow(title string) (tyde.Window, error) {
	win, err := modhelper.FindWindow(title)
	if err != nil {
		return nil, err
	}
	if win.Fullscreened() {
		return nil, errors.New(modhelper.WindowTitle(win) + " is fullscreen, restore it first")
	}
	if win.Maximized() {
		return nil, errors.New(modhelper.WindowTitle(win) + " is maximized, restore it first")
	}
	return win, nil
}

// screenOf returns the screen a window is on and the factor between the pixel
// sizes people give and the units a window is moved and sized in.
func screenOf(win tyde.Window) (*tyde.Screen, float32) {
	screens := tyde.Instance().Screens()
	screen := screens.ScreenForWindow(win)
	if screen == nil {
		screen = screens.Primary()
	}
	return screen, screen.CanvasScale()
}

// desktopOffset is how far below (or above) the visible screens a window sits
// because it belongs to another virtual desktop, in pixels.
func desktopOffset(win tyde.Window) int {
	desk := tyde.Instance()
	if win.Pinned() {
		return 0
	}
	_, rootHeight := desk.RootSizePixels()
	return (win.Desktop() - desk.Desktop()) * int(rootHeight)
}

type placeKind int

const (
	placeNamed placeKind = iota
	placeAt
)

// place is where a move command asked for: a named part of the screen or a position.
type place struct {
	kind placeKind
	name string
	x, y int
}

func (p place) String() string {
	switch p.kind {
	case placeAt:
		return fmt.Sprintf("%d,%d", p.x, p.y)
	}
	if strings.Contains(p.name, " ") {
		return "the " + p.name + " quarter"
	}
	return "the " + p.name + " half"
}

// splitPlace reads "[<window> [to]] <place>", where the place is the last one or
// two words or a pair of numbers, and whatever precedes it names the window.
func splitPlace(args string) (title string, where place, err error) {
	words := strings.Fields(args)
	n := len(words)
	if n == 0 {
		return "", where, errors.New("needs somewhere to move the window: " + placeNames + ", or an x and y")
	}

	if n >= 2 {
		x, errX := strconv.Atoi(words[n-2])
		y, errY := strconv.Atoi(words[n-1])
		if errX == nil && errY == nil {
			return windowBefore(words[:n-2]), place{kind: placeAt, x: x, y: y}, nil
		}
		if name := placeName(words[n-2] + " " + words[n-1]); name != "" {
			return windowBefore(words[:n-2]), place{name: name}, nil
		}
	}

	last := strings.ToLower(words[n-1])
	if name := placeName(last); name != "" {
		return windowBefore(words[:n-1]), place{name: name}, nil
	}
	return "", where, fmt.Errorf("%q is not a place a window can go: %s", words[n-1], placeNames)
}

// placeName normalises "Top-Left", "topleft" or "top left" to the places key.
func placeName(words string) string {
	key := strings.ToLower(strings.ReplaceAll(words, "-", " "))
	if _, ok := places[key]; ok {
		return key
	}
	for name := range places {
		if strings.ReplaceAll(name, " ", "") == key {
			return name
		}
	}
	return ""
}

// splitSize reads "[<window> [to]] <width> <height>" or "<width>x<height>".
func splitSize(args string) (title string, w, h int, err error) {
	words := strings.Fields(args)
	n := len(words)
	if n == 0 {
		return "", 0, 0, errors.New("needs a width and height in pixels")
	}

	var errW, errH error
	rest := words[:n-1]
	if wide, high, joined := strings.Cut(strings.ToLower(words[n-1]), "x"); joined {
		w, errW = strconv.Atoi(wide)
		h, errH = strconv.Atoi(high)
	} else if n >= 2 {
		w, errW = strconv.Atoi(words[n-2])
		h, errH = strconv.Atoi(words[n-1])
		rest = words[:n-2]
	} else {
		errW = errors.New("missing")
	}
	if errW != nil || errH != nil {
		return "", 0, 0, errors.New("needs a width and height in pixels, e.g. '1280 720'")
	}
	if w <= 0 || h <= 0 {
		return "", 0, 0, fmt.Errorf("%dx%d is not a size a window can be", w, h)
	}
	return windowBefore(rest), w, h, nil
}

// windowBefore joins the words naming a window, dropping a trailing "to" that
// reads naturally before a place ("move Firefox to the left").
func windowBefore(words []string) string {
	if n := len(words); n > 0 && strings.EqualFold(words[n-1], "the") {
		words = words[:n-1]
	}
	if n := len(words); n > 0 && strings.EqualFold(words[n-1], "to") {
		words = words[:n-1]
	}
	return strings.Join(words, " ")
}
