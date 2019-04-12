package wm

import (
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/icccm"
	"github.com/BurntSushi/xgbutil/motif"
)

func windowName(x *xgbutil.XUtil, win xproto.Window) string {
	name, err := icccm.WmNameGet(x, win)
	if err != nil {
		name, err = ewmh.WmNameGet(x, win)
		if err != nil {
			return "Noname"
		}
	}

	return name
}

func windowBorderless(x *xgbutil.XUtil, win xproto.Window) bool {
	hints, err := motif.WmHintsGet(x, win)
	if err == nil {
		return !motif.Decor(hints)
	}

	return false
}
