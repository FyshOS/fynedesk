// +build linux,!ci

package wm

import (
	"log"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xcursor"
)

var (
	defaultCursor xproto.Cursor
	closeCursor   xproto.Cursor
	resizeCursor  xproto.Cursor
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
	resizeCursor = createCursor(x, xcursor.BottomRightCorner)
}
