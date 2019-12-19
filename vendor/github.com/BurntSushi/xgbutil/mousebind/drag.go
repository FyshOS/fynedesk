package mousebind

import (
	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xevent"
)

// Drag is the public interface that will make the appropriate connections
// to register a drag event for three functions: the begin function, the
// step function and the end function.
// The 'grabwin' is the window that the grab is placed on (and therefore the
// window where all button events are redirected to after the drag has started),
// and the 'win' is the window that the initial 'begin' callback is set on.
// In typical use cases, these windows should be the same.
// If 'grab' is false, then no pointer grab is issued.
func Drag(xu *xgbutil.XUtil, grabwin xproto.Window, win xproto.Window,
	buttonStr string, grab bool,
	begin xgbutil.MouseDragBeginFun, step xgbutil.MouseDragFun,
	end xgbutil.MouseDragFun) {

	ButtonPressFun(
		func(xu *xgbutil.XUtil, ev xevent.ButtonPressEvent) {
			DragBegin(xu, ev, grabwin, win, begin, step, end)
		}).Connect(xu, win, buttonStr, false, grab)

	// If the grab win isn't the dummy, then setup event handlers for the
	// grab window.
	if grabwin != xu.Dummy() {
		xevent.MotionNotifyFun(dragStep).Connect(xu, grabwin)
		xevent.ButtonReleaseFun(DragEnd).Connect(xu, grabwin)
	}
}

// dragGrab is a shortcut for grabbing the pointer for a drag.
func dragGrab(xu *xgbutil.XUtil, grabwin xproto.Window, win xproto.Window,
	cursor xproto.Cursor) bool {

	status, err := GrabPointer(xu, grabwin, xu.RootWin(), cursor)
	if err != nil {
		xgbutil.Logger.Printf("Mouse dragging was unsuccessful because: %v",
			err)
		return false
	}
	if !status {
		xgbutil.Logger.Println("Mouse dragging was unsuccessful because " +
			"we could not establish a pointer grab.")
		return false
	}

	mouseDragSet(xu, true)
	return true
}

// dragUngrab is a shortcut for ungrabbing the pointer for a drag.
func dragUngrab(xu *xgbutil.XUtil) {
	UngrabPointer(xu)
	mouseDragSet(xu, false)
}

// DragBegin executes the "begin" function registered for the current drag.
// It also initiates the grab with the cursor id return by the begin callback.
//
// N.B. This function is automatically called in the Drag convenience function.
// This should be used when the drag can be started from a source other than
// a button press handled by the WM. If you use this function, then there
// should also be a call to DragEnd when the drag is done. (This is
// automatically done for you if you use Drag.)
func DragBegin(xu *xgbutil.XUtil, ev xevent.ButtonPressEvent,
	grabwin xproto.Window, win xproto.Window,
	begin xgbutil.MouseDragBeginFun, step xgbutil.MouseDragFun,
	end xgbutil.MouseDragFun) {

	// don't start a drag if one is already in progress
	if mouseDrag(xu) {
		return
	}

	// Run begin first. It may tell us to cancel the grab.
	// It can also tell us which cursor to use when grabbing.
	grab, cursor := begin(xu, int(ev.RootX), int(ev.RootY),
		int(ev.EventX), int(ev.EventY))

	// if we couldn't establish a grab, quit
	// Or quit if 'begin' tells us to.
	if !grab || !dragGrab(xu, grabwin, win, cursor) {
		return
	}

	// we're committed. set the drag state and start the 'begin' function
	mouseDragStepSet(xu, step)
	mouseDragEndSet(xu, end)
}

// dragStep executes the "step" function registered for the current drag.
// It also compresses the MotionNotify events.
func dragStep(xu *xgbutil.XUtil, ev xevent.MotionNotifyEvent) {
	// If for whatever reason we don't have any *piece* of a grab,
	// we've gotta back out.
	if !mouseDrag(xu) || mouseDragStep(xu) == nil || mouseDragEnd(xu) == nil {
		dragUngrab(xu)
		mouseDragStepSet(xu, nil)
		mouseDragEndSet(xu, nil)
		return
	}

	// The most recent MotionNotify event that we'll end up returning.
	laste := ev

	// We force a round trip request so that we make sure to read all
	// available events.
	xu.Sync()
	xevent.Read(xu, false)

	// Compress MotionNotify events.
	for i, ee := range xevent.Peek(xu) {
		if ee.Err != nil { // This is an error, skip it.
			continue
		}

		// Use type assertion to make sure this is a MotionNotify event.
		if mn, ok := ee.Event.(xproto.MotionNotifyEvent); ok {
			// Now make sure all appropriate fields are equivalent.
			if ev.Event == mn.Event && ev.Child == mn.Child &&
				ev.Detail == mn.Detail && ev.State == mn.State &&
				ev.Root == mn.Root && ev.SameScreen == mn.SameScreen {

				// Set the most recent/valid motion notify event.
				laste = xevent.MotionNotifyEvent{&mn}

				// We cheat and use the stack semantics of defer to dequeue
				// most recent motion notify events first, so that the indices
				// don't become invalid. (If we dequeued oldest first, we'd
				// have to account for all future events shifting to the left
				// by one.)
				defer func(i int) { xevent.DequeueAt(xu, i) }(i)
			}
		}
	}
	xu.TimeSet(laste.Time)

	// now actually run the step
	mouseDragStep(xu)(xu, int(laste.RootX), int(laste.RootY),
		int(laste.EventX), int(laste.EventY))
}

// DragEnd executes the "end" function registered for the current drag.
// This must be called at some point if DragStart has been called.
func DragEnd(xu *xgbutil.XUtil, ev xevent.ButtonReleaseEvent) {
	if mouseDragEnd(xu) != nil {
		mouseDragEnd(xu)(xu, int(ev.RootX), int(ev.RootY),
			int(ev.EventX), int(ev.EventY))
	}

	dragUngrab(xu)
	mouseDragStepSet(xu, nil)
	mouseDragEndSet(xu, nil)
}
