package xwindow

import (
	"fmt"

	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/keybind"
	"github.com/BurntSushi/xgbutil/mousebind"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xrect"
)

// Window represents an X window. It contains an XUtilValue to simplfy the
// parameter lists for methods declared on the Window type.
// Geom is updated whenever Geometry is called, or when Move, Resize or
// MoveResize are called.
type Window struct {
	X         *xgbutil.XUtil
	Id        xproto.Window
	Geom      xrect.Rect
	Destroyed bool
}

// New creates a new window value from a window id and an XUtil type.
// Geom is initialized to zero values. Use Window.Geometry to load it.
// Note that the geometry is the size of this particular window and nothing
// else. If you want the geometry of a client window including decorations,
// please use Window.DecorGeometry.
func New(xu *xgbutil.XUtil, win xproto.Window) *Window {
	return &Window{
		X:         xu,
		Id:        win,
		Geom:      xrect.New(0, 0, 1, 1),
		Destroyed: false,
	}
}

// Generate is just like New, but generates a new X resource id for you.
// Geom is initialized to (0, 0) 1x1.
// It is possible for id generation to return an error, in which case, an
// error is returned here.
func Generate(xu *xgbutil.XUtil) (*Window, error) {
	wid, err := xproto.NewWindowId(xu.Conn())
	if err != nil {
		return nil, err
	}
	return New(xu, wid), nil
}

// Create is a convenience constructor that will generate a new window id (with
// the Generate constructor) and make a bare-bones call to CreateChecked (with
// geometry (0, 0) 1x1). An error can be generated from Generate or
// CreateChecked.
func Create(xu *xgbutil.XUtil, parent xproto.Window) (*Window, error) {
	win, err := Generate(xu)
	if err != nil {
		return nil, err
	}

	err = win.CreateChecked(parent, 0, 0, 1, 1, 0)
	if err != nil {
		return nil, err
	}

	return win, nil
}

// Must panics if err is non-nil or if win is nil. Otherwise, win is returned.
func Must(win *Window, err error) *Window {
	if err != nil {
		panic(err)
	}
	if win == nil {
		panic("win and err are nil")
	}
	return win
}

// Create issues a CreateWindow request for Window.
// Its purpose is to omit several boiler-plate parameters to CreateWindow
// and expose the commonly useful ones.
// The value mask describes which values are present in valueList.
// Value masks can be found in xgb/xproto with the prefix 'Cw'.
// The value list must contain values in the same order as the constants
// are defined in xgb/xproto.
//
// For example, the following creates a window positioned at (20, 50) with
// width 500 and height 700 with a background color of white.
//
//	w, err := xwindow.Generate(X)
//	if err != nil {
//		log.Fatalf("Could not generate a new resource identifier: %s", err)
//	}
//	w.Create(X.RootWin(), 20, 50, 500, 700,
//		xproto.CwBackPixel, 0xffffff)
func (w *Window) Create(parent xproto.Window, x, y, width, height,
	valueMask int, valueList ...uint32) {

	s := w.X.Screen()
	xproto.CreateWindow(w.X.Conn(),
		xproto.WindowClassCopyFromParent, w.Id, parent,
		int16(x), int16(y), uint16(width), uint16(height), 0,
		xproto.WindowClassInputOutput, s.RootVisual,
		uint32(valueMask), valueList)
}

// CreateChecked issues a CreateWindow checked request for Window.
// A checked request is a synchronous request. Meaning that if the request
// fails, you can get the error returned to you. However, it also forced your
// program to block for a round trip to the X server, so it is slower.
// See the docs for Create for more info.
func (w *Window) CreateChecked(parent xproto.Window, x, y, width, height,
	valueMask int, valueList ...uint32) error {

	s := w.X.Screen()
	return xproto.CreateWindowChecked(w.X.Conn(),
		s.RootDepth, w.Id, parent,
		int16(x), int16(y), uint16(width), uint16(height), 0,
		xproto.WindowClassInputOutput, s.RootVisual,
		uint32(valueMask), valueList).Check()
}

// Change issues a ChangeWindowAttributes request with the provided mask
// and value list. Please see Window.Create for an example on how to use
// the mask and value list.
func (w *Window) Change(valueMask int, valueList ...uint32) {
	xproto.ChangeWindowAttributes(w.X.Conn(), w.Id,
		uint32(valueMask), valueList)
}

// Listen will tell X to report events corresponding to the event masks
// provided for the given window. If a call to Listen is omitted, you will
// not receive the events you desire.
// Event masks are constants declare in the xgb/xproto package starting with the
// EventMask prefix.
func (w *Window) Listen(evMasks ...int) error {
	evMask := 0
	for _, mask := range evMasks {
		evMask |= mask
	}
	return xproto.ChangeWindowAttributesChecked(w.X.Conn(), w.Id,
		xproto.CwEventMask, []uint32{uint32(evMask)}).Check()
}

// Geometry retrieves an up-to-date version of the this window's geometry.
// It also loads the geometry into the Geom member of Window.
func (w *Window) Geometry() (xrect.Rect, error) {
	geom, err := RawGeometry(w.X, xproto.Drawable(w.Id))
	if err != nil {
		return nil, err
	}
	w.Geom = geom
	return geom, err
}

// RawGeometry isn't smart. It just queries the window given for geometry.
func RawGeometry(xu *xgbutil.XUtil, win xproto.Drawable) (xrect.Rect, error) {
	xgeom, err := xproto.GetGeometry(xu.Conn(), win).Reply()
	if err != nil {
		return nil, err
	}
	return xrect.New(int(xgeom.X), int(xgeom.Y),
		int(xgeom.Width), int(xgeom.Height)), nil
}

// RootGeometry gets the geometry of the root window. It will panic on failure.
func RootGeometry(xu *xgbutil.XUtil) xrect.Rect {
	geom, err := RawGeometry(xu, xproto.Drawable(xu.RootWin()))
	if err != nil {
		panic(err)
	}
	return geom
}

// Configure issues a raw Configure request with the parameters given and
// updates the geometry of the window.
// This should probably only be used when passing along ConfigureNotify events
// (from the perspective of the window manager). In other cases, one should
// opt for [WM][Move][Resize] or Stack[Sibling].
func (win *Window) Configure(flags, x, y, w, h int,
	sibling xproto.Window, stackMode byte) {

	if win == nil {
		return
	}

	vals := []uint32{}

	if xproto.ConfigWindowX&flags > 0 {
		vals = append(vals, uint32(x))
		win.Geom.XSet(x)
	}
	if xproto.ConfigWindowY&flags > 0 {
		vals = append(vals, uint32(y))
		win.Geom.YSet(y)
	}
	if xproto.ConfigWindowWidth&flags > 0 {
		if int16(w) <= 0 {
			w = 1
		}
		vals = append(vals, uint32(w))
		win.Geom.WidthSet(w)
	}
	if xproto.ConfigWindowHeight&flags > 0 {
		if int16(h) <= 0 {
			h = 1
		}
		vals = append(vals, uint32(h))
		win.Geom.HeightSet(h)
	}
	if xproto.ConfigWindowSibling&flags > 0 {
		vals = append(vals, uint32(sibling))
	}
	if xproto.ConfigWindowStackMode&flags > 0 {
		vals = append(vals, uint32(stackMode))
	}

	// Nobody should be setting border widths any more.
	// We toss it out since `vals` must have length equal to the number
	// of bits set in `flags`.
	flags &= ^xproto.ConfigWindowBorderWidth
	xproto.ConfigureWindow(win.X.Conn(), win.Id, uint16(flags), vals)
}

// MROpt is like MoveResize, but exposes the X value mask so that any
// combination of x/y/width/height can be set. It's a strictly convenience
// function. (i.e., when you need to set 'y' and 'height' but not 'x' or
// 'width'.)
func (w *Window) MROpt(flags, x, y, width, height int) {
	// Make sure only x/y/width/height are used.
	flags &= xproto.ConfigWindowX | xproto.ConfigWindowY |
		xproto.ConfigWindowWidth | xproto.ConfigWindowHeight
	w.Configure(flags, x, y, width, height, 0, 0)
}

// MoveResize issues a ConfigureRequest for this window with the provided
// x, y, width and height. Note that if width or height is 0, X will stomp
// all over you. Really hard. Don't do it.
// If you're trying to move/resize a top-level window in a window manager that
// supports EWMH, please use WMMoveResize instead.
func (w *Window) MoveResize(x, y, width, height int) {
	w.Configure(xproto.ConfigWindowX|xproto.ConfigWindowY|
		xproto.ConfigWindowWidth|xproto.ConfigWindowHeight, x, y, width, height,
		0, 0)
}

// Move issues a ConfigureRequest for this window with the provided
// x and y positions.
// If you're trying to move a top-level window in a window manager that
// supports EWMH, please use WMMove instead.
func (w *Window) Move(x, y int) {
	w.Configure(xproto.ConfigWindowX|xproto.ConfigWindowY, x, y, 0, 0, 0, 0)
}

// Resize issues a ConfigureRequest for this window with the provided
// width and height. Note that if width or height is 0, X will stomp
// all over you. Really hard. Don't do it.
// If you're trying to resize a top-level window in a window manager that
// supports EWMH, please use WMResize instead.
func (w *Window) Resize(width, height int) {
	w.Configure(xproto.ConfigWindowWidth|xproto.ConfigWindowHeight, 0, 0,
		width, height, 0, 0)
}

// Stack issues a configure request to change the stack mode of Window.
// If you're using a window manager that supports EWMH, you may want to try
// and use ewmh.RestackWindow instead. Although this should still work.
// 'mode' values can be found as constants in xgb/xproto with the prefix
// StackMode.
// A value of xproto.StackModeAbove will put the window to the top of the stack,
// while a value of xproto.StackMoveBelow will put the window to the
// bottom of the stack.
// Remember that stacking is at the discretion of the window manager, and
// therefore may not always work as one would expect.
func (w *Window) Stack(mode byte) {
	xproto.ConfigureWindow(w.X.Conn(), w.Id,
		xproto.ConfigWindowStackMode, []uint32{uint32(mode)})
}

// StackSibling issues a configure request to change the sibling and stack mode
// of Window.
// If you're using a window manager that supports EWMH, you may want to try
// and use ewmh.RestackWindowExtra instead. Although this should still work.
// 'mode' values can be found as constants in xgb/xproto with the prefix
// StackMode.
// 'sibling' refers to the sibling window in the stacking order through which
// 'mode' is interpreted. Note that 'sibling' should be taken literally. A
// window can only be stacked with respect to a *sibling* in the window tree.
// This means that a client window that has been wrapped in decorations cannot
// be stacked with respect to another client window. (This is why you should
// use ewmh.RestackWindowExtra instead.)
func (w *Window) StackSibling(sibling xproto.Window, mode byte) {
	xproto.ConfigureWindow(w.X.Conn(), w.Id,
		xproto.ConfigWindowSibling|xproto.ConfigWindowStackMode,
		[]uint32{uint32(sibling), uint32(mode)})
}

// Map is a simple alias to map the window.
func (w *Window) Map() {
	if w == nil {
		return
	}

	xproto.MapWindow(w.X.Conn(), w.Id)
}

// Unmap is a simple alias to unmap the window.
func (w *Window) Unmap() {
	if w == nil {
		return
	}

	xproto.UnmapWindow(w.X.Conn(), w.Id)
}

// Destroy is a simple alias to destroy a window. You should use this when
// you no longer intend to use this window. (It will free the X resource
// identifier for use in other places.)
func (w *Window) Destroy() {
	if !w.Destroyed {
		w.Detach()
		err := xproto.DestroyWindowChecked(w.X.Conn(), w.Id).Check()
		if err != nil {
			xgbutil.Logger.Println(err)
		}

		w.Destroyed = true
	}
}

// Detach will detach this window's event handlers from all xevent, keybind
// and mousebind callbacks.
func (w *Window) Detach() {
	keybind.Detach(w.X, w.Id)
	mousebind.Detach(w.X, w.Id)
	xevent.Detach(w.X, w.Id)
}

// Focus tries to issue a SetInputFocus to get the focus.
// If you're trying to change the top-level active window, please use
// ewmh.ActiveWindowReq instead.
func (w *Window) Focus() {
	mode := byte(xproto.InputFocusPointerRoot)
	err := xproto.SetInputFocusChecked(w.X.Conn(), mode, w.Id, 0).Check()
	if err != nil {
		xgbutil.Logger.Println(err)
	}
}

// FocusParent is just like Focus, except it sets the "revert-to" mode to
// Parent. This should be used when setting focus to a sub-window.
func (w *Window) FocusParent(tstamp xproto.Timestamp) {
	mode := byte(xproto.InputFocusParent)
	err := xproto.SetInputFocusChecked(w.X.Conn(), mode, w.Id, tstamp).Check()
	if err != nil {
		xgbutil.Logger.Println(err)
	}
}

// Kill forcefully destroys a client. It is almost never what you want, and if
// you do it to one your clients, you'll lose your connection.
// (This is typically used in a special client like `xkill` or in a window
// manager.)
func (w *Window) Kill() {
	xproto.KillClient(w.X.Conn(), uint32(w.Id))
}

// Clear paints the region of the window specified with the corresponding
// background pixmap. If the window doesn't have a background pixmap,
// this has no effect.
// If width/height is 0, then it is set to the width/height of the background
// pixmap minus x/y.
func (w *Window) Clear(x, y, width, height int) {
	xproto.ClearArea(w.X.Conn(), false, w.Id,
		int16(x), int16(y), uint16(width), uint16(height))
}

// ClearAll is the same as Clear, but does it for the entire background pixmap.
func (w *Window) ClearAll() {
	xproto.ClearArea(w.X.Conn(), false, w.Id, 0, 0, 0, 0)
}

// Parent queries the QueryTree and finds the parent window.
func (w *Window) Parent() (*Window, error) {
	tree, err := xproto.QueryTree(w.X.Conn(), w.Id).Reply()
	if err != nil {
		return nil, fmt.Errorf("ParentWindow: Error retrieving parent window "+
			"for %x: %s", w.Id, err)
	}
	return New(w.X, tree.Parent), nil
}
