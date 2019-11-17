// +build linux,!ci

package wm

import (
	"bytes"
	"fyne.io/fyne"
	"fyne.io/fyne/theme"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/icccm"
	"github.com/BurntSushi/xgbutil/motif"
	"github.com/BurntSushi/xgbutil/xgraphics"
	"github.com/BurntSushi/xgbutil/xprop"
	"image"
	"image/color"
	"image/draw"
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
	red, blue, green, alpha := theme.BackgroundColor().RGBA()
	col := color.RGBA{R: uint8(red), G: uint8(green), B: uint8(blue), A: uint8(alpha)}
	sub := image.NewRGBA(image.Rect(0, 0, img.Rect.Dx(), img.Rect.Dy()))
	draw.Draw(sub, sub.Bounds(), &image.Uniform{C: col}, image.ZP,  draw.Src)
	xgraphics.Blend(img.SubImage(sub.Rect).(*xgraphics.Image), img, image.ZP)
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
