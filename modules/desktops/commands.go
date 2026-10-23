package desktops

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"fyshos.com/tyde"
	"fyshos.com/tyde/internal/modhelper"
)

// Commands offers the virtual desktop actions to tyde_ctl and the AI assistant.
func (d *desktops) Commands() []tyde.Command {
	return []tyde.Command{
		{
			Name:        "desktop",
			Description: "Switch to a virtual desktop, or with no argument say which is current",
			Args:        "a desktop number or name, or 'next' / 'previous'",
			Examples:    []string{"desktop 2", "desktop next", "desktop Work"},
			Run:         d.switchCommand,
		},
		{
			Name:        "send",
			Description: "Move a window to another virtual desktop, the focused window unless one is named",
			Args:        "a desktop number, name, 'next' or 'previous', optionally preceded by the window title and 'to'",
			Examples:    []string{"send next", "send to 2", "send Firefox to Work"},
			Run:         sendCommand,
		},
		{
			Name:        "overview",
			Description: "Show every virtual desktop at once so one can be picked",
			Run: func(string) (string, error) {
				tyde.Instance().ShowDesktopOverview(deskCount())
				return "Desktop overview", nil
			},
		},
	}
}

// switchCommand performs the "desktop" command.
func (d *desktops) switchCommand(args string) (string, error) {
	desk := tyde.Instance()
	if args == "" {
		current, count := desk.Desktop(), deskCount()
		if name := givenName(current); name != "" {
			return fmt.Sprintf("Desktop %s (%d of %d)", name, current+1, count), nil
		}
		return fmt.Sprintf("Desktop %d of %d", current+1, count), nil
	}

	id, err := parseDesktop(args)
	if err != nil {
		return "", err
	}
	d.setDesktop(id)
	return "Desktop " + tyde.DesktopName(desk.Settings(), id), nil
}

// sendCommand performs the "send" command: "[<window> to] <desktop>".
func sendCommand(args string) (string, error) {
	if args == "" {
		return "", errors.New("needs a desktop to send the window to")
	}

	title, dest, named := strings.Cut(args, " to ")
	if !named {
		title, dest = "", strings.TrimPrefix(args, "to ")
	}

	id, err := parseDesktop(dest)
	if err != nil {
		return "", err
	}
	win, err := modhelper.FindWindow(title)
	if err != nil {
		return "", err
	}

	win.SetDesktop(id)
	return modhelper.WindowTitle(win) + " to desktop " + tyde.DesktopName(tyde.Instance().Settings(), id), nil
}

// parseDesktop turns the way a person refers to a desktop - its number from 1,
// its name, or a step from the current one - into its index.
func parseDesktop(arg string) (int, error) {
	count := deskCount()
	current := tyde.Instance().Desktop()
	want := strings.ToLower(strings.TrimSpace(arg))

	switch want {
	case "":
		return 0, errors.New("needs a desktop number or name")
	case "next":
		if current >= count-1 {
			return 0, errors.New("already on the last desktop")
		}
		return current + 1, nil
	case "previous", "prev":
		if current <= 0 {
			return 0, errors.New("already on the first desktop")
		}
		return current - 1, nil
	}

	if n, err := strconv.Atoi(want); err == nil {
		if n < 1 || n > count {
			return 0, fmt.Errorf("there are %d desktops, no desktop %d", count, n)
		}
		return n - 1, nil
	}

	for i, name := range tyde.Instance().Settings().DesktopNames() {
		if i < count && strings.ToLower(strings.TrimSpace(name)) == want {
			return i, nil
		}
	}
	return 0, fmt.Errorf("no desktop called %q", strings.TrimSpace(arg))
}

// givenName is the name a person gave desktop id, or "" when it has only its
// number - the names list may be shorter than the desktop count.
func givenName(id int) string {
	names := tyde.Instance().Settings().DesktopNames()
	if id < 0 || id >= len(names) {
		return ""
	}
	return strings.TrimSpace(names[id])
}
