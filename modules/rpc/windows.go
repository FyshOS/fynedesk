package rpc

import (
	"errors"
	"strings"

	"fyshos.com/tyde"
	"fyshos.com/tyde/internal/modhelper"
)

// windowNameArg describes the argument every window command takes.
const windowNameArg = "the window's title or part of it; the focused window when none is given"

// windowCommands are the desktop's own commands for the open windows. Each
// names its window by title, or acts on the focused window when given none.
var windowCommands = []tyde.Command{
	{
		Name:        "windows",
		Description: "List the open windows, front to back, with their desktop and state",
		Run:         listWindows,
	},
	{
		Name:        "focus",
		Description: "Bring a window to the front and give it keyboard focus, switching to its desktop if needed",
		Args:        "the window's title or part of it",
		Examples:    []string{"focus Firefox"},
		Run:         focusWindow,
	},
	{
		Name:        "close",
		Description: "Close a window",
		Args:        windowNameArg,
		Examples:    []string{"close", "close Firefox"},
		Run: onWindow("Closed", func(win tyde.Window) {
			win.Close()
		}),
	},
	{
		Name:        "minimize",
		Description: "Minimize a window to the taskbar",
		Args:        windowNameArg,
		Examples:    []string{"minimize", "minimize Terminal"},
		Run: onWindow("Minimized", func(win tyde.Window) {
			win.Iconify()
		}),
	},
	{
		Name:        "maximize",
		Description: "Make a window fill the screen, apart from the desktop bar and panel",
		Args:        windowNameArg,
		Examples:    []string{"maximize", "maximize Firefox"},
		Run: onWindow("Maximized", func(win tyde.Window) {
			if win.Iconic() {
				win.Uniconify()
			}
			win.Maximize()
		}),
	},
	{
		Name:        "fullscreen",
		Description: "Make a window cover the whole screen",
		Args:        windowNameArg,
		Examples:    []string{"fullscreen", "fullscreen Video Player"},
		Run: onWindow("Fullscreened", func(win tyde.Window) {
			if win.Iconic() {
				win.Uniconify()
			}
			win.Fullscreen()
		}),
	},
	{
		Name:        "restore",
		Description: "Return a window to its normal size and place, undoing minimize, maximize and fullscreen",
		Args:        windowNameArg,
		Examples:    []string{"restore", "restore Firefox"},
		Run: onWindow("Restored", func(win tyde.Window) {
			if win.Iconic() {
				win.Uniconify()
			}
			if win.Fullscreened() {
				win.Unfullscreen()
			}
			if win.Maximized() {
				win.Unmaximize()
			}
		}),
	},
	{
		Name:        "pin",
		Description: "Keep a window visible on every virtual desktop",
		Args:        windowNameArg,
		Examples:    []string{"pin", "pin Music"},
		Run: onWindow("Pinned", func(win tyde.Window) {
			win.Pin()
		}),
	},
	{
		Name:        "unpin",
		Description: "Stop a window following you between virtual desktops",
		Args:        windowNameArg,
		Examples:    []string{"unpin", "unpin Music"},
		Run: onWindow("Unpinned", func(win tyde.Window) {
			win.Unpin()
		}),
	},
}

// onWindow builds a command action that finds the named window, performs act on
// it and reports what was done to which window.
func onWindow(did string, act func(tyde.Window)) func(string) (string, error) {
	return func(args string) (string, error) {
		win, err := modhelper.FindWindow(args)
		if err != nil {
			return "", err
		}
		act(win)
		return did + " " + modhelper.WindowTitle(win), nil
	}
}

func focusWindow(args string) (string, error) {
	if strings.TrimSpace(args) == "" {
		return "", errors.New("needs the title of the window to focus")
	}
	win, err := modhelper.FindWindow(args)
	if err != nil {
		return "", err
	}

	desk := tyde.Instance()
	if !win.Pinned() && win.Desktop() != desk.Desktop() {
		desk.SetDesktop(win.Desktop())
	}
	if win.Iconic() {
		win.Uniconify()
	}
	win.RaiseToTop()
	win.Focus()
	return "Focused " + modhelper.WindowTitle(win), nil
}

func listWindows(string) (string, error) {
	desk := tyde.Instance()
	if desk.WindowManager() == nil {
		return "", errors.New("no window manager running")
	}

	var lines []string
	for _, win := range desk.WindowManager().Windows() {
		if win.Properties().SkipTaskbar() {
			continue
		}
		lines = append(lines, describeWindow(desk, win))
	}
	if len(lines) == 0 {
		return "No windows are open", nil
	}
	return strings.Join(lines, "\n"), nil
}

// describeWindow writes one line of the windows listing: the title, then where
// and how the window is, e.g. "Firefox: desktop Work, maximized, focused".
func describeWindow(desk tyde.Desktop, win tyde.Window) string {
	details := []string{"desktop " + tyde.DesktopName(desk.Settings(), win.Desktop())}
	if win.Pinned() {
		details[0] = "every desktop"
	}
	if win.Iconic() {
		details = append(details, "minimized")
	}
	if win.Fullscreened() {
		details = append(details, "fullscreen")
	} else if win.Maximized() {
		details = append(details, "maximized")
	}
	if win.Focused() {
		details = append(details, "focused")
	}
	return modhelper.WindowTitle(win) + ": " + strings.Join(details, ", ")
}
