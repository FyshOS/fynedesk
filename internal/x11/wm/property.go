//go:build linux || openbsd || freebsd || netbsd
// +build linux openbsd freebsd netbsd

package wm

import (
	"fyne.io/fyne/v2"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/icccm"
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

func windowActiveSet(x *xgbutil.XUtil, win xproto.Window) {
	err := ewmh.ActiveWindowSet(x, win)
	if err != nil {
		fyne.LogError("", err)
	}
}

func windowClientListUpdate(wm *x11WM) {
	err := ewmh.ClientListSet(wm.X(), wm.getWindowsFromClients(wm.mappingOrder))
	if err != nil {
		fyne.LogError("", err)
	}
}

func windowClientListStackingUpdate(wm *x11WM) {
	err := ewmh.ClientListStackingSet(wm.X(), wm.getWindowsFromClients(wm.clients))
	if err != nil {
		fyne.LogError("", err)
	}
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

func windowTypeGet(x *xgbutil.XUtil, win xproto.Window) []string {
	winType, err := ewmh.WmWindowTypeGet(x, win)
	if err != nil || len(winType) == 0 {
		return []string{windowTypeNormal}
	}
	return winType
}
