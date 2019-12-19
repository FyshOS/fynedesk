package mousebind

/*
mousebind/xutil.go contains a collection of functions that modify the
Mousebinds and Mousegrabs state of an XUtil value.

They could have been placed inside the core xgbutil package, but they would
have to be exported for use by the mousebind package. In which case, the API
would become cluttered with functions that should not be used.
*/

import (
	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil"
)

// attachMouseBindCallback associates an (event, window, mods, button)
// with a callback.
func attachMouseBindCallback(xu *xgbutil.XUtil, evtype int, win xproto.Window,
	mods uint16, button xproto.Button, fun xgbutil.CallbackMouse) {

	xu.MousebindsLck.Lock()
	defer xu.MousebindsLck.Unlock()

	// Create key
	key := xgbutil.MouseKey{evtype, win, mods, button}

	// Do we need to allocate?
	if _, ok := xu.Mousebinds[key]; !ok {
		xu.Mousebinds[key] = make([]xgbutil.CallbackMouse, 0)
	}

	xu.Mousebinds[key] = append(xu.Mousebinds[key], fun)
	xu.Mousegrabs[key] += 1
}

// mouseKeys returns a copy of all the keys in the 'Mousebinds' map.
func mouseKeys(xu *xgbutil.XUtil) []xgbutil.MouseKey {
	xu.MousebindsLck.RLock()
	defer xu.MousebindsLck.RUnlock()

	keys := make([]xgbutil.MouseKey, len(xu.Mousebinds))
	i := 0
	for key, _ := range xu.Mousebinds {
		keys[i] = key
		i++
	}
	return keys
}

// mouseBindCallbacks returns a slice of callbacks for a particular key.
func mouseCallbacks(xu *xgbutil.XUtil,
	key xgbutil.MouseKey) []xgbutil.CallbackMouse {

	xu.MousebindsLck.RLock()
	defer xu.MousebindsLck.RUnlock()

	cbs := make([]xgbutil.CallbackMouse, len(xu.Mousebinds[key]))
	copy(cbs, xu.Mousebinds[key])
	return cbs
}

// runMouseBindCallbacks executes every callback corresponding to a
// particular event/window/mod/button tuple.
func runMouseBindCallbacks(xu *xgbutil.XUtil, event interface{}, evtype int,
	win xproto.Window, mods uint16, button xproto.Button) {

	key := xgbutil.MouseKey{evtype, win, mods, button}
	for _, cb := range mouseCallbacks(xu, key) {
		cb.Run(xu, event)
	}
}

// connectedMouseBind checks to see if there are any key binds for a particular
// event type already in play. This is to work around comparing function
// pointers (not allowed in Go), which would be used in 'Connected'.
func connectedMouseBind(xu *xgbutil.XUtil, evtype int, win xproto.Window) bool {
	xu.MousebindsLck.RLock()
	defer xu.MousebindsLck.RUnlock()

	// Since we can't create a full key, loop through all mouse binds
	// and check if evtype and window match.
	for key, _ := range xu.Mousebinds {
		if key.Evtype == evtype && key.Win == win {
			return true
		}
	}
	return false
}

// detachMouseBindWindow removes all callbacks associated with a particular
// window and event type (either ButtonPress or ButtonRelease)
// Also decrements the counter in the corresponding 'Mousegrabs' map
// appropriately.
func detachMouseBindWindow(xu *xgbutil.XUtil, evtype int, win xproto.Window) {
	xu.MousebindsLck.Lock()
	defer xu.MousebindsLck.Unlock()

	// Since we can't create a full key, loop through all mouse binds
	// and check if evtype and window match.
	for key, _ := range xu.Mousebinds {
		if key.Evtype == evtype && key.Win == win {
			xu.Mousegrabs[key] -= len(xu.Mousebinds[key])
			delete(xu.Mousebinds, key)
		}
	}
}

// mouseBindGrabs returns the number of grabs on a particular
// event/window/mods/button combination. Namely, this combination
// uniquely identifies a grab. If it's repeated, we get BadAccess.
func mouseBindGrabs(xu *xgbutil.XUtil, evtype int, win xproto.Window,
	mods uint16, button xproto.Button) int {

	xu.MousebindsLck.RLock()
	defer xu.MousebindsLck.RUnlock()

	key := xgbutil.MouseKey{evtype, win, mods, button}
	return xu.Mousegrabs[key] // returns 0 if key does not exist
}

// mouseDrag true when a mouse drag is in progress.
func mouseDrag(xu *xgbutil.XUtil) bool {
	return xu.InMouseDrag
}

// mouseDragSet sets whether a mouse drag is in progress.
func mouseDragSet(xu *xgbutil.XUtil, dragging bool) {
	xu.InMouseDrag = dragging
}

// mouseDragStep returns the function currently associated with each
// step of a mouse drag.
func mouseDragStep(xu *xgbutil.XUtil) xgbutil.MouseDragFun {
	return xu.MouseDragStepFun
}

// mouseDragStepSet sets the function associated with the step of a drag.
func mouseDragStepSet(xu *xgbutil.XUtil, f xgbutil.MouseDragFun) {
	xu.MouseDragStepFun = f
}

// mouseDragEnd returns the function currently associated with the
// end of a mouse drag.
func mouseDragEnd(xu *xgbutil.XUtil) xgbutil.MouseDragFun {
	return xu.MouseDragEndFun
}

// mouseDragEndSet sets the function associated with the end of a drag.
func mouseDragEndSet(xu *xgbutil.XUtil, f xgbutil.MouseDragFun) {
	xu.MouseDragEndFun = f
}
