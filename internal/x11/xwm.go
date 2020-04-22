// +build linux

package x11 // import "fyne.io/fynedesk/internal/x11"

import (
	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"

	"fyne.io/fynedesk"
)

type XWM interface {
	fynedesk.WindowManager

	X() *xgbutil.XUtil
	Conn() *xgb.Conn

	WinIDForScreen(screen *fynedesk.Screen) xproto.Window
	BindKeys(win XWin)
}

func XConn() *xgbutil.XUtil {
	return fynedesk.Instance().WindowManager().(XWM).X()
}
