// +build linux,!ci

package wm // import "fyne.io/desktop/wm"

import (
	"errors"
	"log"

	"fyne.io/desktop"
	"fyne.io/desktop/theme"
	"fyne.io/fyne"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/icccm"
	"github.com/BurntSushi/xgbutil/xprop"
)

type x11WM struct {
	stack
	x      *xgbutil.XUtil
	root   fyne.Window
	rootID xproto.Window
	loaded bool
}

func (x *x11WM) Close() {
	log.Println("Disconnecting from X server")

	for _, child := range x.clients {
		child.(*client).frame.unFrame()
	}

	x.x.Conn().Close()
}

func (x *x11WM) AddStackListener(l desktop.StackListener) {
	x.stack.listeners = append(x.stack.listeners, l)
}

func (x *x11WM) SetRoot(win fyne.Window) {
	x.root = win
}

// NewX11WindowManager sets up a new X11 Window Manager to control a desktop in X11.
func NewX11WindowManager(a fyne.App) (desktop.WindowManager, error) {
	conn, err := xgbutil.NewConn()
	if err != nil {
		fyne.LogError("Failed to connect to the XServer", err)
		return nil, err
	}

	mgr := &x11WM{x: conn}
	root := conn.RootWin()
	eventMask := xproto.EventMaskPropertyChange |
		xproto.EventMaskFocusChange |
		xproto.EventMaskButtonPress |
		xproto.EventMaskButtonRelease |
		xproto.EventMaskKeyPress |
		xproto.EventMaskVisibilityChange |
		xproto.EventMaskStructureNotify |
		xproto.EventMaskSubstructureNotify |
		xproto.EventMaskSubstructureRedirect
	if err := xproto.ChangeWindowAttributesChecked(conn.Conn(), root, xproto.CwEventMask,
		[]uint32{uint32(eventMask)}).Check(); err != nil {
		conn.Conn().Close()

		return nil, errors.New("window manager detected, running embedded")
	}

	loadCursors(conn)
	go mgr.runLoop()

	listener := make(chan fyne.Settings)
	a.Settings().AddChangeListener(listener)
	go func() {
		for {
			<-listener
			for _, c := range mgr.clients {
				c.(*client).frame.applyTheme()
			}
		}
	}()

	return mgr, nil
}

func (x *x11WM) runLoop() {
	conn := x.x.Conn()

	for {
		ev, err := conn.WaitForEvent()
		if err != nil {
			fyne.LogError("X11 Error:", err)
			continue
		}

		if ev != nil {
			switch ev := ev.(type) {
			case xproto.MapRequestEvent:
				x.showWindow(ev.Window)
			case xproto.UnmapNotifyEvent:
				x.hideWindow(ev.Window)
			case xproto.ConfigureRequestEvent:
				x.configureWindow(ev.Window, ev)
			case xproto.DestroyNotifyEvent:
				x.destroyWindow(ev.Window)
			case xproto.PropertyNotifyEvent:
				// TODO
			case xproto.ClientMessageEvent:
				x.handleClientMessage(ev)
			case xproto.ExposeEvent:
				border := x.clientForWin(ev.Window)
				if border != nil {
					border.(*client).frame.applyTheme()
				}
			case xproto.ButtonPressEvent:
				for _, c := range x.clients {
					if c.(*client).id == ev.Event {
						c.(*client).frame.press(ev.EventX, ev.EventY)
					}
				}
			case xproto.ButtonReleaseEvent:
				for _, c := range x.clients {
					if c.(*client).id == ev.Event {
						c.(*client).frame.release(ev.EventX, ev.EventY)
					}
				}
			case xproto.MotionNotifyEvent:
				for _, c := range x.clients {
					if c.(*client).id == ev.Event {
						if ev.State&xproto.ButtonMask1 != 0 {
							c.(*client).frame.drag(ev.EventX, ev.EventY)
						} else {
							c.(*client).frame.motion(ev.EventX, ev.EventY)
						}
					}
				}
			case xproto.KeyPressEvent:
				winList := x.Windows()
				winCount := len(winList)
				if winCount <= 1 {
					return
				}

				if ev.State&xproto.ModMaskShift != 0 {
					x.RaiseToTop(winList[winCount-1])
				} else {
					x.RaiseToTop(winList[1])
				}
			}
		}
	}
}

func (x *x11WM) configureWindow(win xproto.Window, ev xproto.ConfigureRequestEvent) {
	c := x.clientForWin(win)
	width := ev.Width
	height := ev.Height

	if c != nil {
		f := c.(*client).frame
		if f != nil {
			f.minWidth, f.minHeight = windowMinSize(x.x, win)
			if c.Decorated() {
				err := xproto.ConfigureWindowChecked(x.x.Conn(), win, xproto.ConfigWindowX|xproto.ConfigWindowY|
					xproto.ConfigWindowWidth|xproto.ConfigWindowHeight,
					[]uint32{uint32(x.borderWidth()), uint32(x.borderWidth() + x.titleHeight()),
						uint32(width - 1), uint32(height - 1)}).Check()

				if err != nil {
					fyne.LogError("Configure Frame Error", err)
				}
			}
		}
		return
	}

	name := windowName(x.x, win)
	isRoot := x.root != nil && name == x.root.Title()
	if isRoot {
		width = x.x.Screen().WidthInPixels
		height = x.x.Screen().HeightInPixels
	}

	err := xproto.ConfigureWindowChecked(x.x.Conn(), win, xproto.ConfigWindowX|xproto.ConfigWindowY|
		xproto.ConfigWindowWidth|xproto.ConfigWindowHeight,
		[]uint32{uint32(ev.X), uint32(ev.Y), uint32(width), uint32(height)}).Check()
	if err != nil {
		fyne.LogError("Configure Window Error", err)
	}

	if isRoot {
		if x.loaded {
			return
		}
		x.rootID = win
		x.loaded = true
	}
}

func (x *x11WM) handleStateActionRequest(ev xproto.ClientMessageEvent, removeState func(xproto.Window), addState func(xproto.Window), toggleCheck bool) {
	switch clientMessageStateAction(ev.Data.Data32[0]) {
	case clientMessageStateActionRemove:
		removeState(ev.Window)
	case clientMessageStateActionAdd:
		addState(ev.Window)
	case clientMessageStateActionToggle:
		if toggleCheck {
			removeState(ev.Window)
		} else {
			addState(ev.Window)
		}
	}
}

func (x *x11WM) handleClientMessage(ev xproto.ClientMessageEvent) {
	msgAtom, err := xprop.AtomName(x.x, ev.Type)
	if err != nil {
		fyne.LogError("Error getting event", err)
		return
	}
	switch msgAtom {
	case "WM_STATE_CHANGE":
		switch ev.Data.Bytes()[0] {
		case icccm.StateIconic:
			x.iconifyWindow(ev.Window)
		case icccm.StateNormal:
			x.uniconifyWindow(ev.Window)
		}
	case "_NET_WM_STATE":
		subMsgAtom, err := xprop.AtomName(x.x, xproto.Atom(ev.Data.Data32[1]))
		if err != nil {
			fyne.LogError("Error getting event", err)
			return
		}
		c := x.clientForWin(ev.Window)
		if c == nil {
			fyne.LogError("Could not retrieve client", nil)
			return
		}
		switch subMsgAtom {
		case "_NET_WM_STATE_HIDDEN":
			x.handleStateActionRequest(ev, x.uniconifyWindow, x.iconifyWindow, c.Iconic())
		case "_NET_WM_STATE_MAXIMIZED_VERT", "_NET_WM_STATE_MAXIMIZED_HORZ":
			extraMsgAtom, err := xprop.AtomName(x.x, xproto.Atom(ev.Data.Data32[2]))
			if err != nil {
				fyne.LogError("Error getting event", err)
				return
			}
			if extraMsgAtom == subMsgAtom {
				return
			}
			if extraMsgAtom == "_NET_WM_STATE_MAXIMIZED_VERT" || extraMsgAtom == "_NET_WM_STATE_MAXIMIZED_HORZ" {
				x.handleStateActionRequest(ev, x.unmaximizeWindow, x.maximizeWindow, c.Maximized())
			}
		}
	}
}

func (x *x11WM) iconifyWindow(win xproto.Window) {
	c := x.clientForWin(win)
	if c == nil {
		fyne.LogError("Could not retrieve client", nil)
		return
	}
	xproto.ReparentWindow(x.x.Conn(), win, x.x.RootWin(), c.(*client).frame.x, c.(*client).frame.y)
	xproto.UnmapWindow(x.x.Conn(), win)
}

func (x *x11WM) uniconifyWindow(win xproto.Window) {
	c := x.clientForWin(win)
	if c == nil {
		fyne.LogError("Could not retrieve client", nil)
		return
	}
	c.(*client).newFrame()
	xproto.MapWindow(x.x.Conn(), win)
}

func (x *x11WM) maximizeWindow(win xproto.Window) {
	c := x.clientForWin(win)
	if c == nil {
		fyne.LogError("Could not retrieve client", nil)
		return
	}
	c.(*client).frame.maximize()
}

func (x *x11WM) unmaximizeWindow(win xproto.Window) {
	c := x.clientForWin(win)
	if c == nil {
		fyne.LogError("Could not retrieve client", nil)
		return
	}
	c.(*client).frame.unmaximize()
}

func (x *x11WM) showWindow(win xproto.Window) {
	c := x.clientForWin(win)
	name := windowName(x.x, win)

	if c != nil || name == x.root.Title() {
		err := xproto.MapWindowChecked(x.x.Conn(), win).Check()
		if err != nil {
			fyne.LogError("Show Window Error", err)
		}

		if name != x.root.Title() {
			return
		}

		x.bindKeys(win)
		go x.frameExisting()

		return
	}
	if x.rootID == 0 {
		return
	}

	x.setupWindow(win)
}

func (x *x11WM) hideWindow(win xproto.Window) {
	c := x.clientForWin(win)
	if c == nil {
		fyne.LogError("Could not retrieve client", nil)
		return
	}
	xproto.UnmapWindow(x.x.Conn(), c.(*client).id)
}

func (x *x11WM) setupWindow(win xproto.Window) {
	c := x.clientForWin(win)
	if c != nil {
		c.(*client).newFrame()
	} else {
		c = newClient(win, x)
	}

	x.bindKeys(win)
	if x.root != nil && windowName(x.x, win) == x.root.Title() {
		return
	}
	x.AddWindow(c)
	x.RaiseToTop(c)
}

func (x *x11WM) destroyWindow(win xproto.Window) {
	xproto.ChangeSaveSet(x.x.Conn(), xproto.SetModeDelete, win)
	c := x.clientForWin(win)
	if c != nil {
		x.RemoveWindow(c)
	}
}

func (x *x11WM) bindKeys(win xproto.Window) {
	xproto.GrabKey(x.x.Conn(), true, win, xproto.ModMask1, 23, xproto.GrabModeAsync, xproto.GrabModeAsync)
	xproto.GrabKey(x.x.Conn(), true, win, xproto.ModMaskShift|xproto.ModMask1, 23, xproto.GrabModeAsync, xproto.GrabModeAsync)
}

func (x *x11WM) frameExisting() {
	tree, err := xproto.QueryTree(x.x.Conn(), x.x.RootWin()).Reply()
	if err != nil {
		fyne.LogError("Query Tree Error", err)
		return
	}

	for _, child := range tree.Children {
		name := windowName(x.x, child)
		if x.root != nil && name == x.root.Title() {
			continue
		}

		attrs, err := xproto.GetWindowAttributes(x.x.Conn(), child).Reply()
		if err != nil {
			fyne.LogError("Get Window Attributes Error", err)
			continue
		}
		if attrs.MapState == xproto.MapStateUnmapped {
			continue
		}

		x.setupWindow(child)
	}
}

func (x *x11WM) borderWidth() uint16 {
	scale := float32(1.0)
	if x.root != nil {
		scale = x.root.Canvas().Scale()
	}
	return uint16(float32(theme.BorderWidth) * scale)
}

func (x *x11WM) buttonWidth() uint16 {
	scale := float32(1.0)
	if x.root != nil {
		scale = x.root.Canvas().Scale()
	}
	return uint16(float32(theme.ButtonWidth) * scale)
}

func (x *x11WM) titleHeight() uint16 {
	scale := float32(1.0)
	if x.root != nil {
		scale = x.root.Canvas().Scale()
	}
	return uint16(float32(theme.TitleHeight) * scale)
}
