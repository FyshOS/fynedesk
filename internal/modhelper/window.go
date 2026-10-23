package modhelper

import (
	"errors"
	"fmt"
	"strings"

	"fyshos.com/tyde"
)

// FindWindow returns the managed window a person means by name: the one whose
// title matches exactly, else the first whose title contains the words, else the
// first whose class or command does. An empty name means the current focused window.
func FindWindow(name string) (tyde.Window, error) {
	desk := tyde.Instance()
	if desk == nil || desk.WindowManager() == nil {
		return nil, errors.New("no window manager running")
	}

	var wins []tyde.Window
	for _, win := range desk.WindowManager().Windows() {
		if !win.Properties().SkipTaskbar() {
			wins = append(wins, win)
		}
	}
	if len(wins) == 0 {
		return nil, errors.New("no windows are open")
	}

	want := strings.ToLower(strings.TrimSpace(name))
	if want == "" {
		return wins[0], nil
	}

	for _, win := range wins {
		if strings.ToLower(win.Properties().Title()) == want {
			return win, nil
		}
	}
	for _, win := range wins {
		if strings.Contains(strings.ToLower(win.Properties().Title()), want) {
			return win, nil
		}
	}
	for _, win := range wins {
		props := win.Properties()
		for _, class := range props.Class() {
			if strings.ToLower(class) == want {
				return win, nil
			}
		}
		if strings.Contains(strings.ToLower(props.Command()), want) {
			return win, nil
		}
	}
	return nil, fmt.Errorf("no window matching %q", strings.TrimSpace(name))
}

// WindowTitle names a window in a reply, falling back to its class or command
// for the few that set no title.
func WindowTitle(win tyde.Window) string {
	props := win.Properties()
	if title := strings.TrimSpace(props.Title()); title != "" {
		return title
	}
	if class := props.Class(); len(class) > 0 && class[0] != "" {
		return class[0]
	}
	if cmd := props.Command(); cmd != "" {
		return cmd
	}
	return "untitled window"
}
