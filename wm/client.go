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

type client struct {
	id, win xproto.Window

	framed    bool
	iconic    bool
	maximized bool
	title     string
	class     []string
	command   string
	iconName  string

	allowedActions []string
	stateHints     []string

	fr *frame
	wm *x11WM
}

func (s *stack) clientForWin(id xproto.Window) desktop.Window {
	for _, w := range s.clients {
		if w.(*client).id == id || w.(*client).win == id {
			return w
		}
	}

	return nil
}

func (cli *client) Decorated() bool {
	return cli.framed
}

func (cli *client) Title() string {
	if cli.title == "" {
		cli.title = windowName(cli.wm.x, cli.win)
	}
	return cli.title
}

func (cli *client) Class() []string {
	if len(cli.class) == 0 {
		cli.class = windowClass(cli.wm.x, cli.win)
	}
	return cli.class
}

func (cli *client) Command() string {
	if cli.command == "" {
		cli.command = windowCommand(cli.wm.x, cli.win)
	}
	return cli.command
}

func (cli *client) IconName() string {
	if cli.iconName == "" {
		cli.iconName = windowIconName(cli.wm.x, cli.win)
	}
	return cli.iconName
}

func (cli *client) Iconic() bool {
	return cli.iconic
}

func (cli *client) Maximized() bool {
	return cli.maximized
}

func (cli *client) TopWindow() bool {
	if cli.wm.TopWindow() == cli {
		return true
	}
	return false
}

func (cli *client) Close() {
	winProtos, err := icccm.WmProtocolsGet(cli.wm.x, cli.win)
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
		err := xproto.DestroyWindowChecked(cli.wm.x.Conn(), cli.win).Check()
		if err != nil {
			fyne.LogError("Close Error", err)
		}

		return
	}

	protocols, err := xprop.Atm(cli.wm.x, "WM_PROTOCOLS")
	if err != nil {
		fyne.LogError("Get Protocols Error", err)
		return
	}

	delWin, err := xprop.Atm(cli.wm.x, "WM_DELETE_WINDOW")
	if err != nil {
		fyne.LogError("Get Delete Window Error", err)
		return
	}
	cm, err := xevent.NewClientMessage(32, cli.win, protocols,
		int(delWin))
	err = xproto.SendEventChecked(cli.wm.x.Conn(), false, cli.win, 0,
		string(cm.Bytes())).Check()
	if err != nil {
		fyne.LogError("Window Delete Error", err)
	}
}

func (cli *client) iconifyMessage(state int) {
	iconifyAtm, err := xprop.Atm(cli.wm.x, "WM_STATE_CHANGE")
	if err != nil {
		fyne.LogError("Error getting X Atom", err)
		return
	}
	cm, err := xevent.NewClientMessage(32, cli.win, iconifyAtm,
		state)
	if err != nil {
		fyne.LogError("Error creating client message", err)
		return
	}
	err = xevent.SendRootEvent(cli.wm.x, cm, xproto.EventMaskSubstructureNotify|xproto.EventMaskSubstructureRedirect)
}

func (cli *client) Iconify() {
	cli.iconifyMessage(icccm.StateIconic)
	cli.iconic = true
}

func (cli *client) Uniconify() {
	cli.iconifyMessage(icccm.StateNormal)
	cli.iconic = false
}

func (cli *client) maximizeMessage(action int) {
	ewmh.WmStateReqExtra(cli.wm.x, cli.win, action, "_NET_WM_STATE_MAXIMIZED_VERT",
		"_NET_WM_STATE_MAXIMIZED_HORZ", 1)
}

func (cli *client) Maximize() {
	cli.maximized = true
	//1 is for adding a new state
	cli.maximizeMessage(1)
}

func (cli *client) Unmaximize() {
	cli.maximized = false
	//0 is for removing a state
	cli.maximizeMessage(0)
}

func (cli *client) Focus() {
	xproto.SetInputFocus(cli.wm.x.Conn(), 0, cli.win, 0)
}

func (cli *client) RaiseToTop() {
	cli.wm.RaiseToTop(cli)
}

func (cli *client) RaiseAbove(win desktop.Window) {
	topID := cli.wm.rootID
	if win != nil {
		topID = win.(*client).id
	}

	cli.Focus()
	if cli.id == topID {
		return
	}

	cli.wm.raiseWinAboveID(cli.id, topID)
}

func (x *x11WM) raiseWinAboveID(win, top xproto.Window) {
	err := xproto.ConfigureWindowChecked(x.x.Conn(), win, xproto.ConfigWindowSibling|xproto.ConfigWindowStackMode,
		[]uint32{uint32(top), uint32(xproto.StackModeAbove)}).Check()
	if err != nil {
		fyne.LogError("Restack Error", err)
	}
}

func (cli *client) addStateHint(hint string) {
	cli.stateHints = append(cli.stateHints, hint)
	ewmh.WmStateSet(cli.wm.x, cli.win, cli.stateHints)
}

func (cli *client) removeStateHint(hint string) {
	for i, curHint := range cli.stateHints {
		if curHint == hint {
			cli.stateHints = append(cli.stateHints[:i], cli.stateHints[i+1:]...)
			ewmh.WmStateSet(cli.wm.x, cli.win, cli.stateHints)
			return
		}
	}
}

func (cli *client) maximizeFrame() {
	if cli.fr != nil {
		cli.fr.maximizeFrame()
	}
}

func (cli *client) unmaximizeFrame() {
	if cli.fr != nil {
		cli.fr.unmaximizeFrame()
	}
}

func (cli *client) newFrame() {
	if !windowBorderless(cli.wm.x, cli.win) {
		cli.framed = true
		cli.fr = newFrame(cli)
	} else {
		cli.framed = false
		cli.fr = newFrameBorderless(cli)
	}
}

func newClient(win xproto.Window, wm *x11WM) *client {
	cli := &client{win: win, wm: wm}
	cli.newFrame()

	cli.allowedActions = []string{
		"_NET_WM_ACTION_MOVE",
		"_NET_WM_ACTION_RESIZE",
		"_NET_WM_ACTION_MINIMIZE",
		"_NET_WM_ACTION_MAXIMIZE_HORZ",
		"_NET_WM_ACTION_MAXIMIZE_VERT",
		"_NET_WM_ACTION_CLOSE",
	}

	ewmh.WmAllowedActionsSet(wm.x, win, cli.allowedActions)

	return cli
}
