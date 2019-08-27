// +build linux,!ci

package wm

import (
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/icccm"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xprop"

	"fyne.io/desktop"
	"fyne.io/fyne"
)

type clientMessageStateAction int

const (
	clientMessageStateActionRemove clientMessageStateAction = 0
	clientMessageStateActionAdd    clientMessageStateAction = 1
	clientMessageStateActionToggle clientMessageStateAction = 2
)

type client struct {
	id, win xproto.Window

	iconic    bool
	maximized bool

	frame *frame
	wm    *x11WM
}

func (s *stack) clientForWin(id xproto.Window) desktop.Window {
	for _, w := range s.clients {
		if w.(*client).id == id || w.(*client).win == id {
			return w
		}
	}

	return nil
}

func (c *client) Decorated() bool {
	if c.frame != nil {
		return c.frame.framed
	}
	return false
}

func (c *client) Title() string {
	return windowName(c.wm.x, c.win)
}

func (c *client) Class() []string {
	return windowClass(c.wm.x, c.win)
}

func (c *client) Command() string {
	return windowCommand(c.wm.x, c.win)
}

func (c *client) IconName() string {
	return windowIconName(c.wm.x, c.win)
}

func (c *client) Iconic() bool {
	return c.iconic
}

func (c *client) Maximized() bool {
	return c.maximized
}

func (c *client) TopWindow() bool {
	if c.wm.TopWindow() == c {
		return true
	}
	return false
}

func (c *client) Close() {
	winProtos, err := icccm.WmProtocolsGet(c.wm.x, c.win)
	if err != nil {
		fyne.LogError("Get Protocols Error", err)
	}

	askNicely := false
	for _, proto := range winProtos {
		if proto == "WM_DELETE_WINDOW" {
			askNicely = true
		}
	}

	if !askNicely {
		err := xproto.DestroyWindowChecked(c.wm.x.Conn(), c.win).Check()
		if err != nil {
			fyne.LogError("Close Error", err)
		}

		return
	}

	protocols, err := xprop.Atm(c.wm.x, "WM_PROTOCOLS")
	if err != nil {
		fyne.LogError("Get Protocols Error", err)
		return
	}

	delWin, err := xprop.Atm(c.wm.x, "WM_DELETE_WINDOW")
	if err != nil {
		fyne.LogError("Get Delete Window Error", err)
		return
	}
	cm, err := xevent.NewClientMessage(32, c.win, protocols,
		int(delWin))
	err = xproto.SendEventChecked(c.wm.x.Conn(), false, c.win, 0,
		string(cm.Bytes())).Check()
	if err != nil {
		fyne.LogError("Window Delete Error", err)
	}
}

func (c *client) sendStateMessage(state int) {
	stateChangeAtom, err := xprop.Atm(c.wm.x, "WM_STATE_CHANGE")
	if err != nil {
		fyne.LogError("Error getting X Atom", err)
		return
	}
	cm, err := xevent.NewClientMessage(32, c.win, stateChangeAtom,
		state)
	if err != nil {
		fyne.LogError("Error creating client message", err)
		return
	}
	err = xevent.SendRootEvent(c.wm.x, cm, xproto.EventMaskSubstructureNotify|xproto.EventMaskSubstructureRedirect)
}

func (c *client) Iconify() {
	c.sendStateMessage(icccm.StateIconic)
	windowStateSet(c.wm.x, c.win, icccm.StateIconic)
	windowExtendedHintsAdd(c.wm.x, c.win, "_NET_WM_STATE_HIDDEN")
	c.iconic = true
}

func (c *client) Uniconify() {
	c.sendStateMessage(icccm.StateNormal)
	windowStateSet(c.wm.x, c.win, icccm.StateNormal)
	windowExtendedHintsRemove(c.wm.x, c.win, "_NET_WM_STATE_HIDDEN")
	c.iconic = false
}

func (c *client) maximizeMessage(action clientMessageStateAction) {
	ewmh.WmStateReqExtra(c.wm.x, c.win, int(action), "_NET_WM_STATE_MAXIMIZED_VERT",
		"_NET_WM_STATE_MAXIMIZED_HORZ", 1)
}

func (c *client) Maximize() {
	c.maximizeMessage(clientMessageStateActionAdd)
	windowExtendedHintsAdd(c.wm.x, c.win, "_NET_WM_STATE_MAXIMIZED_VERT")
	windowExtendedHintsAdd(c.wm.x, c.win, "_NET_WM_STATE_MAXIMIZED_HORZ")
	c.maximized = true
}

func (c *client) Unmaximize() {
	c.maximizeMessage(clientMessageStateActionRemove)
	windowExtendedHintsRemove(c.wm.x, c.win, "_NET_WM_STATE_MAXIMIZED_VERT")
	windowExtendedHintsRemove(c.wm.x, c.win, "_NET_WM_STATE_MAXIMIZED_HORZ")
	c.maximized = false
}

func (c *client) Focus() {
	xproto.SetInputFocus(c.wm.x.Conn(), 0, c.win, 0)
}

func (c *client) RaiseToTop() {
	c.wm.RaiseToTop(c)
}

func (c *client) RaiseAbove(win desktop.Window) {
	topID := c.wm.rootID
	if win != nil {
		topID = win.(*client).id
	}

	c.Focus()
	if c.id == topID {
		return
	}

	c.wm.raiseWinAboveID(c.id, topID)
}

func (x *x11WM) raiseWinAboveID(win, top xproto.Window) {
	err := xproto.ConfigureWindowChecked(x.x.Conn(), win, xproto.ConfigWindowSibling|xproto.ConfigWindowStackMode,
		[]uint32{uint32(top), uint32(xproto.StackModeAbove)}).Check()
	if err != nil {
		fyne.LogError("Restack Error", err)
	}
}

func (c *client) newFrame() {
	if !windowBorderless(c.wm.x, c.win) {
		c.frame = newFrame(c)
	} else {
		c.frame = newFrameBorderless(c)
	}
}

func newClient(win xproto.Window, wm *x11WM) *client {
	c := &client{win: win, wm: wm}
	c.newFrame()
	allowedActions := []string{
		"_NET_WM_ACTION_MOVE",
		"_NET_WM_ACTION_RESIZE",
		"_NET_WM_ACTION_MINIMIZE",
		"_NET_WM_ACTION_MAXIMIZE_HORZ",
		"_NET_WM_ACTION_MAXIMIZE_VERT",
		"_NET_WM_ACTION_CLOSE",
	}
	windowAllowedActionsSet(wm.x, win, allowedActions)

	return c
}
