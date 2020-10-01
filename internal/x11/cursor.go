// +build linux openbsd freebsd netbsd

package x11

import (
	"log"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xcursor"
)

var (
	// DefaultCursor is the default X11 cursor
	DefaultCursor xproto.Cursor
	// CloseCursor is the window close X11 cursor
	CloseCursor xproto.Cursor
	// ResizeBottomCursor is the bottom only resize X11 cursor
	ResizeBottomCursor xproto.Cursor
	// ResizeBottomLeftCursor is the bottom left resize X11 cursor
	ResizeBottomLeftCursor xproto.Cursor
	// ResizeBottomRightCursor is the bottom right resize X11 cursor
	ResizeBottomRightCursor xproto.Cursor
	// ResizeLeftCursor is the left resize X11 cursor
	ResizeLeftCursor xproto.Cursor
	// ResizeRightCursor is the right resize X11 cursor
	ResizeRightCursor xproto.Cursor
)

func createCursor(x *xgbutil.XUtil, builtin uint16) xproto.Cursor {
	cursor, err := xcursor.CreateCursor(x, builtin)
	if err != nil {
		log.Println("CreateCursor err", err)
		return 0
	}

	return cursor
}

// LoadCursors sets up the X11 cursors
func LoadCursors(x *xgbutil.XUtil) {
	DefaultCursor = createCursor(x, xcursor.LeftPtr)
	CloseCursor = createCursor(x, xcursor.XCursor)
	ResizeBottomCursor = createCursor(x, xcursor.BottomSide)
	ResizeBottomLeftCursor = createCursor(x, xcursor.BottomLeftCorner)
	ResizeBottomRightCursor = createCursor(x, xcursor.BottomRightCorner)
	ResizeLeftCursor = createCursor(x, xcursor.LeftSide)
	ResizeRightCursor = createCursor(x, xcursor.RightSide)
}
