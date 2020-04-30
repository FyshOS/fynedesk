// +build linux

package win

import (
	"bytes"
	"math"

	"fyne.io/fyne"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/icccm"
	"github.com/BurntSushi/xgbutil/motif"
	"github.com/BurntSushi/xgbutil/xgraphics"
	"github.com/BurntSushi/xgbutil/xprop"

	"fyne.io/fynedesk"
	"fyne.io/fynedesk/internal/x11"
)

func windowActiveReq(x *xgbutil.XUtil, win xproto.Window) {
	ewmh.ActiveWindowReq(x, win)
}

func windowAllowedActionsSet(x *xgbutil.XUtil, win xproto.Window, actions []string) {
	ewmh.WmAllowedActionsSet(x, win, actions)
}

func windowBorderless(x *xgbutil.XUtil, win xproto.Window) bool {
	hints, err := motif.WmHintsGet(x, win)
	if err == nil {
		return !motif.Decor(hints)
	}

	return false
}

func windowClass(x *xgbutil.XUtil, win xproto.Window) []string {
	class, err := xprop.PropValStrs(xprop.GetProperty(x, win, "WM_CLASS"))
	if err != nil {
		class, err := xprop.PropValStrs(xprop.GetProperty(x, win, "_NET_WM_CLASS"))
		if err != nil {
			return []string{""}
		}
		return class
	}

	return class
}

func windowCommand(x *xgbutil.XUtil, win xproto.Window) string {
	command, err := xprop.PropValStr(xprop.GetProperty(x, win, "WM_COMMAND"))
	if err != nil {
		command, err := xprop.PropValStr(xprop.GetProperty(x, win, "_NET_WM_COMMAND"))
		if err != nil {
			return ""
		}
		return command
	}

	return command
}

func windowIcon(x *xgbutil.XUtil, win xproto.Window, width int, height int) bytes.Buffer {
	var w bytes.Buffer
	img, err := xgraphics.FindIcon(x, win, width, height)
	if err != nil {
		fyne.LogError("Could not get window icon", err)
		return w
	}
	err = img.WritePng(&w)
	if err != nil {
		fyne.LogError("Could not convert icon to png", err)
	}
	return w
}

func windowIconName(x *xgbutil.XUtil, win xproto.Window) string {
	icon, err := icccm.WmIconNameGet(x, win)
	if err != nil {
		icon, err = ewmh.WmIconNameGet(x, win)
		if err != nil {
			return ""
		}
	}

	return icon
}

func windowSizeCanMaximize(x *xgbutil.XUtil, win fynedesk.Window) bool {
	screen := fynedesk.Instance().Screens().ScreenForWindow(win)

	maxWidth, maxHeight := windowSizeMax(x, win.(x11.XWin).ChildID())
	if maxWidth == 0 && maxHeight == 0 {
		return true
	}
	if maxWidth < screen.Width || maxHeight < screen.Height {
		return false
	}
	return true
}

func windowSizeConstrain(x *xgbutil.XUtil, win xproto.Window, width uint, height uint) (uint, uint) {
	minWidth, minHeight := windowSizeMin(x, win)
	maxWidth, maxHeight := windowSizeMax(x, win)
	if width < minWidth {
		width = minWidth
	}
	if height < minHeight {
		height = minHeight
	}
	if maxWidth != 0 && width > maxWidth {
		width = maxWidth
	}
	if maxHeight != 0 && height > maxHeight {
		height = maxHeight
	}
	return width, height
}

func windowSizeFixed(x *xgbutil.XUtil, win xproto.Window) bool {
	minWidth, minHeight := windowSizeMin(x, win)
	maxWidth, maxHeight := windowSizeMax(x, win)
	if minWidth == maxWidth && minHeight == maxHeight {
		return true
	}
	return false
}

func windowSizeMax(x *xgbutil.XUtil, win xproto.Window) (uint, uint) {
	nh, err := icccm.WmNormalHintsGet(x, win)
	if err != nil {
		return 0, 0
	}
	if (nh.Flags & icccm.SizeHintPMaxSize) > 0 {
		return nh.MaxWidth, nh.MaxHeight
	}
	return 0, 0
}

func windowSizeMin(x *xgbutil.XUtil, win xproto.Window) (uint, uint) {
	nh, err := icccm.WmNormalHintsGet(x, win)
	if err != nil {
		return 0, 0
	}
	if (nh.Flags & icccm.SizeHintPMinSize) > 0 {
		return nh.MinWidth, nh.MinHeight
	}
	return 0, 0
}

func windowSizeWithIncrement(x *xgbutil.XUtil, win xproto.Window, width uint, height uint) (uint, uint) {
	nh, err := icccm.WmNormalHintsGet(x, win)
	if err != nil {
		fyne.LogError("Could not apply requested increment", err)
		return width, height
	}
	if (nh.Flags & icccm.SizeHintPResizeInc) > 0 {
		var baseWidth, baseHeight uint = 0, 0
		if nh.BaseWidth > 0 {
			baseWidth = nh.BaseWidth
		} else {
			minWidth, _ := windowSizeMin(x, win)
			baseWidth = minWidth
		}
		if nh.BaseHeight > 0 {
			baseHeight = nh.BaseHeight
		} else {
			_, minHeight := windowSizeMin(x, win)
			baseHeight = minHeight
		}
		if nh.WidthInc > 0 {
			width = baseWidth + uint((math.Round(float64(width-baseWidth)/float64(nh.WidthInc)))*float64(nh.WidthInc))
		}
		if nh.HeightInc > 0 {
			height = baseHeight + uint((math.Round(float64(height-baseHeight)/float64(nh.HeightInc)))*float64(nh.HeightInc))
		}
	}
	return width, height
}

func windowStateGet(x *xgbutil.XUtil, win xproto.Window) uint {
	state, err := icccm.WmStateGet(x, win)
	if err != nil {
		return icccm.StateNormal
	}
	return state.State
}

func windowStateSet(x *xgbutil.XUtil, win xproto.Window, state uint) {
	icccm.WmStateSet(x, win, &icccm.WmState{State: state})
}
