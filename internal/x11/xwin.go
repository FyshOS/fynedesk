// +build linux

package x11

import (
	"github.com/BurntSushi/xgb/xproto"

	"fyne.io/fynedesk"
)

type XWin interface {
	fynedesk.Window

	FrameID() xproto.Window
	ChildID() xproto.Window

	SizeMin() (uint, uint)
	SizeMax() (int, int)
	Geometry() (int, int, uint, uint)

	Expose()
	Refresh()
	SettingsChanged()

	NotifyBorderChange()
	NotifyGeometry(int, int, uint, uint)
	NotifyMoveResizeEnded()

	NotifyMaximize()
	NotifyUnMaximize()
	NotifyFullscreen()
	NotifyUnFullscreen()
	NotifyIconify()
	NotifyUnIconify()

	NotifyMouseDrag(int16, int16)
	NotifyMouseMotion(int16, int16)
	NotifyMousePress(int16, int16)
	NotifyMouseRelease(int16, int16)
}
