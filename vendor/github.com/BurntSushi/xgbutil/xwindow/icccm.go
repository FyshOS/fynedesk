package xwindow

import (
	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/icccm"
	"github.com/BurntSushi/xgbutil/xevent"
)

// WMGracefulClose will do all the necessary setup to implement the
// WM_DELETE_WINDOW protocol. This will prevent well-behaving window managers
// from killing your client whenever one of your windows is closed. (Killing
// a client is bad because it will destroy your X connection and any other
// clients you have open.)
// You must provide a callback function that is called when the window manager
// asks you to close your window. (You may provide some means of confirmation
// to the user, i.e., "Do you really want to quit?", but you should probably
// just wrap things up and call DestroyWindow.)
func (w *Window) WMGracefulClose(cb func(w *Window)) {
	// Get the current protocols so we don't overwrite anything.
	prots, _ := icccm.WmProtocolsGet(w.X, w.Id)

	// If WM_DELETE_WINDOW isn't here, add it. Otherwise, move on.
	wmdelete := false
	for _, prot := range prots {
		if prot == "WM_DELETE_WINDOW" {
			wmdelete = true
			break
		}
	}
	if !wmdelete {
		icccm.WmProtocolsSet(w.X, w.Id, append(prots, "WM_DELETE_WINDOW"))
	}

	// Attach a ClientMessage event handler. It will determine whether the
	// ClientMessage is a 'close' request, and if so, run the callback 'cb'
	// provided.
	xevent.ClientMessageFun(
		func(X *xgbutil.XUtil, ev xevent.ClientMessageEvent) {
			if icccm.IsDeleteProtocol(X, ev) {
				cb(w)
			}
		}).Connect(w.X, w.Id)
}

// WMTakeFocus will do all the necessary setup to support the WM_TAKE_FOCUS
// protocol using the "LocallyActive" input model described in Section 4.1.7
// of the ICCCM. Namely, listening to ClientMessage events and running the
// callback function provided when a WM_TAKE_FOCUS ClientMessage has been
// received.
//
// Typically, the callback function should include a call to SetInputFocus
// with the "Parent" InputFocus type, the sub-window id of the window that
// should have focus, and the 'tstamp' timestamp.
func (w *Window) WMTakeFocus(cb func(w *Window, tstamp xproto.Timestamp)) {
	// Make sure the Input flag is set to true in WM_HINTS. We first
	// must retrieve the current WM_HINTS, so we don't overwrite the flags.
	curFlags := uint(0)
	if hints, err := icccm.WmHintsGet(w.X, w.Id); err == nil {
		curFlags = hints.Flags
	}
	icccm.WmHintsSet(w.X, w.Id, &icccm.Hints{
		Flags: curFlags | icccm.HintInput,
		Input: 1,
	})

	// Get the current protocols so we don't overwrite anything.
	prots, _ := icccm.WmProtocolsGet(w.X, w.Id)

	// If WM_TAKE_FOCUS isn't here, add it. Otherwise, move on.
	wmfocus := false
	for _, prot := range prots {
		if prot == "WM_TAKE_FOCUS" {
			wmfocus = true
			break
		}
	}
	if !wmfocus {
		icccm.WmProtocolsSet(w.X, w.Id, append(prots, "WM_TAKE_FOCUS"))
	}

	// Attach a ClientMessage event handler. It will determine whether the
	// ClientMessage is a 'focus' request, and if so, run the callback 'cb'
	// provided.
	xevent.ClientMessageFun(
		func(X *xgbutil.XUtil, ev xevent.ClientMessageEvent) {
			if icccm.IsFocusProtocol(X, ev) {
				cb(w, xproto.Timestamp(ev.Data.Data32[1]))
			}
		}).Connect(w.X, w.Id)
}
