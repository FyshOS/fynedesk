//go:build linux || openbsd || freebsd || netbsd
// +build linux openbsd freebsd netbsd

package x11

import (
	"fyne.io/fyne/v2"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/icccm"
)

// WindowStateAction defines actions for manipulating window state
type WindowStateAction int

const (
	// WindowStateActionRemove removes a state
	WindowStateActionRemove WindowStateAction = 0
	// WindowStateActionAdd adds a state
	WindowStateActionAdd WindowStateAction = 1
	// WindowStateActionToggle inverts the state of an state
	WindowStateActionToggle WindowStateAction = 2
)

var (
	// AllowedActions is the list of actions the window manager allows
	AllowedActions = []string{
		"_NET_WM_ACTION_MOVE",
		"_NET_WM_ACTION_RESIZE",
		"_NET_WM_ACTION_MINIMIZE",
		"_NET_WM_ACTION_MAXIMIZE_HORZ",
		"_NET_WM_ACTION_MAXIMIZE_VERT",
		"_NET_WM_ACTION_CLOSE",
		"_NET_WM_ACTION_FULLSCREEN",
	}

	// SupportedHints is the complete list of hints that we support
	SupportedHints = append(AllowedActions, "_NET_ACTIVE_WINDOW",
		"_NET_CLIENT_LIST",
		"_NET_CLIENT_LIST_STACKING",
		"_NET_CURRENT_DESKTOP",
		"_NET_DESKTOP_GEOMETRY",
		"_NET_DESKTOP_VIEWPORT",
		"_NET_FRAME_EXTENTS",
		"_NET_MOVERESIZE_WINDOW",
		"_NET_NUMBER_OF_DESKTOPS",
		"_NET_WM_FULL_PLACEMENT",
		"_NET_WM_FULLSCREEN_MONITORS",
		"_NET_WM_MOVERESIZE",
		"_NET_WM_NAME",
		"_NET_WM_STATE",
		"_NET_WM_STATE_FULLSCREEN",
		"_NET_WM_STATE_HIDDEN",
		"_NET_WM_STATE_MAXIMIZED_HORZ",
		"_NET_WM_STATE_MAXIMIZED_VERT",
		"_NET_WM_STATE_SKIP_PAGER",
		"_NET_WM_STATE_SKIP_TASKBAR",
		"_NET_WORKAREA",
		"_NET_SUPPORTED",
	)
)

// WindowActiveGet returns the currently active window
func WindowActiveGet(x *xgbutil.XUtil) (xproto.Window, error) {
	return ewmh.ActiveWindowGet(x)
}

// WindowExtendedHintsAdd adds a hint to the window
func WindowExtendedHintsAdd(x *xgbutil.XUtil, win xproto.Window, hint string) {
	extendedHints, _ := ewmh.WmStateGet(x, win) // error unimportant
	extendedHints = append(extendedHints, hint)
	err := ewmh.WmStateSet(x, win, extendedHints)
	if err != nil {
		fyne.LogError("it was not possible to add the hint ", err)
	}
}

// WindowExtendedHintsGet returns a hint from the window
func WindowExtendedHintsGet(x *xgbutil.XUtil, win xproto.Window) []string {
	extendedHints, err := ewmh.WmStateGet(x, win)
	if err != nil {
		return nil
	}
	return extendedHints
}

// WindowExtendedHintsRemove removes a hint from the window
func WindowExtendedHintsRemove(x *xgbutil.XUtil, win xproto.Window, hint string) {
	extendedHints, err := ewmh.WmStateGet(x, win)
	if err != nil {
		return
	}
	for i, curHint := range extendedHints {
		if curHint == hint {
			extendedHints = append(extendedHints[:i], extendedHints[i+1:]...)
			e := ewmh.WmStateSet(x, win, extendedHints)
			if e != nil {
				fyne.LogError("it was not possible to add the hint ", e)
			}
			return
		}
	}
}

// WindowName gets the name of an X window
func WindowName(x *xgbutil.XUtil, win xproto.Window) string {
	//Spec says _NET_WM_NAME is preferred to WM_NAME
	name, err := ewmh.WmNameGet(x, win)
	if err != nil {
		name, err = icccm.WmNameGet(x, win)
		if err != nil {
			return ""
		}
	}

	return name
}

// WindowTransientForGet returns the window ID that the requested window is transient for.
// A return of 0 means that it is not transient
func WindowTransientForGet(x *xgbutil.XUtil, win xproto.Window) xproto.Window {
	transient, err := icccm.WmTransientForGet(x, win)
	if err == nil {
		return transient
	}
	hints, err := icccm.WmHintsGet(x, win)
	if err == nil {
		if hints.Flags&icccm.HintWindowGroup > 0 {
			return hints.WindowGroup
		}
	}
	return 0
}
