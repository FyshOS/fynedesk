// +build linux

package wm

import (
	"log"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xcursor"
)

var (
	defaultCursor           xproto.Cursor
	closeCursor             xproto.Cursor
	resizeBottomCursor      xproto.Cursor
	resizeBottomLeftCursor  xproto.Cursor
	resizeBottomRightCursor xproto.Cursor
	resizeLeftCursor        xproto.Cursor
	resizeRightCursor       xproto.Cursor
)

func createCursor(x *xgbutil.XUtil, builtin uint16) xproto.Cursor {
	cursor, err := xcursor.CreateCursor(x, builtin)
	if err != nil {
		log.Println("CreateCursor err", err)
		return 0
	}

	return cursor
}

func loadCursors(x *xgbutil.XUtil) {
	defaultCursor = createCursor(x, xcursor.LeftPtr)
	closeCursor = createCursor(x, xcursor.XCursor)
	resizeBottomCursor = createCursor(x, xcursor.BottomSide)
	resizeBottomLeftCursor = createCursor(x, xcursor.BottomLeftCorner)
	resizeBottomRightCursor = createCursor(x, xcursor.BottomRightCorner)
	resizeLeftCursor = createCursor(x, xcursor.LeftSide)
	resizeRightCursor = createCursor(x, xcursor.RightSide)
}
