// +build linux

package wm

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
)

const (
	windowTypeDesktop      = "_NET_WM_WINDOW_TYPE_DESKTOP"
	windowTypeDock         = "_NET_WM_WINDOW_TYPE_DOCK"
	windowTypeToolbar      = "_NET_WM_WINDOW_TYPE_TOOLBAR"
	windowTypeMenu         = "_NET_WM_WINDOW_TYPE_MENU"
	windowTypeUtility      = "_NET_WM_WINDOW_TYPE_UTILITY"
	windowTypeSplash       = "_NET_WM_WINDOW_TYPE_SPLASH"
	windowTypeDialog       = "_NET_WM_WINDOW_TYPE_DIALOG"
	windowTypeDropdownMenu = "_NET_WM_WINDOW_TYPE_DROPDOWN_MENU"
	windowTypePopupMenu    = "_NET_WM_WINDOW_TYPE_POPUP_MENU"
	windowTypeTooltip      = "_NET_WM_WINDOW_TYPE_TOOLTIP"
	windowTypeNotification = "_NET_WM_WINDOW_TYPE_NOTIFICATION"
	windowTypeCombo        = "_NET_WM_WINDOW_TYPE_COMBO"
	windowTypeDND          = "_NET_WM_WINDOW_TYPE_DND"
	windowTypeNormal       = "_NET_WM_WINDOW_TYPE_NORMAL"
)

func windowName(x *xgbutil.XUtil, win xproto.Window) string {
	//Spec says _NET_WM_NAME is preferred to WM_NAME
	name, err := ewmh.WmNameGet(x, win)
	if err != nil {
		name, err = icccm.WmNameGet(x, win)
		if err != nil {
			return ""
		}
	}

	return name
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

func windowBorderless(x *xgbutil.XUtil, win xproto.Window) bool {
	hints, err := motif.WmHintsGet(x, win)
	if err == nil {
		return !motif.Decor(hints)
	}

	return false
}

func windowMinSize(x *xgbutil.XUtil, win xproto.Window) (uint, uint) {
	hints, err := icccm.WmNormalHintsGet(x, win)
	if err == nil {
		return hints.MinWidth, hints.MinHeight
	}

	return 0, 0
}

func windowSizeWithIncrement(x *xgbutil.XUtil, win xproto.Window, width uint16, height uint16) (uint16, uint16) {
	nh, err := icccm.WmNormalHintsGet(x, win)
	if err != nil {
		fyne.LogError("Could not apply requested increment", err)
		return width, height
	}
	if (nh.Flags & icccm.SizeHintPResizeInc) > 0 {
		var baseWidth, baseHeight uint16
		if nh.BaseWidth > 0 {
			baseWidth = uint16(nh.BaseWidth)
		} else {
			baseWidth = uint16(nh.MinWidth)
		}
		if nh.BaseHeight > 0 {
			baseHeight = uint16(nh.BaseHeight)
		} else {
			baseHeight = uint16(nh.MinHeight)
		}
		if nh.WidthInc > 0 {
			width = baseWidth + (uint16(math.Round((float64(width-baseWidth) / float64(nh.WidthInc)) * float64(nh.WidthInc))))
		}
		if nh.HeightInc > 0 {
			height = baseHeight + (uint16(math.Round((float64(height-baseHeight) / float64(nh.HeightInc)) * float64(nh.HeightInc))))
		}
	}
	return width, height
}

func windowAllowedActionsSet(x *xgbutil.XUtil, win xproto.Window, actions []string) {
	ewmh.WmAllowedActionsSet(x, win, actions)
}

func windowStateSet(x *xgbutil.XUtil, win xproto.Window, state uint) {
	icccm.WmStateSet(x, win, &icccm.WmState{State: state})
}

func windowStateGet(x *xgbutil.XUtil, win xproto.Window) uint {
	state, err := icccm.WmStateGet(x, win)
	if err != nil {
		return icccm.StateNormal
	}
	return state.State
}

func windowTransientForGet(x *xgbutil.XUtil, win xproto.Window) xproto.Window {
	transient, err := icccm.WmTransientForGet(x, win)
	if err != nil {
		return 0
	}
	return transient
}

func windowOverrideGet(x *xgbutil.XUtil, win xproto.Window) bool {
	hints, err := icccm.WmHintsGet(x, win)
	if err == nil && (hints.Flags&xproto.CwOverrideRedirect) != 0 {
		return true
	}
	attrs, err := xproto.GetWindowAttributes(x.Conn(), win).Reply()
	if err == nil && attrs.OverrideRedirect {
		return true
	}
	return false
}

func windowActiveReq(x *xgbutil.XUtil, win xproto.Window) {
	ewmh.ActiveWindowReq(x, win)
}

func windowActiveSet(x *xgbutil.XUtil, win xproto.Window) {
	ewmh.ActiveWindowSet(x, win)
}

func windowActiveGet(x *xgbutil.XUtil) (xproto.Window, error) {
	return ewmh.ActiveWindowGet(x)
}

func windowExtendedHintsGet(x *xgbutil.XUtil, win xproto.Window) []string {
	extendedHints, err := ewmh.WmStateGet(x, win)
	if err != nil {
		return nil
	}
	return extendedHints
}

func windowExtendedHintsAdd(x *xgbutil.XUtil, win xproto.Window, hint string) {
	extendedHints, _ := ewmh.WmStateGet(x, win) // error unimportant
	extendedHints = append(extendedHints, hint)
	ewmh.WmStateSet(x, win, extendedHints)
}

func windowExtendedHintsRemove(x *xgbutil.XUtil, win xproto.Window, hint string) {
	extendedHints, err := ewmh.WmStateGet(x, win)
	if err != nil {
		return
	}
	for i, curHint := range extendedHints {
		if curHint == hint {
			extendedHints = append(extendedHints[:i], extendedHints[i+1:]...)
			ewmh.WmStateSet(x, win, extendedHints)
			return
		}
	}
}

func windowTypeGet(x *xgbutil.XUtil, win xproto.Window) []string {
	winType, err := ewmh.WmWindowTypeGet(x, win)
	if err != nil || len(winType) == 0 {
		return []string{windowTypeNormal}
	}
	return winType
}

func windowClientListUpdate(wm *x11WM) {
	ewmh.ClientListSet(wm.x, wm.getWindowsFromClients(wm.mappingOrder))
}

func windowClientListStackingUpdate(wm *x11WM) {
	ewmh.ClientListStackingSet(wm.x, wm.getWindowsFromClients(wm.clients))
}
