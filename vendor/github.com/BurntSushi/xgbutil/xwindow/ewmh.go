package xwindow

/*
xwindow/ewmh.go contains several methods that rely on EWMH support in
the currently running window manager.
*/

import (
	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/xrect"
)

// DecorGeometry retrieves the client's width and height including decorations.
// This can be tricky. In a non-parenting window manager, the width/height of
// a client can be found by inspecting the client directly. In a reparenting
// window manager like Openbox, the parent of the client reflects the true
// width/height. Still yet, in KWin, it's the parent of the parent of the
// client that reflects the true width/height.
// The idea then is to traverse up the tree until we hit the root window.
// Therefore, we're at a top-level window which should accurately reflect
// the width/height.
func (w *Window) DecorGeometry() (xrect.Rect, error) {
	parent := w
	for {
		tempParent, err := parent.Parent()
		if err != nil || tempParent.Id == w.X.RootWin() {
			break
		}
		parent = tempParent
	}
	return RawGeometry(w.X, xproto.Drawable(parent.Id))
}

// WMMoveResize is an accurate means of resizing a window, accounting for
// decorations. Usually, the x,y coordinates are fine---we just need to
// adjust the width and height.
// This should be used when moving/resizing top-level client windows with
// reparenting window managers that support EWMH.
func (w *Window) WMMoveResize(x, y, width, height int) error {
	neww, newh, err := adjustSize(w.X, w.Id, width, height)
	if err != nil {
		return err
	}
	return ewmh.MoveresizeWindowExtra(w.X, w.Id, x, y, neww, newh,
		xproto.GravityBitForget, 2, true, true)
}

// WMMove changes the position of a window without touching the size.
// This should be used when moving a top-level client window with
// reparenting winow managers that support EWMH.
func (w *Window) WMMove(x, y int) error {
	return ewmh.MoveWindow(w.X, w.Id, x, y)
}

// WMResize changes the size of a window without touching the position.
// This should be used when resizing a top-level client window with
// reparenting window managers that support EWMH.
func (w *Window) WMResize(width, height int) error {
	neww, newh, err := adjustSize(w.X, w.Id, width, height)
	if err != nil {
		return err
	}
	return ewmh.ResizeWindow(w.X, w.Id, neww, newh)
}

// adjustSize takes a client and dimensions, and adjust them so that they'll
// account for window decorations. For example, if you want a window to be
// 200 pixels wide, a window manager will typically determine that as
// you wanting the *client* to be 200 pixels wide. The end result is that
// the client plus decorations ends up being
// (200 + left decor width + right decor width) pixels wide. Which is probably
// not what you want. Therefore, transform 200 into
// 200 - decoration window width - client window width.
// Similarly for height.
func adjustSize(xu *xgbutil.XUtil, win xproto.Window,
	w, h int) (int, int, error) {

	// raw client geometry
	cGeom, err := RawGeometry(xu, xproto.Drawable(win))
	if err != nil {
		return 0, 0, err
	}

	// geometry with decorations
	pGeom, err := RawGeometry(xu, xproto.Drawable(win))
	if err != nil {
		return 0, 0, err
	}

	neww := w - (pGeom.Width() - cGeom.Width())
	newh := h - (pGeom.Height() - cGeom.Height())
	if neww < 1 {
		neww = 1
	}
	if newh < 1 {
		newh = 1
	}
	return neww, newh, nil
}
