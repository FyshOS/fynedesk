package xcursor

import (
	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil"
)

// CreateCursor sets some default colors for nice and easy cursor creation.
// Just supply a cursor constant from 'xcursor/cursordef.go'.
func CreateCursor(xu *xgbutil.XUtil, cursor uint16) (xproto.Cursor, error) {
	return CreateCursorExtra(xu, cursor, 0, 0, 0, 0xffff, 0xffff, 0xffff)
}

// CreateCursorExtra features all available parameters to creating a cursor.
// It will return an error if there is a problem with any of the requests
// made to create the cursor.
// (This implies each request is a checked request. The performance loss is
// probably acceptable since cursors should be created once and reused.)
func CreateCursorExtra(xu *xgbutil.XUtil, cursor, foreRed, foreGreen,
	foreBlue, backRed, backGreen, backBlue uint16) (xproto.Cursor, error) {

	fontId, err := xproto.NewFontId(xu.Conn())
	if err != nil {
		return 0, err
	}

	cursorId, err := xproto.NewCursorId(xu.Conn())
	if err != nil {
		return 0, err
	}

	err = xproto.OpenFontChecked(xu.Conn(), fontId,
		uint16(len("cursor")), "cursor").Check()
	if err != nil {
		return 0, err
	}

	err = xproto.CreateGlyphCursorChecked(xu.Conn(), cursorId, fontId, fontId,
		cursor, cursor+1,
		foreRed, foreGreen, foreBlue,
		backRed, backGreen, backBlue).Check()
	if err != nil {
		return 0, err
	}

	err = xproto.CloseFontChecked(xu.Conn(), fontId).Check()
	if err != nil {
		return 0, err
	}

	return cursorId, nil
}
