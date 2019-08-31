package mousebind

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xevent"
)

var modifiers []uint16 = []uint16{ // order matters!
	xproto.ModMaskShift, xproto.ModMaskLock, xproto.ModMaskControl,
	xproto.ModMask1, xproto.ModMask2, xproto.ModMask3,
	xproto.ModMask4, xproto.ModMask5,
	xproto.ButtonMask1, xproto.ButtonMask2, xproto.ButtonMask3,
	xproto.ButtonMask4, xproto.ButtonMask5,
	xproto.ButtonMaskAny,
}

var pointerMasks uint16 = xproto.EventMaskPointerMotion |
	xproto.EventMaskButtonRelease |
	xproto.EventMaskButtonPress

// Initialize attaches the appropriate callbacks to make mouse bindings easier.
// i.e., prep the dummy window to handle mouse dragging events
func Initialize(xu *xgbutil.XUtil) {
	xevent.MotionNotifyFun(dragStep).Connect(xu, xu.Dummy())
	xevent.ButtonReleaseFun(DragEnd).Connect(xu, xu.Dummy())
}

// ParseString takes a string of the format '[Mod[-Mod[...]]]-BUTTONNUMBER',
// i.e., 'Mod4-1', and returns a modifiers/button combination.
// "Mod" could also be one of {button1, button2, button3, button4, button5}.
// An error is returned if the string is malformed or if no BUTTONNUMBER
// could be found.
func ParseString(xu *xgbutil.XUtil, str string) (uint16, xproto.Button, error) {
	mods, button := uint16(0), xproto.Button(0)
	for _, part := range strings.Split(str, "-") {
		switch strings.ToLower(part) {
		case "shift":
			mods |= xproto.ModMaskShift
		case "lock":
			mods |= xproto.ModMaskLock
		case "control":
			mods |= xproto.ModMaskControl
		case "mod1":
			mods |= xproto.ModMask1
		case "mod2":
			mods |= xproto.ModMask2
		case "mod3":
			mods |= xproto.ModMask3
		case "mod4":
			mods |= xproto.ModMask4
		case "mod5":
			mods |= xproto.ModMask5
		case "button1":
			mods |= xproto.ButtonMask1
		case "button2":
			mods |= xproto.ButtonMask2
		case "button3":
			mods |= xproto.ButtonMask3
		case "button4":
			mods |= xproto.ButtonMask4
		case "button5":
			mods |= xproto.ButtonMask5
		case "any":
			mods |= xproto.ButtonMaskAny
		default: // a button!
			if button == 0 { // only accept the first button we see
				possible, err := strconv.ParseUint(part, 10, 8)
				if err == nil {
					button = xproto.Button(possible)
				} else {
					return 0, 0, fmt.Errorf("Could not convert '%s' to a "+
						"valid 8-bit integer.", part)
				}
			}
		}
	}

	if button == 0 {
		return 0, 0, fmt.Errorf("Could not find a valid button in the "+
			"string '%s'. Mouse binding failed.", str)
	}

	return mods, button, nil
}

// Grab grabs a button with mods on a particular window.
// Will also grab all combinations of modifiers found in xevent.IgnoreMods
// If 'sync' is True, then no further events can be processed until the
// grabbing client allows them to be. (Which is done via AllowEvents. Thus,
// if sync is True, you *must* make some call to AllowEvents at some
// point, or else your client will lock.)
func Grab(xu *xgbutil.XUtil, win xproto.Window, mods uint16,
	button xproto.Button, sync bool) {

	var pSync byte = xproto.GrabModeAsync
	if sync {
		pSync = xproto.GrabModeSync
	}

	for _, m := range xevent.IgnoreMods {
		xproto.GrabButton(xu.Conn(), true, win, pointerMasks, pSync,
			xproto.GrabModeAsync, 0, 0, byte(button), mods|m)
	}
}

// GrabChecked grabs a button with mods on a particular window. It does the
// same thing as Grab, but issues a checked request and returns an error
// on failure.
// Will also grab all combinations of modifiers found in xevent.IgnoreMods
// If 'sync' is True, then no further events can be processed until the
// grabbing client allows them to be. (Which is done via AllowEvents. Thus,
// if sync is True, you *must* make some call to AllowEvents at some
// point, or else your client will lock.)
func GrabChecked(xu *xgbutil.XUtil, win xproto.Window, mods uint16,
	button xproto.Button, sync bool) error {

	var pSync byte = xproto.GrabModeAsync
	if sync {
		pSync = xproto.GrabModeSync
	}

	var err error
	for _, m := range xevent.IgnoreMods {
		err = xproto.GrabButtonChecked(xu.Conn(), true, win, pointerMasks,
			pSync, xproto.GrabModeAsync, 0, 0, byte(button), mods|m).Check()
		if err != nil {
			return err
		}
	}
	return nil
}

// Ungrab undoes Grab. It will handle all combinations of modifiers found
// in xevent.IgnoreMods.
func Ungrab(xu *xgbutil.XUtil, win xproto.Window, mods uint16,
	button xproto.Button) {

	for _, m := range xevent.IgnoreMods {
		xproto.UngrabButtonChecked(xu.Conn(), byte(button), win, mods|m).Check()
	}
}

// GrabPointer grabs the entire pointer.
// Returns whether GrabStatus is successful and an error if one is reported by
// XGB. It is possible to not get an error and the grab to be unsuccessful.
// The purpose of 'win' is that after a grab is successful, ALL Button*Events
// will be sent to that window. Make sure you have a callback attached :-)
func GrabPointer(xu *xgbutil.XUtil, win xproto.Window, confine xproto.Window,
	cursor xproto.Cursor) (bool, error) {

	reply, err := xproto.GrabPointer(xu.Conn(), false, win, pointerMasks,
		xproto.GrabModeAsync, xproto.GrabModeAsync,
		confine, cursor, 0).Reply()
	if err != nil {
		return false, fmt.Errorf("GrabPointer: Error grabbing pointer on "+
			"window '%x': %s", win, err)
	}

	return reply.Status == xproto.GrabStatusSuccess, nil
}

// UngrabPointer undoes GrabPointer.
func UngrabPointer(xu *xgbutil.XUtil) {
	xproto.UngrabPointer(xu.Conn(), 0)
}
