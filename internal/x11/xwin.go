// +build linux

package x11

import (
	"github.com/BurntSushi/xgb/xproto"

	"fyne.io/fynedesk"
)

// XWin describes the additional functions that X windows need to expose to be managed
type XWin interface {
	fynedesk.Window

	FrameID() xproto.Window
	ChildID() xproto.Window

	SizeMin() (uint, uint)
	SizeMax() (uint, uint)

	Expose()
	Refresh()
	SettingsChanged()

	NotifyBorderChange()
	NotifyGeometry(geometry fynedesk.Geometry)
	NotifyMoveResizeEnded()

	NotifyMaximize()
	NotifyUnMaximize()
	NotifyFullscreen()
	NotifyUnFullscreen()
	NotifyIconify()
	NotifyUnIconify()

	NotifyMouseDrag(int, int)
	NotifyMouseMotion(int, int)
	NotifyMousePress(int, int)
	NotifyMouseRelease(int, int)
}
