package keybind

import (
	"fmt"

	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xevent"
)

// connect is essentially 'Connect' for either KeyPress or KeyRelease events.
// Namely, it parses the key string, issues a grab request if necessary,
// sets up the appropriate event handlers for the main event loop, and attaches
// the callback to the keybinding state.
func connect(xu *xgbutil.XUtil, callback xgbutil.CallbackKey,
	evtype int, win xproto.Window, keyStr string, grab, reconnect bool) error {

	// Get the mods/key first
	mods, keycodes, err := ParseString(xu, keyStr)
	if err != nil {
		return err
	}

	// Only do the grab if we haven't yet on this window.
	for _, keycode := range keycodes {
		if grab && keyBindGrabs(xu, evtype, win, mods, keycode) == 0 {
			if err := GrabChecked(xu, win, mods, keycode); err != nil {
				// If a bad access, let's be nice and give a good error message.
				switch err.(type) {
				case xproto.AccessError:
					return fmt.Errorf("Got a bad access error when trying to "+
						"bind '%s'. This usually means another client has "+
						"already grabbed this keybinding.", keyStr)
				default:
					return fmt.Errorf("Could not bind '%s' because: %s",
						keyStr, err)
				}
			}
		}

		// If we've never grabbed anything on this window before, we need to
		// make sure we can respond to it in the main event loop.
		// Never do this if we're reconnecting.
		if !reconnect {
			var allCb xgbutil.Callback
			if evtype == xevent.KeyPress {
				allCb = xevent.KeyPressFun(runKeyPressCallbacks)
			} else {
				allCb = xevent.KeyReleaseFun(runKeyReleaseCallbacks)
			}

			// If this is the first Key{Press|Release}Event on this window,
			// then we need to listen to Key{Press|Release} events in the main
			// loop.
			if !connectedKeyBind(xu, evtype, win) {
				allCb.Connect(xu, win)
			}
		}

		// Finally, attach the callback.
		attachKeyBindCallback(xu, evtype, win, mods, keycode, callback)
	}

	// Keep track of all unique key connections.
	if !reconnect {
		addKeyString(xu, callback, evtype, win, keyStr, grab)
	}

	return nil
}

// DeduceKeyInfo AND's the "ignored modifiers" out of the state returned by
// a Key{Press,Release} event. This is useful to connect a (state, keycode)
// tuple from an event with a tuple specified by the user.
func DeduceKeyInfo(state uint16,
	detail xproto.Keycode) (uint16, xproto.Keycode) {

	mods, kc := state, detail
	for _, m := range xevent.IgnoreMods {
		mods &= ^m
	}
	return mods, kc
}

// KeyPressFun represents a function that is called when a particular key
// binding is fired.
type KeyPressFun xevent.KeyPressFun

func (callback KeyPressFun) Connect(xu *xgbutil.XUtil, win xproto.Window,
	keyStr string, grab bool) error {

	return connect(xu, callback, xevent.KeyPress, win, keyStr, grab, false)
}

func (callback KeyPressFun) Run(xu *xgbutil.XUtil, event interface{}) {
	callback(xu, event.(xevent.KeyPressEvent))
}

// KeyReleaseFun represents a function that is called when a particular key
// binding is fired.
type KeyReleaseFun xevent.KeyReleaseFun

func (callback KeyReleaseFun) Connect(xu *xgbutil.XUtil, win xproto.Window,
	keyStr string, grab bool) error {

	return connect(xu, callback, xevent.KeyRelease, win, keyStr, grab, false)
}

func (callback KeyReleaseFun) Run(xu *xgbutil.XUtil, event interface{}) {
	callback(xu, event.(xevent.KeyReleaseEvent))
}

// runKeyPressCallbacks infers the window, keycode and modifiers from a
// KeyPressEvent and runs the corresponding callbacks.
func runKeyPressCallbacks(xu *xgbutil.XUtil, ev xevent.KeyPressEvent) {
	mods, kc := DeduceKeyInfo(ev.State, ev.Detail)

	runKeyBindCallbacks(xu, ev, xevent.KeyPress, ev.Event, mods, kc)
}

// runKeyReleaseCallbacks infers the window, keycode and modifiers from a
// KeyPressEvent and runs the corresponding callbacks.
func runKeyReleaseCallbacks(xu *xgbutil.XUtil, ev xevent.KeyReleaseEvent) {
	mods, kc := DeduceKeyInfo(ev.State, ev.Detail)

	runKeyBindCallbacks(xu, ev, xevent.KeyRelease, ev.Event, mods, kc)
}

// Detach removes all handlers for all key events for the provided window id.
// This should be called whenever a window is no longer receiving events to make
// sure the garbage collector can release memory used to store the handler info.
func Detach(xu *xgbutil.XUtil, win xproto.Window) {
	detach(xu, xevent.KeyPress, win)
	detach(xu, xevent.KeyRelease, win)
}

// DetachPress is the same as Detach, except it only removes handlers for
// key *press* events.
func DetachPress(xu *xgbutil.XUtil, win xproto.Window) {
	detach(xu, xevent.KeyPress, win)
}

// DetachRelease is the same as Detach, except it only removes handlers for
// key *release* events.
func DetachRelease(xu *xgbutil.XUtil, win xproto.Window) {
	detach(xu, xevent.KeyRelease, win)
}

// detach removes all handlers for the provided window and event type
// combination. This will also issue an ungrab request for each grab that
// drops to zero.
func detach(xu *xgbutil.XUtil, evtype int, win xproto.Window) {
	mkeys := keyKeys(xu)
	detachKeyBindWindow(xu, evtype, win)
	for _, key := range mkeys {
		if keyBindGrabs(xu, key.Evtype, key.Win, key.Mod, key.Code) == 0 {
			Ungrab(xu, key.Win, key.Mod, key.Code)
		}
	}
}
