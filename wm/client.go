// +build linux,!ci

package wm

import "C"
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

	full      bool
	iconic    bool
	maximized bool

	restoreX, restoreY          int16
	restoreWidth, restoreHeight uint16

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
	return !windowBorderless(c.wm.x, c.win)
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

func (c *client) Fullscreened() bool {
	return c.full
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

func (c *client) SkipTaskbar() bool {
	extendedHints := windowExtendedHintsGet(c.wm.x, c.win)
	if extendedHints == nil {
		return false
	}
	for _, hint := range extendedHints {
		if hint == "_NET_WM_STATE_SKIP_TASKBAR" {
			return true
		}
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

func (c *client) fullscreenMessage(action clientMessageStateAction) {
	ewmh.WmStateReq(c.wm.x, c.win, int(action), "_NET_WM_STATE_FULLSCREEN")
}

func (c *client) Fullscreen() {
	c.maximizeMessage(clientMessageStateActionAdd)
	windowExtendedHintsAdd(c.wm.x, c.win, "_NET_WM_STATE_FULLSCREEN")
}

func (c *client) Unfullscreen() {
	c.maximizeMessage(clientMessageStateActionRemove)
	windowExtendedHintsRemove(c.wm.x, c.win, "_NET_WM_STATE_FULLSCREEN")
}

func (c *client) sendStateMessage(state int) {
	stateChangeAtom, err := xprop.Atm(c.wm.x, "WM_CHANGE_STATE")
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
}

func (c *client) Uniconify() {
	c.sendStateMessage(icccm.StateNormal)
	windowStateSet(c.wm.x, c.win, icccm.StateNormal)
}

func (c *client) maximizeMessage(action clientMessageStateAction) {
	ewmh.WmStateReqExtra(c.wm.x, c.win, int(action), "_NET_WM_STATE_MAXIMIZED_VERT",
		"_NET_WM_STATE_MAXIMIZED_HORZ", 1)
}

func (c *client) Maximize() {
	c.maximizeMessage(clientMessageStateActionAdd)
}

func (c *client) Unmaximize() {
	c.maximizeMessage(clientMessageStateActionRemove)
}

func (c *client) Focus() {
	windowActiveReq(c.wm.x, c.win)
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

func (c *client) fullscreenClient() {
	c.full = true
	c.frame.changeBorder()
	c.frame.maximizeApply()
}

func (c *client) unfullscreenClient() {
	c.full = false
	c.frame.changeBorder()
	c.frame.unmaximizeApply()
}

func (c *client) iconifyClient() {
	c.frame.hide()
	c.iconic = true
	windowExtendedHintsAdd(c.wm.x, c.win, "_NET_WM_STATE_HIDDEN")
}

func (c *client) uniconifyClient() {
	c.newFrame()
	if c.frame == nil {
		return
	}

	c.iconic = false
	c.frame.show()
	windowExtendedHintsRemove(c.wm.x, c.win, "_NET_WM_STATE_HIDDEN")
}

func (c *client) maximizeClient() {
	c.maximized = true
	c.frame.maximizeApply()
	windowExtendedHintsAdd(c.wm.x, c.win, "_NET_WM_STATE_MAXIMIZED_VERT")
	windowExtendedHintsAdd(c.wm.x, c.win, "_NET_WM_STATE_MAXIMIZED_HORZ")
}

func (c *client) unmaximizeClient() {
	c.maximized = false
	c.frame.unmaximizeApply()
	windowExtendedHintsRemove(c.wm.x, c.win, "_NET_WM_STATE_MAXIMIZED_VERT")
	windowExtendedHintsRemove(c.wm.x, c.win, "_NET_WM_STATE_MAXIMIZED_HORZ")
}

func (c *client) updateTitle() {
	if c.frame == nil {
		return
	}
	c.frame.updateTitle()
}

func (c *client) setWindowGeometry(x int16, y int16, width uint16, height uint16) {
	c.frame.updateGeometry(x, y, width, height, true)
}

func (c *client) getWindowGeometry() (int16, int16, uint16, uint16) {
	return c.frame.getGeometry()
}

func (c *client) changeBorder() {
	c.frame.changeBorder()
}

func (c *client) newFrame() {
	c.frame = newFrame(c)
}

func newClient(win xproto.Window, wm *x11WM) *client {
	c := &client{win: win, wm: wm}
	err := xproto.ChangeWindowAttributesChecked(wm.x.Conn(), win, xproto.CwEventMask, []uint32{xproto.EventMaskPropertyChange}).Check()
	if err != nil {
		fyne.LogError("Could not change window attributes", err)
	}
	windowAllowedActionsSet(wm.x, win, wm.allowedActions)
	initialHints := windowExtendedHintsGet(c.wm.x, c.win)
	for _, hint := range initialHints {
		switch hint {
		case "_NET_WM_STATE_FULLSCREEN":
			c.full = true
		}
		// TODO Handle more of these possible hints
	}
	c.newFrame()

	return c
}
