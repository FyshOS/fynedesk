package keybind

/*
keybind/xutil.go contains a collection of functions that modify the
Keybinds and Keygrabs state of an XUtil value.

They could have been placed inside the core xgbutil package, but they would
have to be exported for use by the keybind package. In which case, the API
would become cluttered with functions that should not be used.
*/

import (
	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil"
)

// attachKeyBindCallback associates an (event, window, mods, keycode)
// with a callback.
// This is exported for use in the keybind package. It should not be used.
func attachKeyBindCallback(xu *xgbutil.XUtil, evtype int, win xproto.Window,
	mods uint16, keycode xproto.Keycode, fun xgbutil.CallbackKey) {

	xu.KeybindsLck.Lock()
	defer xu.KeybindsLck.Unlock()

	// Create key
	key := xgbutil.KeyKey{evtype, win, mods, keycode}

	// Do we need to allocate?
	if _, ok := xu.Keybinds[key]; !ok {
		xu.Keybinds[key] = make([]xgbutil.CallbackKey, 0)
	}

	xu.Keybinds[key] = append(xu.Keybinds[key], fun)
	xu.Keygrabs[key] += 1
}

// addKeyString adds a new key binding string to XUtil.Keystrings.
// The invariant is that each key string appears once and only once.
func addKeyString(xu *xgbutil.XUtil, callback xgbutil.CallbackKey,
	evtype int, win xproto.Window, keyStr string, grab bool) {

	xu.KeybindsLck.Lock()
	defer xu.KeybindsLck.Unlock()

	k := xgbutil.KeyString{
		Str:      keyStr,
		Callback: callback,
		Evtype:   evtype,
		Win:      win,
		Grab:     grab,
	}
	xu.Keystrings = append(xu.Keystrings, k)
}

// keyBindKeys returns a copy of all the keys in the 'keybinds' map.
func keyKeys(xu *xgbutil.XUtil) []xgbutil.KeyKey {
	xu.KeybindsLck.RLock()
	defer xu.KeybindsLck.RUnlock()

	keys := make([]xgbutil.KeyKey, len(xu.Keybinds))
	i := 0
	for key, _ := range xu.Keybinds {
		keys[i] = key
		i++
	}
	return keys
}

// runKeyBindCallbacks executes every callback corresponding to a
// particular event/window/mod/key tuple.
// This is exported for use in the keybind package. It should not be used.
func runKeyBindCallbacks(xu *xgbutil.XUtil, event interface{}, evtype int,
	win xproto.Window, mods uint16, keycode xproto.Keycode) {

	key := xgbutil.KeyKey{evtype, win, mods, keycode}
	for _, cb := range keyCallbacks(xu, key) {
		cb.Run(xu, event)
	}
}

// keyBindCallbacks returns a slice of callbacks for a particular key.
func keyCallbacks(xu *xgbutil.XUtil,
	key xgbutil.KeyKey) []xgbutil.CallbackKey {

	xu.KeybindsLck.RLock()
	defer xu.KeybindsLck.RUnlock()

	cbs := make([]xgbutil.CallbackKey, len(xu.Keybinds[key]))
	copy(cbs, xu.Keybinds[key])
	return cbs
}

// ConnectedKeyBind checks to see if there are any key binds for a particular
// event type already in play.
func connectedKeyBind(xu *xgbutil.XUtil, evtype int, win xproto.Window) bool {
	xu.KeybindsLck.RLock()
	defer xu.KeybindsLck.RUnlock()

	// Since we can't create a full key, loop through all key binds
	// and check if evtype and window match.
	for key, _ := range xu.Keybinds {
		if key.Evtype == evtype && key.Win == win {
			return true
		}
	}
	return false
}

// detachKeyBindWindow removes all callbacks associated with a particular
// window and event type (either KeyPress or KeyRelease)
// Also decrements the counter in the corresponding 'keygrabs' map
// appropriately.
// This is exported for use in the keybind package. It should not be used.
// To detach a window from a key binding callbacks, please use keybind.Detach.
// (This method will issue an Ungrab requests, while keybind.Detach will.)
func detachKeyBindWindow(xu *xgbutil.XUtil, evtype int, win xproto.Window) {
	xu.KeybindsLck.Lock()
	defer xu.KeybindsLck.Unlock()

	// Since we can't create a full key, loop through all key binds
	// and check if evtype and window match.
	for key, _ := range xu.Keybinds {
		if key.Evtype == evtype && key.Win == win {
			xu.Keygrabs[key] -= len(xu.Keybinds[key])
			delete(xu.Keybinds, key)
		}
	}
}

// keyBindGrabs returns the number of grabs on a particular
// event/window/mods/keycode combination. Namely, this combination
// uniquely identifies a grab. If it's repeated, we get BadAccess.
// The idea is that if there are 0 grabs on a particular (modifiers, keycode)
// tuple, then we issue a grab request. Otherwise, we don't.
// This is exported for use in the keybind package. It should not be used.
func keyBindGrabs(xu *xgbutil.XUtil, evtype int, win xproto.Window, mods uint16,
	keycode xproto.Keycode) int {

	xu.KeybindsLck.RLock()
	defer xu.KeybindsLck.RUnlock()

	key := xgbutil.KeyKey{evtype, win, mods, keycode}
	return xu.Keygrabs[key] // returns 0 if key does not exist
}

// KeyMapGet accessor.
func KeyMapGet(xu *xgbutil.XUtil) *xgbutil.KeyboardMapping {
	return xu.Keymap
}

// KeyMapSet updates XUtil.keymap.
// This is exported for use in the keybind package. You probably shouldn't
// use this. (You may need to use this if you're rolling your own event loop,
// and still want to use the keybind package.)
func KeyMapSet(xu *xgbutil.XUtil, keyMapReply *xproto.GetKeyboardMappingReply) {
	xu.Keymap = &xgbutil.KeyboardMapping{keyMapReply}
}

// ModMapGet accessor.
func ModMapGet(xu *xgbutil.XUtil) *xgbutil.ModifierMapping {
	return xu.Modmap
}

// ModMapSet updates XUtil.modmap.
// This is exported for use in the keybind package. You probably shouldn't
// use this. (You may need to use this if you're rolling your own event loop,
// and still want to use the keybind package.)
func ModMapSet(xu *xgbutil.XUtil, modMapReply *xproto.GetModifierMappingReply) {
	xu.Modmap = &xgbutil.ModifierMapping{modMapReply}
}
