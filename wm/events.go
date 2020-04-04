// +build linux

package wm

import (
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil/icccm"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xprop"

	"fyne.io/desktop"
	"fyne.io/desktop/internal/notify"
	"fyne.io/desktop/internal/ui"

	"fyne.io/fyne"
)

func (x *x11WM) handleActiveWin(ev xproto.ClientMessageEvent) {
	canFocus := true
	notifyFocus := false
	hints, err := icccm.WmHintsGet(x.x, ev.Window)
	if err == nil {
		if hints.Flags&icccm.HintInput > 0 {
			canFocus = hints.Input > 0
		}
	}
	protocols, err := icccm.WmProtocolsGet(x.x, ev.Window)
	if err == nil {
		for _, protocol := range protocols {
			if protocol == "WM_TAKE_FOCUS" {
				notifyFocus = true
				break
			}
		}
	}
	if !canFocus && !notifyFocus {
		return
	}
	windowActiveSet(x.x, ev.Window)
	if canFocus {
		err = xproto.SetInputFocusChecked(x.x.Conn(), 2, ev.Window, xproto.TimeCurrentTime).Check()
		if err != nil {
			fyne.LogError("Could not set focus", err)
			return
		}
	}
	if notifyFocus {
		protocolAtm, err := xprop.Atm(x.x, "WM_PROTOCOLS")
		if err != nil {
			fyne.LogError("Could not get protocol atom", err)
			return
		}
		focusAtm, err := xprop.Atm(x.x, "WM_TAKE_FOCUS")
		if err != nil {
			fyne.LogError("Could not get focus atom", err)
			return
		}
		cm, err := xevent.NewClientMessage(32, ev.Window, protocolAtm, int(focusAtm), xproto.TimeCurrentTime)
		if err != nil {
			fyne.LogError("Could not create client focus message", err)
			return
		}
		xproto.SendEvent(x.x.Conn(), false, ev.Window, 0, string(cm.Bytes()))
	}
}

func (x *x11WM) handleButtonPress(ev xproto.ButtonPressEvent) {
	for _, c := range x.clients {
		if c.(*client).id == ev.Event {
			c.(*client).frame.mousePress(ev.RootX, ev.RootY)
		}
	}
	xevent.ReplayPointer(x.x)
}

func (x *x11WM) handleButtonRelease(ev xproto.ButtonReleaseEvent) {
	for _, c := range x.clients {
		if c.(*client).id == ev.Event {
			if !x.moveResizing {
				c.(*client).frame.mouseRelease(ev.RootX, ev.RootY)
			}
			x.moveResizeEnd(c.(*client))
		}
	}
}

func (x *x11WM) handleClientMessage(ev xproto.ClientMessageEvent) {
	c := x.clientForWin(ev.Window)
	msgAtom, err := xprop.AtomName(x.x, ev.Type)
	if err != nil {
		fyne.LogError("Error getting event", err)
		return
	}
	switch msgAtom {
	case "WM_CHANGE_STATE":
		if c == nil {
			return
		}
		switch ev.Data.Bytes()[0] {
		case icccm.StateIconic:
			c.(*client).iconifyClient()
		case icccm.StateNormal:
			c.(*client).uniconifyClient()
		}
	case "_NET_ACTIVE_WINDOW":
		x.handleActiveWin(ev)
	case "_NET_WM_FULLSCREEN_MONITORS":
		// TODO WHEN WE SUPPORT MULTI-MONITORS - THIS TELLS WHICH/HOW MANY MONITORS
		// TO FULLSCREEN ACROSS
	case "_NET_WM_MOVERESIZE":
		if c == nil {
			return
		}
		if c.Maximized() || c.Fullscreened() {
			return
		}
		x.handleMoveResize(ev, c.(*client))
	case "_NET_WM_STATE":
		subMsgAtom, err := xprop.AtomName(x.x, xproto.Atom(ev.Data.Data32[1]))
		if err != nil {
			fyne.LogError("Error getting event", err)
			return
		}
		if c == nil {
			x.handleInitialHints(ev, subMsgAtom)
			return
		}
		switch subMsgAtom {
		case "_NET_WM_STATE_FULLSCREEN":
			x.handleStateActionRequest(ev, c.(*client).unfullscreenClient, c.(*client).fullscreenClient, c.Fullscreened())
		case "_NET_WM_STATE_HIDDEN":
			fyne.LogError("Extended Window Manager Hints says to ignore the HIDDEN state.", nil)
		//	x.handleStateActionRequest(ev, c.(*client).uniconifyClient, c.(*client).iconifyClient, c.Iconic())
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
				x.handleStateActionRequest(ev, c.(*client).unmaximizeClient, c.(*client).maximizeClient, c.Maximized())
			}
		}
	}
}

func (x *x11WM) handleFocus(win xproto.Window) {
	c := x.clientForWin(win)
	if c == nil {
		return
	}
	c.(*client).frame.applyTheme(true)
}

func (x *x11WM) handleInitialHints(ev xproto.ClientMessageEvent, hint string) {
	switch clientMessageStateAction(ev.Data.Data32[0]) {
	case clientMessageStateActionRemove:
		windowExtendedHintsRemove(x.x, ev.Window, hint)
	case clientMessageStateActionAdd:
		windowExtendedHintsAdd(x.x, ev.Window, hint)
	}
}

func (x *x11WM) handleKeyPress(ev xproto.KeyPressEvent) {
	if ev.Detail == keyCodeSpace {
		if switcherInstance != nil { // we are currently switching windows - select current window
			x.applyAppSwitcher()
		} else {
			go ui.ShowAppLauncher()
		}
	} else {
		// The rest of these methods are about app switcher.
		// Apart from Tab they will only be called once the keyboard grab is in effect.
		if ev.Detail == keyCodeTab {
			shiftPressed := ev.State&xproto.ModMaskShift != 0
			x.showOrSelectAppSwitcher(shiftPressed)
		} else if ev.Detail == keyCodeEscape {
			x.cancelAppSwitcher()
		} else if ev.Detail == keyCodeReturn || ev.Detail == keyCodeEnter {
			x.applyAppSwitcher()
		} else if ev.Detail == keyCodeLeft {
			x.previousAppSwitcher()
		} else if ev.Detail == keyCodeRight {
			x.nextAppSwitcher()
		}
	}
}

func (x *x11WM) handleKeyRelease(ev xproto.KeyReleaseEvent) {
	if ev.Detail == keyCodeAlt {
		x.applyAppSwitcher()
	}
}

func (x *x11WM) handleMouseEnter(ev xproto.EnterNotifyEvent) {
	err := xproto.ChangeWindowAttributesChecked(x.x.Conn(), ev.Event, xproto.CwCursor,
		[]uint32{uint32(defaultCursor)}).Check()
	if err != nil {
		fyne.LogError("Set Cursor Error", err)
	}
	if mouseNotify, ok := desktop.Instance().(notify.MouseNotify); ok {
		mouseNotify.MouseOutNotify()
	}
}

func (x *x11WM) handleMouseLeave(ev xproto.LeaveNotifyEvent) {
	if mouseNotify, ok := desktop.Instance().(notify.MouseNotify); ok {
		screen := desktop.Instance().Screens().ScreenForGeometry(int(ev.RootX), int(ev.RootY), 0, 0)
		mouseNotify.MouseInNotify(fyne.NewPos(int(float32(ev.RootX)/screen.CanvasScale()),
			int(float32(ev.RootY)/screen.CanvasScale())))
	}
}

func (x *x11WM) handleMouseMotion(ev xproto.MotionNotifyEvent) {
	for _, c := range x.clients {
		if c.(*client).id == ev.Event {
			if x.moveResizing {
				x.moveResize(ev.RootX, ev.RootY, c.(*client))
				break
			}
			if ev.State&xproto.ButtonMask1 != 0 {
				c.(*client).frame.mouseDrag(ev.RootX, ev.RootY)
			} else {
				c.(*client).frame.mouseMotion(ev.RootX, ev.RootY)
			}
			break
		}
	}
}

func (x *x11WM) handleMoveResize(ev xproto.ClientMessageEvent, c *client) {
	x.moveResizing = true
	x.moveResizingLastX = int16(ev.Data.Data32[0])
	x.moveResizingLastY = int16(ev.Data.Data32[1])
	x.moveResizingStartX = x.moveResizingLastX
	x.moveResizingStartY = x.moveResizingLastY
	_, _, x.moveResizingStartWidth, x.moveResizingStartHeight = c.getWindowGeometry()
	x.moveResizingType = moveResizeType(ev.Data.Data32[2])
	xproto.GrabPointer(x.x.Conn(), true, c.id,
		xproto.EventMaskButtonPress|xproto.EventMaskButtonRelease|xproto.EventMaskPointerMotion,
		xproto.GrabModeAsync, xproto.GrabModeAsync, x.x.RootWin(), xproto.CursorNone, xproto.TimeCurrentTime)
	xproto.GrabKeyboard(x.x.Conn(), true, c.id, xproto.TimeCurrentTime, xproto.GrabModeAsync, xproto.GrabModeAsync)
}

func (x *x11WM) handlePropertyChange(ev xproto.PropertyNotifyEvent) {
	c := x.clientForWin(ev.Window)
	if c == nil {
		return
	}
	propAtom, err := xprop.AtomName(x.x, ev.Atom)
	if err != nil {
		fyne.LogError("Error getting event", err)
		return
	}
	switch propAtom {
	case "_NET_WM_NAME":
		c.(*client).updateTitle()
	case "WM_NAME":
		c.(*client).updateTitle()
	case "WM_NORMAL_HINTS":
		// Force a reconfigure to make sure the client is constrained to the new size hints
		x, y, width, height := c.(*client).getWindowGeometry()
		c.(*client).setWindowGeometry(x, y, width, height)
	case "_MOTIF_WM_HINTS":
		c.(*client).setupBorder()
	}
}

func (x *x11WM) handleScreenChange(timestamp xproto.Timestamp) {
	if x.screenChangeTimestamp == timestamp {
		return
	}
	x.screenChangeTimestamp = timestamp
	desk := desktop.Instance()
	if desk == nil {
		return
	}
	desk.Screens().RefreshScreens()
	x.configureRoots(x.x.RootWin())
}

func (x *x11WM) handleStateActionRequest(ev xproto.ClientMessageEvent, removeState func(), addState func(), toggleCheck bool) {
	switch clientMessageStateAction(ev.Data.Data32[0]) {
	case clientMessageStateActionRemove:
		removeState()
	case clientMessageStateActionAdd:
		addState()
	case clientMessageStateActionToggle:
		if toggleCheck {
			removeState()
		} else {
			addState()
		}
	}
}

func (x *x11WM) handleVisibilityChange(ev xproto.VisibilityNotifyEvent) {
	attrs, err := xproto.GetWindowAttributes(x.x.Conn(), ev.Window).Reply()
	if err == nil && attrs.Colormap != 0 {
		if ev.State != xproto.VisibilityFullyObscured {
			xproto.InstallColormap(x.x.Conn(), attrs.Colormap)
		} else {
			xproto.UninstallColormap(x.x.Conn(), attrs.Colormap)
		}
	}
	colormaps, err := icccm.WmColormapWindowsGet(x.x, ev.Window)
	if err == nil {
		for _, child := range colormaps {
			chAttrs, err := xproto.GetWindowAttributes(x.x.Conn(), child).Reply()
			if err == nil && chAttrs.Colormap != 0 {
				if ev.State != xproto.VisibilityFullyObscured {
					xproto.InstallColormap(x.x.Conn(), chAttrs.Colormap)
				} else {
					xproto.UninstallColormap(x.x.Conn(), chAttrs.Colormap)
				}
			}
		}
	}
}

func (x *x11WM) moveResizeEnd(c *client) {
	x.moveResizing = false
	xproto.UngrabPointer(x.x.Conn(), xproto.TimeCurrentTime)
	xproto.UngrabKeyboard(x.x.Conn(), xproto.TimeCurrentTime)

	// ensure menus etc update
	c.frame.notifyInnerGeometry()
}

func (x *x11WM) moveResize(moveX, moveY int16, c *client) {
	xcoord, ycoord, width, height := c.getWindowGeometry()
	w := int16(width)
	h := int16(height)
	deltaW := moveX - x.moveResizingLastX
	deltaH := moveY - x.moveResizingLastY
	deltaX := moveX - x.moveResizingStartX
	deltaY := moveY - x.moveResizingStartY

	switch x.moveResizingType {
	case moveResizeTopLeft:
		//Move both X,Y coords and resize both W,H
		xcoord += deltaW
		ycoord += deltaH

		w = int16(x.moveResizingStartWidth) - deltaX
		h = int16(x.moveResizingStartHeight) - deltaY
	case moveResizeTop:
		//Move Y coord and resize H
		ycoord += deltaH
		h = int16(x.moveResizingStartHeight) - deltaY
	case moveResizeTopRight:
		//Move Y coord and resize both W,H
		ycoord += deltaH
		w = int16(x.moveResizingStartWidth) + deltaX
		h = int16(x.moveResizingStartHeight) - deltaY
	case moveResizeRight:
		//Keep X coord and resize W
		w = int16(x.moveResizingStartWidth) + deltaX
	case moveResizeBottomRight, moveResizeKeyboard:
		//Keep both X,Y coords and resize both W,H
		w = int16(x.moveResizingStartWidth) + deltaX
		h = int16(x.moveResizingStartHeight) + deltaY
	case moveResizeBottom:
		//Keep Y coord and resize H
		h = int16(x.moveResizingStartHeight) + deltaY
	case moveResizeBottomLeft:
		//Move X coord and resize both W,H
		xcoord += deltaW
		w = int16(x.moveResizingStartWidth) - deltaX
		h = int16(x.moveResizingStartHeight) + deltaY
	case moveResizeLeft:
		//Move X coord and resize W
		xcoord += deltaW
		w = int16(x.moveResizingStartWidth) - deltaX
	case moveResizeMove, moveResizeMoveKeyboard:
		//Move both X,Y coords and no resize
		xcoord += deltaW
		ycoord += deltaH
	case moveResizeCancel:
		x.moveResizeEnd(c)
	}
	x.moveResizingLastX = moveX
	x.moveResizingLastY = moveY
	c.setWindowGeometry(xcoord, ycoord, uint16(w), uint16(h))
}
