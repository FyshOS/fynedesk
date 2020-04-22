// +build linux

package x11

import (
	"log"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xcursor"
)

var (
	DefaultCursor           xproto.Cursor
	CloseCursor             xproto.Cursor
	ResizeBottomCursor      xproto.Cursor
	ResizeBottomLeftCursor  xproto.Cursor
	ResizeBottomRightCursor xproto.Cursor
	ResizeLeftCursor        xproto.Cursor
	ResizeRightCursor       xproto.Cursor
)

func createCursor(x *xgbutil.XUtil, builtin uint16) xproto.Cursor {
	cursor, err := xcursor.CreateCursor(x, builtin)
	if err != nil {
		log.Println("CreateCursor err", err)
		return 0
	}

	return cursor
}

func LoadCursors(x *xgbutil.XUtil) {
	DefaultCursor = createCursor(x, xcursor.LeftPtr)
	CloseCursor = createCursor(x, xcursor.XCursor)
	ResizeBottomCursor = createCursor(x, xcursor.BottomSide)
	ResizeBottomLeftCursor = createCursor(x, xcursor.BottomLeftCorner)
	ResizeBottomRightCursor = createCursor(x, xcursor.BottomRightCorner)
	ResizeLeftCursor = createCursor(x, xcursor.LeftSide)
	ResizeRightCursor = createCursor(x, xcursor.RightSide)
}
