package mousebind

import (
	"fmt"

	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xevent"
)

// connect is essentially 'Connect' for either ButtonPress or
// ButtonRelease events.
func connect(xu *xgbutil.XUtil, callback xgbutil.CallbackMouse, evtype int,
	win xproto.Window, buttonStr string, sync bool, grab bool) error {

	// Get the mods/button first
	mods, button, err := ParseString(xu, buttonStr)
	if err != nil {
		return err
	}

	// Only do the grab if we haven't yet on this window.
	// And if we WANT a grab...
	if grab && mouseBindGrabs(xu, evtype, win, mods, button) == 0 {
		err := GrabChecked(xu, win, mods, button, sync)
		if err != nil {
			// If a bad access, let's be nice and give a good error message.
			switch err.(type) {
			case xproto.AccessError:
				return fmt.Errorf("Got a bad access error when trying to bind "+
					"'%s'. This usually means another client has already "+
					"grabbed this mouse binding.", buttonStr)
			default:
				return fmt.Errorf("Could not bind '%s' because: %s",
					buttonStr, err)
			}
		}
	}

	// If we've never grabbed anything on this window before, we need to
	// make sure we can respond to it in the main event loop.
	var allCb xgbutil.Callback
	if evtype == xevent.ButtonPress {
		allCb = xevent.ButtonPressFun(runButtonPressCallbacks)
	} else {
		allCb = xevent.ButtonReleaseFun(runButtonReleaseCallbacks)
	}

	// If this is the first Button{Press|Release}Event on this window,
	// then we need to listen to Button{Press|Release} events in the main loop.
	if !connectedMouseBind(xu, evtype, win) {
		allCb.Connect(xu, win)
	}

	// Finally, attach the callback.
	attachMouseBindCallback(xu, evtype, win, mods, button, callback)

	return nil
}

// DeduceButtonInfo takes a (modifiers, button) tuple and returns the relevant
// modifiers that were activated. This accounts for modifiers in
// xevent.IgnoreMods and the the button mask of the button that is pressed.
func DeduceButtonInfo(state uint16,
	detail xproto.Button) (uint16, xproto.Button) {

	mods, button := state, detail
	for _, m := range xevent.IgnoreMods {
		mods &= ^m
	}

	// We also need to mask out the button that has been pressed/released,
	// since it is also typically a modifier.
	modsTemp := int32(mods)
	switch button {
	case 1:
		modsTemp &= ^xproto.ButtonMask1
	case 2:
		modsTemp &= ^xproto.ButtonMask2
	case 3:
		modsTemp &= ^xproto.ButtonMask3
	case 4:
		modsTemp &= ^xproto.ButtonMask4
	case 5:
		modsTemp &= ^xproto.ButtonMask5
	}

	return uint16(modsTemp), button
}

// ButtonPressFun represents a function that is called when a particular mouse
// binding is fired.
type ButtonPressFun xevent.ButtonPressFun

// If 'sync' is True, then no further events can be processed until the
// grabbing client allows them to be. (Which is done via AllowEvents. Thus,
// if sync is True, you *must* make some call to AllowEvents at some
// point, or else your client will lock.)
func (callback ButtonPressFun) Connect(xu *xgbutil.XUtil, win xproto.Window,
	buttonStr string, sync bool, grab bool) error {

	return connect(xu, callback, xevent.ButtonPress, win, buttonStr, sync, grab)
}

func (callback ButtonPressFun) Run(xu *xgbutil.XUtil, event interface{}) {
	callback(xu, event.(xevent.ButtonPressEvent))
}

// ButtonReleaseFun represents a function that is called when a particular mouse
// binding is fired.
type ButtonReleaseFun xevent.ButtonReleaseFun

// If 'sync' is True, then no further events can be processed until the
// grabbing client allows them to be. (Which is done via AllowEvents. Thus,
// if sync is True, you *must* make some call to AllowEvents at some
// point, or else your client will lock.)
func (callback ButtonReleaseFun) Connect(xu *xgbutil.XUtil, win xproto.Window,
	buttonStr string, sync bool, grab bool) error {

	return connect(xu, callback, xevent.ButtonRelease, win, buttonStr,
		sync, grab)
}

func (callback ButtonReleaseFun) Run(xu *xgbutil.XUtil, event interface{}) {
	callback(xu, event.(xevent.ButtonReleaseEvent))
}

// runButtonPressCallbacks infers the window, button and modifiers from a
// ButtonPressEvent and runs the corresponding callbacks.
func runButtonPressCallbacks(xu *xgbutil.XUtil, ev xevent.ButtonPressEvent) {
	mods, button := DeduceButtonInfo(ev.State, ev.Detail)

	runMouseBindCallbacks(xu, ev, xevent.ButtonPress, ev.Event, mods, button)
}

// runButtonReleaseCallbacks infers the window, keycode and modifiers from a
// ButtonPressEvent and runs the corresponding callbacks.
func runButtonReleaseCallbacks(xu *xgbutil.XUtil,
	ev xevent.ButtonReleaseEvent) {

	mods, button := DeduceButtonInfo(ev.State, ev.Detail)

	runMouseBindCallbacks(xu, ev, xevent.ButtonRelease, ev.Event, mods, button)
}

// Detach removes all handlers for all mouse events for the provided window id.
// This should be called whenever a window is no longer receiving events to make
// sure the garbage collector can release memory used to store the handler info.
func Detach(xu *xgbutil.XUtil, win xproto.Window) {
	detach(xu, xevent.ButtonPress, win)
	detach(xu, xevent.ButtonRelease, win)
}

// DetachPress is the same as Detach, except it only removes handlers for
// button *press* events.
func DetachPress(xu *xgbutil.XUtil, win xproto.Window) {
	detach(xu, xevent.ButtonPress, win)
}

// DetachRelease is the same as Detach, except it only removes handlers for
// mouse *release* events.
func DetachRelease(xu *xgbutil.XUtil, win xproto.Window) {
	detach(xu, xevent.ButtonRelease, win)
}

// detach removes all handlers for the provided window and event type
// combination. This will also issue an ungrab request for each grab that
// drops to zero.
func detach(xu *xgbutil.XUtil, evtype int, win xproto.Window) {
	mkeys := mouseKeys(xu)
	detachMouseBindWindow(xu, evtype, win)
	for _, key := range mkeys {
		if mouseBindGrabs(xu, key.Evtype, key.Win, key.Mod, key.Button) == 0 {
			Ungrab(xu, key.Win, key.Mod, key.Button)
		}
	}
}
